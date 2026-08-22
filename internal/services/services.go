package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/leomfn/rueckenwind/internal/models"
)

const (
	weatherRequestTimeout = 10 * time.Second
	// Overpass queries are heavy and the public instances are often busy, so
	// they get a more generous budget than the weather API.
	overpassRequestTimeout = 30 * time.Second
)

// Weather
type WeatherService interface {
	GetWeatherForecast(ctx context.Context, lon float64, lat float64) (models.WeatherSummary, error)
}

type openWeatherService struct {
	client           *http.Client
	forecastUrl      string
	apiKey           string
	maxForecastCount int64
}

func NewOpenWeatherService(apiKey string) WeatherService {
	return &openWeatherService{
		client:           &http.Client{Timeout: weatherRequestTimeout},
		forecastUrl:      "https://api.openweathermap.org/data/2.5/forecast",
		apiKey:           apiKey,
		maxForecastCount: 2,
	}
}

// Request weather forecast for next 12 hours in 3-hour blocks (4 items in total)
func (s *openWeatherService) GetWeatherForecast(ctx context.Context, lon float64, lat float64) (models.WeatherSummary, error) {
	query := fmt.Sprintf("?lat=%f&lon=%f&appid=%s&units=metric&cnt=%d",
		lat,
		lon,
		s.apiKey,
		s.maxForecastCount,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.forecastUrl+query, nil)
	if err != nil {
		log.Println("Could not build openweather request:", err)
		return models.WeatherSummary{}, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		log.Println("Error when fetching weather from openweather:", err)
		return models.WeatherSummary{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Openweather returned non-OK status: %d", resp.StatusCode)
		return models.WeatherSummary{}, fmt.Errorf("openweather returned status %d", resp.StatusCode)
	}

	var weatherForecast models.WeatherForecast

	if err := json.NewDecoder(resp.Body).Decode(&weatherForecast); err != nil {
		log.Println("Error when unmarshalling openweathermap response:", err)
		return models.WeatherSummary{}, err
	}

	// The summary pairs the current forecast block with the following one, so a
	// shorter list cannot be summarized and must not be indexed into.
	if len(weatherForecast.List) < 2 {
		log.Printf("Openweather returned %d forecast entries, expected at least 2", len(weatherForecast.List))
		return models.WeatherSummary{}, fmt.Errorf("openweather returned %d forecast entries", len(weatherForecast.List))
	}

	currentWeather := weatherForecast.List[0]
	nextWeather := weatherForecast.List[1]

	weatherSummary := models.WeatherSummary{
		CurrentTemperature: int64(math.Round(currentWeather.Main.Temp)),
		FutureTemperature:  int64(math.Round(nextWeather.Main.Temp)),
		CurrentWindSpeed:   int64(math.Round(currentWeather.Wind.Speed * 3.6)),
		FutureWindSpeed:    int64(math.Round(nextWeather.Wind.Speed * 3.6)),
		CurrentWindGust:    int64(math.Round(currentWeather.Wind.Gust * 3.6)),
		FutureWindGust:     int64(math.Round(nextWeather.Wind.Gust * 3.6)),
		CurrentWindDegrees: currentWeather.Wind.Deg,
		FutureWindDegrees:  nextWeather.Wind.Deg,
		CurrentWindScale:   currentWeather.Wind.Scale(),
		FutureWindScale:    nextWeather.Wind.Scale(),
		CurrentRain:        currentWeather.Rain.RainIntensity(),
		FutureRain:         nextWeather.Rain.RainIntensity(),
		CurrentRainText:    currentWeather.Rain.RainText(),
		FutureRainText:     nextWeather.Rain.RainText(),
		SunsetTime:         weatherForecast.SunsetLocalTime(),
	}

	return weatherSummary, nil
}

// Overpass
type overpassElement struct {
	OverpassType string  `json:"type"`
	Lon          float64 `json:"lon"`
	Lat          float64 `json:"lat"`
	Bounds       struct {
		MinLat float64 `json:"minlat"`
		MinLon float64 `json:"minLon"`
		MaxLat float64 `json:"maxLat"`
		MaxLon float64 `json:"maxLon"`
	} `json:"bounds"`
	Tags struct {
		Name        string `json:"name"`
		Website     string `json:"website"`
		Street      string `json:"addr:street"`
		Housenumber string `json:"addr:housenumber"`
		Postcode    string `json:"addr:postcode"`
		City        string `json:"addr:city"`
	} `json:"tags"`
}

func (e *overpassElement) GetAddress() string {
	if e.Tags.City == "" {
		return ""
	}

	var addressParts []string

	if e.Tags.Street != "" {
		street := e.Tags.Street
		if e.Tags.Housenumber != "" {
			street = strings.Join([]string{street, e.Tags.Housenumber}, " ")
		}

		street += ","
		addressParts = append(addressParts, street)
	}

	if e.Tags.Postcode != "" {
		addressParts = append(addressParts, e.Tags.Postcode)
	}

	addressParts = append(addressParts, e.Tags.City)

	return strings.Join(addressParts, " ")
}

type overpassResult struct {
	Elements []overpassElement `json:"elements"`
}

type PoiService interface {
	GetCampingPois(ctx context.Context, lon float64, lat float64) (models.OverpassSites, error)
	GetDrinkingWaterPois(ctx context.Context, lon float64, lat float64) (models.OverpassSites, error)
	GetCafePois(ctx context.Context, lon float64, lat float64) (models.OverpassSites, error)
	GetObservationPois(ctx context.Context, lon float64, lat float64) (models.OverpassSites, error)
}

type overpassPoiService struct {
	client      *http.Client
	url         string
	maxDistance int64
	userAgent   string
}

func NewOverpassPoiService(maxDistance int64, userAgent string) PoiService {
	return &overpassPoiService{
		client:      &http.Client{Timeout: overpassRequestTimeout},
		url:         "https://overpass-api.de/api/interpreter",
		maxDistance: maxDistance,
		userAgent:   userAgent,
	}
}

func (s *overpassPoiService) query(ctx context.Context, query string) (*overpassResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewBuffer([]byte(query)))
	if err != nil {
		log.Println("Could not build Overpass request")
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")
	// Overpass usage rules require a custom, identifying user agent for scripts;
	// stock or faked UAs get blocked. Configurable via OVERPASS_USER_AGENT so
	// self-hosters can set their own. See https://overpass-api.de/.
	req.Header.Set("User-Agent", s.userAgent)

	resp, err := s.client.Do(req)

	if err != nil {
		log.Println("Could not fetch POIs")
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Overpass returned non-OK status: %d", resp.StatusCode)
		return nil, fmt.Errorf("overpass returned status %d", resp.StatusCode)
	}

	var overpassResult = overpassResult{}
	if err := json.NewDecoder(resp.Body).Decode(&overpassResult); err != nil {
		log.Println("Error unmarshalling overpass result:", err)
		return nil, err
	}

	return &overpassResult, nil
}

func (s *overpassPoiService) convertOverpassResults(pois *overpassResult, lon float64, lat float64) models.OverpassSites {
	sites := models.OverpassSites{}

	for _, element := range pois.Elements {
		var siteLon, siteLat models.Coordinate
		switch element.OverpassType {
		case "node":
			siteLon = models.Coordinate(element.Lon)
			siteLat = models.Coordinate(element.Lat)
		case "way", "relation":
			siteLon = models.Coordinate((element.Bounds.MinLon + element.Bounds.MaxLon) / 2)
			siteLat = models.Coordinate((element.Bounds.MinLat + element.Bounds.MaxLat) / 2)
		}

		site := models.NewSite(models.Location{Lon: siteLon, Lat: siteLat}, models.Location{Lon: models.Coordinate(lon), Lat: models.Coordinate(lat)}, s.maxDistance)

		// TODO: add properties to filtered POIs instead of all overpass results
		site.Name = element.Tags.Name
		site.Website = element.Tags.Website
		site.Address = element.GetAddress()

		sites = append(sites, site)
	}

	return sites
}

func (s *overpassPoiService) GetCampingPois(ctx context.Context, lon float64, lat float64) (models.OverpassSites, error) {
	query := fmt.Sprintf(`[out:json];nwr["tourism"="camp_site"]["tent"!="no"](around:%d,%v,%v);out geom;`,
		s.maxDistance*1000,
		lat,
		lon)

	foundPois, err := s.query(ctx, query)

	if err != nil {
		log.Println("Could not fetch campsites")
		return nil, err
	}

	pois := s.convertOverpassResults(foundPois, lon, lat)
	pois.SortByDistance()
	pois.FilterByBearing()

	return pois, nil
}

func (s *overpassPoiService) GetDrinkingWaterPois(ctx context.Context, lon float64, lat float64) (models.OverpassSites, error) {
	query := fmt.Sprintf(`[out:json];(nwr["amenity"="drinking_water"]["access"!="permissive"]["access"!="private"](around:%d,%v,%v);nwr["drinking_water"="yes"]["access"!="permissive"]["access"!="private"](around:%d,%v,%v);nwr["disused:amenity"="drinking_water"]["access"!="permissive"]["access"!="private"](around:%d,%v,%v););out geom;`,
		s.maxDistance*1000, lat, lon,
		s.maxDistance*1000, lat, lon,
		s.maxDistance*1000, lat, lon)

	foundPois, err := s.query(ctx, query)

	if err != nil {
		log.Println("Could not fetch drinking water")
		return nil, err
	}

	pois := s.convertOverpassResults(foundPois, lon, lat)
	pois.SortByDistance()
	pois.FilterByBearing()

	return pois, nil
}

func (s *overpassPoiService) GetCafePois(ctx context.Context, lon float64, lat float64) (models.OverpassSites, error) {
	query := fmt.Sprintf(`[out:json];nwr["amenity"="cafe"](around:%d,%v,%v);out geom;`,
		s.maxDistance*1000,
		lat,
		lon)

	foundPois, err := s.query(ctx, query)

	if err != nil {
		log.Println("Could not fetch cafes")
		return nil, err
	}

	pois := s.convertOverpassResults(foundPois, lon, lat)
	pois.SortByDistance()
	pois.FilterByBearing()

	return pois, nil
}

func (s *overpassPoiService) GetObservationPois(ctx context.Context, lon float64, lat float64) (models.OverpassSites, error) {
	query := fmt.Sprintf(`[out:json];(nwr["man_made"="tower"]["tower:type"="observation"](around:%d,%v,%v);nwr["leisure"="bird_hide"](around:%d,%v,%v););out geom;`,
		s.maxDistance*1000,
		lat,
		lon,
		s.maxDistance*1000,
		lat,
		lon)

	foundPois, err := s.query(ctx, query)

	if err != nil {
		log.Println("Could not fetch observation sites")
		return nil, err
	}

	pois := s.convertOverpassResults(foundPois, lon, lat)
	pois.SortByDistance()
	pois.FilterByBearing()

	return pois, nil
}
