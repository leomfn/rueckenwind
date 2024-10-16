package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"

	"github.com/leomfn/rueckenwind/internal/models"
)

// Weather
type WeatherService interface {
	GetWeatherForecast(lon float64, lat float64) (models.WeatherSummary, error)
}

type openWeatherService struct {
	forecastUrl      string
	apiKey           string
	maxForecastCount int64
	forecastInterval int64 // in hours
}

func NewOpenWeatherService(apiKey string) WeatherService {
	return &openWeatherService{
		forecastUrl:      "https://api.openweathermap.org/data/2.5/forecast",
		apiKey:           apiKey,
		maxForecastCount: 2,
		forecastInterval: 3,
	}
}

// Request weather forecast for next 12 hours in 3-hour blocks (4 items in total)
func (s *openWeatherService) GetWeatherForecast(lon float64, lat float64) (models.WeatherSummary, error) {
	query := fmt.Sprintf("?lat=%f&lon=%f&appid=%s&units=metric&cnt=%d",
		lat,
		lon,
		s.apiKey,
		s.maxForecastCount,
	)

	resp, err := http.Get(s.forecastUrl + query)
	if err != nil {
		log.Println("Error when fetching weather from openweather:", err)
		return models.WeatherSummary{}, err
	}
	defer resp.Body.Close()

	var weatherForecast models.WeatherForecast

	if err := json.NewDecoder(resp.Body).Decode(&weatherForecast); err != nil {
		log.Println("Error when unmarshalling openweathermap response:", err)
		return models.WeatherSummary{}, err
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
type overpassResult struct {
	Elements []struct {
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
			Name    string `json:"name"`
			Website string `json:"website"`
		} `json:"tags"`
	} `json:"elements"`
}

type PoiService interface {
	GetCampingPois(lon float64, lat float64) (models.OverpassSites, error)
	GetDrinkingWaterPois(lon float64, lat float64) (models.OverpassSites, error)
	GetCafePois(lon float64, lat float64) (models.OverpassSites, error)
	GetObservationPois(lon float64, lat float64) (models.OverpassSites, error)
}

type overpassPoiService struct {
	url         string
	maxDistance int64
}

func NewOverpassPoiService(maxDistance int64) PoiService {
	return &overpassPoiService{
		url:         "https://overpass-api.de/api/interpreter",
		maxDistance: maxDistance,
	}
}

func (s *overpassPoiService) query(query string) (*overpassResult, error) {
	resp, err := http.Post(s.url, "text/plain", bytes.NewBuffer([]byte(query)))

	if err != nil {
		log.Println("Could not fetch POIs")
		return nil, err
	}

	defer resp.Body.Close()

	var overpassResult = overpassResult{}
	if err := json.NewDecoder(resp.Body).Decode(&overpassResult); err != nil {
		log.Println("Error unmarshalling overpass result:", err)
		return nil, err
	}

	return &overpassResult, nil
}

func (s *overpassPoiService) convertOverpassResults(pois *overpassResult, lon float64, lat float64) models.OverpassSites {
	campsites := models.OverpassSites{}

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

		campsite := models.NewSite(models.Location{Lon: siteLon, Lat: siteLat}, models.Location{Lon: models.Coordinate(lon), Lat: models.Coordinate(lat)}, s.maxDistance)
		campsite.Name = element.Tags.Name
		campsite.Website = element.Tags.Website

		campsites = append(campsites, campsite)
	}

	return campsites
}

func (s *overpassPoiService) GetCampingPois(lon float64, lat float64) (models.OverpassSites, error) {
	query := fmt.Sprintf(`[out:json];nwr["tourism"="camp_site"]["tent"!="no"](around:%d,%v,%v);out geom;`,
		s.maxDistance*1000,
		lat,
		lon)

	foundPois, err := s.query(query)

	if err != nil {
		log.Println("Could not fetch campsites")
		return nil, err
	}

	pois := s.convertOverpassResults(foundPois, lon, lat)
	pois.SortByDistance()
	pois.FilterByBearing()

	return pois, nil
}

func (s *overpassPoiService) GetDrinkingWaterPois(lon float64, lat float64) (models.OverpassSites, error) {
	query := fmt.Sprintf(`[out:json];(nwr["amenity"="drinking_water"]["access"!="permissive"]["access"!="private"](around:%d,%v,%v);nwr["drinking_water"="yes"]["access"!="permissive"]["access"!="private"](around:%d,%v,%v);nwr["disused:amenity"="drinking_water"]["access"!="permissive"]["access"!="private"](around:%d,%v,%v););out geom;`,
		s.maxDistance*1000, lat, lon,
		s.maxDistance*1000, lat, lon,
		s.maxDistance*1000, lat, lon)

	foundPois, err := s.query(query)

	if err != nil {
		log.Println("Could not fetch drinking water")
		return nil, err
	}

	pois := s.convertOverpassResults(foundPois, lon, lat)
	pois.SortByDistance()
	pois.FilterByBearing()

	return pois, nil
}

func (s *overpassPoiService) GetCafePois(lon float64, lat float64) (models.OverpassSites, error) {
	query := fmt.Sprintf(`[out:json];nwr["amenity"="cafe"](around:%d,%v,%v);out geom;`,
		s.maxDistance*1000,
		lat,
		lon)

	foundPois, err := s.query(query)

	if err != nil {
		log.Println("Could not fetch cafes")
		return nil, err
	}

	pois := s.convertOverpassResults(foundPois, lon, lat)
	pois.SortByDistance()
	pois.FilterByBearing()

	return pois, nil
}

func (s *overpassPoiService) GetObservationPois(lon float64, lat float64) (models.OverpassSites, error) {
	query := fmt.Sprintf(`[out:json];(nwr["man_made"="tower"]["tower:type"="observation"](around:%d,%v,%v);nwr["leisure"="bird_hide"](around:%d,%v,%v););out geom;`,
		s.maxDistance*1000,
		lat,
		lon,
		s.maxDistance*1000,
		lat,
		lon)

	foundPois, err := s.query(query)

	if err != nil {
		log.Println("Could not fetch observation sites")
		return nil, err
	}

	pois := s.convertOverpassResults(foundPois, lon, lat)
	pois.SortByDistance()
	pois.FilterByBearing()

	return pois, nil
}
