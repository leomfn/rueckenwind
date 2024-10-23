package models

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type Coordinate float64

func (c *Coordinate) toRadians() float64 {
	return float64(*c) / 180 * math.Pi
}

type Location struct {
	Lon Coordinate `json:"lon"`
	Lat Coordinate `json:"lat"`
}

// Haversine distance
func (l1 Location) distance(l2 Location) float64 {
	earthRadius := 6371.0 // km

	lat1 := l1.Lat.toRadians()
	lon1 := l1.Lon.toRadians()
	lat2 := l2.Lat.toRadians()
	lon2 := l2.Lon.toRadians()

	return 2 * earthRadius * math.Asin(
		math.Sqrt(
			(1-math.Cos(lat2-lat1)+math.Cos(lat1)*math.Cos(lat2)*(1-math.Cos(lon2-lon1)))/2,
		),
	)
}

// Bearing angle between two points, where l1 is the reference point and the
// bearing expresses the angle between north and the line through l2
func (l1 Location) bearing(l2 Location) float64 {
	lat1 := l1.Lat.toRadians()
	lon1 := l1.Lon.toRadians()
	lat2 := l2.Lat.toRadians()
	lon2 := l2.Lon.toRadians()

	return math.Atan2(
		math.Sin(lon2-lon1)*math.Cos(lat2),
		math.Cos(lat1)*math.Sin(lat2)-math.Sin(lat1)*math.Cos(lat2)*math.Cos(lon2-lon1),
	) / math.Pi * 180
}

type Weather struct {
	Id          int64  `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type Main struct {
	Temp        float64 `json:"temp"`
	FeelsLike   float64 `json:"feels_like"`
	TempMin     float64 `json:"temp_min"`
	TempMax     float64 `json:"temp_max"`
	Pressure    int64   `json:"pressure"`
	SeaLevel    int64   `json:"sea_level"`
	GroundLevel int64   `json:"grnd_level"`
	Humidity    int64   `json:"humidity"`
	TempKf      float64 `json:"temp_kf"`
}

type Wind struct {
	Speed float64 `json:"speed"`
	Deg   int64   `json:"deg"`
	Gust  float64 `json:"gust"`
}

func (w Wind) Scale() float64 {
	maxSpeed := 80.0
	speed := w.Speed * 3.6

	if w.Speed > maxSpeed {
		speed = maxSpeed
	}

	scale := 0.2 + 0.8*math.Tanh(0.03*speed)

	return scale
}

type Rain struct {
	ThreeHours float64 `json:"3h"`
}

// Returns rain intensity, as an integer from 0 (no rain) to 3 (heavy rain).
func (r Rain) RainIntensity() int64 {
	hourlyRain := r.ThreeHours / 3

	switch {
	case hourlyRain == 0:
		return 0
	case hourlyRain < 0.1:
		return 1
	case hourlyRain < 0.5:
		return 2
	default:
		return 3
	}
}

func (r Rain) RainText() string {

	switch r.RainIntensity() {
	case 0:
		return "dry"
	case 1:
		return "light"
	case 2:
		return "medium"
	case 3:
		return "heavy"
	}

	return ""
}

type Clouds struct {
	All int64 `json:"all"`
}

type Sys struct {
	Pod string `json:"pod"`
}

type City struct {
	Id         int64    `json:"id"`
	Name       string   `json:"name"`
	Coord      Location `json:"coord"`
	Country    string   `json:"country"`
	Population int64    `json:"population"`
	Timezone   int64    `json:"timezone"`
	Sunrise    int64    `json:"sunrise"`
	Sunset     int64    `json:"sunset"`
}

type ForecastEntry struct {
	Timestamp     int64     `json:"dt"`
	Main          Main      `json:"main"`
	Weather       []Weather `json:"weather"`
	Clouds        Clouds    `json:"clouds"`
	Wind          Wind      `json:"wind"`
	Visibility    int64     `json:"visibility"`
	Pop           float64   `json:"pop"`
	Rain          Rain      `json:"rain"`
	Sys           Sys       `json:"sys"`
	TimestampText string    `json:"dt_txt"`
}

type WeatherForecast struct {
	List    []ForecastEntry `json:"list"`
	COD     string          `json:"cod"`
	Message int64           `json:"message"`
	Count   int64           `json:"cnt"`
	City    City            `json:"city"`
}

func (w WeatherForecast) SunsetLocalTime() string {
	utcTime := time.Unix(int64(w.City.Sunset), 0).UTC()
	return utcTime.Add(time.Duration(w.City.Timezone) * time.Second).Format("15:04")

}

type WeatherSummary struct {
	// Temperature in Degree Celsius
	CurrentTemperature int64 `json:"temp_current"`
	FutureTemperature  int64 `json:"temp_future"`

	// Wind speed in km/h
	CurrentWindSpeed   int64   `json:"wind_current"`
	FutureWindSpeed    int64   `json:"wind_future"`
	CurrentWindDegrees int64   `json:"wind_deg_current"`
	FutureWindDegrees  int64   `json:"wind_deg_future"`
	CurrentWindGust    int64   `json:"wind_gust_current"`
	FutureWindGust     int64   `json:"wind_gust_future"`
	CurrentWindScale   float64 `json:"wind_scale_current"`
	FutureWindScale    float64 `json:"wind_scale_future"`
	CurrentRain        int64   `json:"rain_current"`
	FutureRain         int64   `json:"rain_future"`
	CurrentRainText    string  `json:"rain_current_text"`
	FutureRainText     string  `json:"rain_future_text"`

	// Local time
	SunsetTime string `json:"sunset"`
}

type poi struct {
	location Location
	distance float64
	bearing  float64
}

type campingSite struct {
	poi
	name         string
	address      string
	website      string
	openingHours string
}

type drinkingWaterSite struct {
	poi
}

type cafeSite struct {
	poi
	name         string
	address      string
	website      string
	openingHours string
}

type overpassSite struct {
	Bearing       float64 `json:"bearing"`
	Distance      float64 `json:"distance"`
	DistanceText  string  `json:"distance_text"`
	DistancePixel float64 `json:"distance_pixel"`
	Name          string  `json:"name"`
	Website       string  `json:"website"`
	Lon           float64 `json:"lon"`
	Lat           float64 `json:"lat"`
	Address       string  `json:"address"`
}

type Pois struct {
	pois []poi
}

type campingSites struct {
	poisInterface
}

type poisInterface interface {
	sortByDistance()
	filterByBearing()
}

func newCampingSites() poisInterface {
	return nil
}

type OverpassSites []overpassSite

func (p *OverpassSites) SortByDistance() {
	sort.Slice(*p, func(i, j int) bool {
		return (*p)[i].Distance < (*p)[j].Distance
	})
}

func (p *Pois) sortByDistance() {
	sort.Slice(p.pois, func(i, j int) bool {
		return p.pois[i].distance < p.pois[j].distance
	})
}

// POIs are filtered by bearing angle, so that only one POI remains per bearing
// 'bucket'. 360 degrees are split into 12 buckets of 30 degrees each.
//
// TODO: Find a better solution for overlapping POIs at bucket boundaries.
// TODO: Maybe find a solution for close-to-nearest POIs not being shown because
// bucket is already full.
func (p *Pois) filterByBearing() {
	angleFractions := map[int]bool{
		0:  false,
		1:  false,
		2:  false,
		3:  false,
		4:  false,
		5:  false,
		6:  false,
		7:  false,
		8:  false,
		9:  false,
		10: false,
		11: false,
	}

	var filteredPois Pois
	for _, poi := range p.pois {
		angleFraction := int((poi.bearing + 180) / 30)
		if !angleFractions[angleFraction] {
			filteredPois.pois = append(filteredPois.pois, poi)
			angleFractions[angleFraction] = true
		}
	}

	*p = filteredPois
}

func NewSite(siteLocation Location, referenceLocation Location, maxDistance int64) overpassSite {
	distance := referenceLocation.distance(siteLocation)
	var distanceText string

	// Show first decimal place for distances under 2km
	if distance < 2 {
		distanceText = fmt.Sprintf("%.1f", distance)
	} else {
		distanceText = fmt.Sprintf("%.0f", distance)
	}
	maxPixel := 50.0
	minPixel := 20.0
	distancePixel := minPixel + (maxPixel-minPixel)*distance/float64(maxDistance)

	return overpassSite{
		Bearing:       referenceLocation.bearing(siteLocation),
		Distance:      distance,
		DistanceText:  distanceText,
		DistancePixel: distancePixel,
		Lon:           float64(siteLocation.Lon),
		Lat:           float64(siteLocation.Lat),
	}
}

func (sites *OverpassSites) FilterByBearing() {
	angleFractions := map[int]bool{
		0:  false,
		1:  false,
		2:  false,
		3:  false,
		4:  false,
		5:  false,
		6:  false,
		7:  false,
		8:  false,
		9:  false,
		10: false,
		11: false,
	}

	var filteredSites OverpassSites

	for _, site := range *sites {
		angleFraction := int((site.Bearing + 180) / 30)
		if !angleFractions[angleFraction] {
			filteredSites = append(filteredSites, site)
			angleFractions[angleFraction] = true
		}
	}

	*sites = filteredSites
}
