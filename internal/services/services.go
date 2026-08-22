package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/leomfn/rueckenwind/internal/cache"
	"github.com/leomfn/rueckenwind/internal/models"
)

const (
	weatherRequestTimeout = 10 * time.Second
	// Overpass queries are heavy and the public instances are often busy, so
	// they get a more generous budget than the weather API.
	overpassRequestTimeout = 30 * time.Second
)

// How long upstream results are reused. OpenStreetMap data changes slowly, the
// weather forecast is published in three hour blocks.
const (
	poiCacheTtl     = time.Hour
	weatherCacheTtl = 10 * time.Minute
)

// Before a location is used as a cache key it is snapped to a grid, so that
// nearby users share one upstream result instead of each triggering their own.
// One hundredth of a degree is roughly 1.1 km, one fiftieth roughly 2.2 km.
const (
	poiCacheGrid     = 0.01
	weatherCacheGrid = 0.02
)

// Rounds a coordinate to the nearest multiple of grid.
func snapToGrid(value float64, grid float64) float64 {
	return math.Round(value/grid) * grid
}

// Weather
type WeatherService interface {
	GetWeatherForecast(ctx context.Context, lon float64, lat float64) (models.WeatherSummary, error)
}

type openWeatherService struct {
	client           *http.Client
	cache            *cache.Cache[models.WeatherSummary]
	forecastUrl      string
	apiKey           string
	maxForecastCount int64
}

func NewOpenWeatherService(apiKey string) WeatherService {
	return &openWeatherService{
		client:           &http.Client{Timeout: weatherRequestTimeout},
		cache:            cache.New[models.WeatherSummary](weatherCacheTtl),
		forecastUrl:      "https://api.openweathermap.org/data/2.5/forecast",
		apiKey:           apiKey,
		maxForecastCount: 2,
	}
}

// Request weather forecast for next 12 hours in 3-hour blocks (4 items in
// total). Results are cached per grid cell, so that a moving user or a group of
// nearby users does not consume one API call per request.
func (s *openWeatherService) GetWeatherForecast(ctx context.Context, lon float64, lat float64) (models.WeatherSummary, error) {
	// The snapped location is used for both the cache key and the request, so
	// that the cached summary always matches the location it was fetched for.
	queryLat := snapToGrid(lat, weatherCacheGrid)
	queryLon := snapToGrid(lon, weatherCacheGrid)

	key := fmt.Sprintf("%.2f/%.2f", queryLat, queryLon)

	return s.cache.Get(ctx, key, func(ctx context.Context) (models.WeatherSummary, error) {
		return s.fetchWeatherForecast(ctx, queryLon, queryLat)
	})
}

func (s *openWeatherService) fetchWeatherForecast(ctx context.Context, lon float64, lat float64) (models.WeatherSummary, error) {
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

		if isTransientStatus(resp.StatusCode) {
			return models.WeatherSummary{}, fmt.Errorf("%w: openweather returned status %d", ErrUpstreamBusy, resp.StatusCode)
		}

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

// ErrUnknownCategory is returned for POI categories that have no Overpass query.
var ErrUnknownCategory = errors.New("unknown poi category")

// ErrUpstreamBusy is returned when an upstream service is temporarily unable to
// answer, rather than the request being wrong. Overpass reports this as 429
// when the per-address slots are used up, and as 502/503/504 when its own
// backend is overloaded. It is worth distinguishing, because the caller only
// has to try again.
var ErrUpstreamBusy = errors.New("upstream service is busy")

// Reports whether a status means "try again later" rather than "this request
// was wrong".
func isTransientStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// Number of decimals used when writing coordinates into an Overpass query,
// which is roughly centimetre precision.
const coordinateDecimals = 7

// Overpass query per POI category. Each template is formatted with the search
// radius in meters, the latitude and the longitude, in that order; clauses that
// need the location more than once refer back to the same arguments by index.
var overpassQueries = map[string]string{
	"camping": `[out:json];nwr["tourism"="camp_site"]["tent"!="no"](around:%[1]d,%[2]s,%[3]s);out geom;`,

	"water": `[out:json];(nwr["amenity"="drinking_water"]["access"!="permissive"]["access"!="private"](around:%[1]d,%[2]s,%[3]s);` +
		`nwr["drinking_water"="yes"]["access"!="permissive"]["access"!="private"](around:%[1]d,%[2]s,%[3]s);` +
		`nwr["disused:amenity"="drinking_water"]["access"!="permissive"]["access"!="private"](around:%[1]d,%[2]s,%[3]s););out geom;`,

	"cafe": `[out:json];nwr["amenity"="cafe"](around:%[1]d,%[2]s,%[3]s);out geom;`,

	"observation": `[out:json];(nwr["man_made"="tower"]["tower:type"="observation"](around:%[1]d,%[2]s,%[3]s);` +
		`nwr["leisure"="bird_hide"](around:%[1]d,%[2]s,%[3]s););out geom;`,
}

type PoiService interface {
	GetPois(ctx context.Context, category string, lon float64, lat float64) (models.OverpassSites, error)
}

// Overpass grants only a small number of query slots per address, and answers
// with 429 once they are used up. Requests beyond this limit wait for a slot
// instead of being sent and rejected. See https://overpass-api.de/.
const maxConcurrentOverpassRequests = 2

type overpassPoiService struct {
	client *http.Client
	cache  *cache.Cache[*overpassResult]
	// Buffered to the number of requests that may be in flight at once. Holding
	// a slot means holding one of the buffer's places.
	slots       chan struct{}
	url         string
	maxDistance int64
	userAgent   string
}

func NewOverpassPoiService(maxDistance int64, userAgent string) PoiService {
	return &overpassPoiService{
		client:      &http.Client{Timeout: overpassRequestTimeout},
		cache:       cache.New[*overpassResult](poiCacheTtl),
		slots:       make(chan struct{}, maxConcurrentOverpassRequests),
		url:         "https://overpass-api.de/api/interpreter",
		maxDistance: maxDistance,
		userAgent:   userAgent,
	}
}

func (s *overpassPoiService) query(ctx context.Context, query string) (*overpassResult, error) {
	// Wait for a slot rather than adding to the load Overpass is already
	// refusing. Giving up when the caller does keeps a queue from building up
	// behind requests nobody is waiting for any more.
	select {
	case s.slots <- struct{}{}:
		defer func() { <-s.slots }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

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

		if isTransientStatus(resp.StatusCode) {
			return nil, fmt.Errorf("%w: overpass returned status %d", ErrUpstreamBusy, resp.StatusCode)
		}

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

// GetPois looks up the POIs of the given category around the location, sorted
// by distance and thinned out to one POI per bearing bucket.
func (s *overpassPoiService) GetPois(ctx context.Context, category string, lon float64, lat float64) (models.OverpassSites, error) {
	queryTemplate, known := overpassQueries[category]
	if !known {
		return nil, fmt.Errorf("%w: %s", ErrUnknownCategory, category)
	}

	// The search is centred on the snapped location, so that nearby users share
	// a cache entry. The circle therefore sits up to about a kilometre off the
	// user, which can miss POIs sitting exactly at the edge of the radius.
	//
	// Format the coordinates with a fixed number of decimals. The default float
	// formatting switches to scientific notation for small values, which
	// Overpass does not accept.
	queryLat := strconv.FormatFloat(snapToGrid(lat, poiCacheGrid), 'f', coordinateDecimals, 64)
	queryLon := strconv.FormatFloat(snapToGrid(lon, poiCacheGrid), 'f', coordinateDecimals, 64)

	query := fmt.Sprintf(queryTemplate, s.maxDistance*1000, queryLat, queryLon)

	// Only the raw Overpass response is cached. Distances and bearings are
	// derived per request from the caller's exact location, so sharing a cache
	// entry never costs accuracy in what the user is shown.
	foundPois, err := s.cache.Get(ctx, category+"/"+queryLat+"/"+queryLon,
		func(ctx context.Context) (*overpassResult, error) {
			return s.query(ctx, query)
		})

	if err != nil {
		log.Printf("Could not fetch POIs of category %s", category)
		return nil, err
	}

	pois := s.convertOverpassResults(foundPois, lon, lat)
	pois.SortByDistance()
	pois.FilterByBearing()

	return pois, nil
}
