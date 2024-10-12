package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
)

// Weather

func getWeather(lat coordinate, lon coordinate) (WeatherSummary, error) {
	// Request weather forecast for next 12 hours in 3-hour blocks (4 items in total)
	owmUrl := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?lat=%f&lon=%f&appid=%s&units=metric&cnt=2", lat, lon, owmApiKey)

	resp, err := http.Get(owmUrl)
	if err != nil {
		log.Println("Error when fetching weather from openweather:", err)
		return WeatherSummary{}, err
	}
	defer resp.Body.Close()

	var weatherForecast WeatherForecast

	if err := json.NewDecoder(resp.Body).Decode(&weatherForecast); err != nil {
		log.Println("Error when unmarshalling openweathermap response:", err)
		return WeatherSummary{}, err
	}

	currentWeather := weatherForecast.List[0]
	nextWeather := weatherForecast.List[1]

	weatherSummary := WeatherSummary{
		CurrentTemperature: int(math.Round(currentWeather.Main.Temp)),
		FutureTemperature:  int(math.Round(nextWeather.Main.Temp)),
		CurrentWindSpeed:   int(math.Round(currentWeather.Wind.Speed * 3.6)),
		FutureWindSpeed:    int(math.Round(nextWeather.Wind.Speed * 3.6)),
		CurrentWindGust:    int(math.Round(currentWeather.Wind.Gust * 3.6)),
		FutureWindGust:     int(math.Round(nextWeather.Wind.Gust * 3.6)),
		CurrentWindDegrees: currentWeather.Wind.Deg,
		FutureWindDegrees:  nextWeather.Wind.Deg,
		CurrentWindScale:   currentWeather.Wind.scale(),
		FutureWindScale:    nextWeather.Wind.scale(),
		CurrentRain:        currentWeather.Rain.rainIntensity(),
		FutureRain:         nextWeather.Rain.rainIntensity(),
		CurrentRainText:    currentWeather.Rain.rainText(),
		FutureRainText:     nextWeather.Rain.rainText(),
		SunsetTime:         weatherForecast.SunsetLocalTime(),
	}

	return weatherSummary, nil

	// TODO: maybe useful for debugging/testing:

	// return WeatherSummary{
	// 	CurrentTemperature: 10,
	// 	FutureTemperature:  13,
	// 	CurrentWindSpeed:   8,
	// 	FutureWindSpeed:    23,
	// 	CurrentWindGust:    13,
	// 	FutureWindGust:     51,
	// 	CurrentWindDegrees: 13,
	// 	FutureWindDegrees:  113,
	// 	CurrentRain:        0,
	// 	FutureRain:         0,
	// 	CurrentRainText:    "dry",
	// 	FutureRainText:     "dry",
	// 	SunsetTime:         "19:13",
	// }, nil
}

// Overpass

type overpassQuery struct {
	url                string
	center             location
	maxDistance        int64
	campSites          overpassSites
	drinkingWaterSites overpassSites
	cafeSites          overpassSites
}

func newOverpassQuery(center location) *overpassQuery {
	return &overpassQuery{
		url:         "https://overpass-api.de/api/interpreter",
		center:      center,
		maxDistance: maxOverpassDistance,
	}
}

// Executes queries to Overpass API to get drinking water and campsites
func (o *overpassQuery) execute() error {
	var err error
	type chResult struct {
		siteType string
		result   overpassSites
		err      error
	}

	// TODO: maybe use goroutines to get results in parallel?
	ch := make(chan chResult)
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		wg.Wait()
		close(ch)
	}()

	// Fetch campingsites in goroutine
	go func() {
		defer wg.Done()
		campSites, err := o.getCampSites()
		campSites.sortByDistance()
		campSites.filterByBearing()
		ch <- chResult{
			siteType: "campsites",
			result:   campSites,
			err:      err,
		}
	}()

	// Fetch drinking water sites in goroutine
	go func() {
		defer wg.Done()
		drinkingWaterSites, err := o.getDrinkingWaterSites()
		drinkingWaterSites.sortByDistance()
		drinkingWaterSites.filterByBearing()
		ch <- chResult{
			siteType: "water",
			result:   drinkingWaterSites,
			err:      err,
		}
	}()

	// Fetch cafes in goroutine
	go func() {
		defer wg.Done()
		cafeSites, err := o.getCafeSites()
		cafeSites.sortByDistance()
		cafeSites.filterByBearing()
		ch <- chResult{
			siteType: "cafes",
			result:   cafeSites,
			err:      err,
		}
	}()

	for sites := range ch {
		if sites.err != nil {
			return err
		}

		switch sites.siteType {
		case "campsites":
			o.campSites = sites.result
		case "water":
			o.drinkingWaterSites = sites.result
		case "cafes":
			o.cafeSites = sites.result
		default:
			return errors.New("something went wrong during concurrently fetching overpass data")
		}
	}

	return nil
}

func (o *overpassQuery) getCampSites() (overpassSites, error) {

	query := fmt.Sprintf(`[out:json];nwr["tourism"="camp_site"]["tent"!="no"](around:%d,%v,%v);out geom;`, o.maxDistance*1000, *o.center.Lat, *o.center.Lon)

	resp, err := http.Post(o.url, "text/plain", bytes.NewBuffer([]byte(query)))

	if err != nil {
		log.Println("Could not fetch campsites")
		return nil, err
	}

	defer resp.Body.Close()

	var overpassResult = overpassResult{}
	if err := json.NewDecoder(resp.Body).Decode(&overpassResult); err != nil {
		log.Println("Error unmarshalling overpass result:", err)
		return nil, err
	}

	campsites := overpassSites{}

	for _, element := range overpassResult.Elements {
		var lon, lat coordinate
		switch element.OverpassType {
		case "node":
			lon = coordinate(element.Lon)
			lat = coordinate(element.Lat)
		case "way", "relation":
			lon = coordinate((element.Bounds.MinLon + element.Bounds.MaxLon) / 2)
			lat = coordinate((element.Bounds.MinLat + element.Bounds.MaxLat) / 2)
		}

		campsite := newSite(location{Lon: &lon, Lat: &lat}, o.center)
		campsite.Name = element.Tags.Name
		campsite.Website = element.Tags.Website

		campsites = append(campsites, campsite)
	}

	return campsites, nil
}

func (o *overpassQuery) getDrinkingWaterSites() (overpassSites, error) {
	// TODO: optimize query, which takes very long due to the many conditions
	query := fmt.Sprintf(`[out:json];(nwr["amenity"="drinking_water"]["access"!="permissive"]["access"!="private"](around:%d,%v,%v);nwr["drinking_water"="yes"]["access"!="permissive"]["access"!="private"](around:%d,%v,%v);nwr["disused:amenity"="drinking_water"]["access"!="permissive"]["access"!="private"](around:%d,%v,%v););out geom;`,
		o.maxDistance*1000, *o.center.Lat, *o.center.Lon,
		o.maxDistance*1000, *o.center.Lat, *o.center.Lon,
		o.maxDistance*1000, *o.center.Lat, *o.center.Lon)

	resp, err := http.Post(o.url, "text/plain", bytes.NewBuffer([]byte(query)))

	if err != nil {
		log.Println("Could not fetch drinking water sites")
		return nil, err
	}

	defer resp.Body.Close()

	var overpassResult = overpassResult{}
	if err := json.NewDecoder(resp.Body).Decode(&overpassResult); err != nil {
		log.Println("Error unmarshalling overpass result:", err)
		return nil, err
	}

	drinkingWaterSites := overpassSites{}

	for _, element := range overpassResult.Elements {
		var lon, lat coordinate
		switch element.OverpassType {
		case "node":
			lon = coordinate(element.Lon)
			lat = coordinate(element.Lat)
		case "way", "relation":
			lon = coordinate((element.Bounds.MinLon + element.Bounds.MaxLon) / 2)
			lat = coordinate((element.Bounds.MinLat + element.Bounds.MaxLat) / 2)
		}

		drinkingWaterSite := newSite(location{Lon: &lon, Lat: &lat}, o.center)

		drinkingWaterSites = append(drinkingWaterSites, drinkingWaterSite)
	}

	return drinkingWaterSites, nil
}

func (o *overpassQuery) getCafeSites() (overpassSites, error) {
	// TODO: optimize query, which takes very long due to the many conditions
	query := fmt.Sprintf(`[out:json];nwr["amenity"="cafe"](around:%d,%v,%v);out geom;`,
		o.maxDistance*1000, *o.center.Lat, *o.center.Lon)

	resp, err := http.Post(o.url, "text/plain", bytes.NewBuffer([]byte(query)))

	if err != nil {
		log.Println("Could not fetch cafe sites")
		return nil, err
	}

	defer resp.Body.Close()

	var overpassResult = overpassResult{}
	if err := json.NewDecoder(resp.Body).Decode(&overpassResult); err != nil {
		log.Println("Error unmarshalling overpass result:", err)
		return nil, err
	}

	cafeSites := overpassSites{}

	for _, element := range overpassResult.Elements {
		var lon, lat coordinate
		switch element.OverpassType {
		case "node":
			lon = coordinate(element.Lon)
			lat = coordinate(element.Lat)
		case "way", "relation":
			lon = coordinate((element.Bounds.MinLon + element.Bounds.MaxLon) / 2)
			lat = coordinate((element.Bounds.MinLat + element.Bounds.MaxLat) / 2)
		}

		cafeSite := newSite(location{Lon: &lon, Lat: &lat}, o.center)

		cafeSites = append(cafeSites, cafeSite)
	}

	return cafeSites, nil
}
