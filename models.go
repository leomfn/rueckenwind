package main

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type coordinate float64

func (c coordinate) toRadians() float64 {
	return float64(c) / 180 * math.Pi
}

type location struct {
	Lon *coordinate `json:"lon"`
	Lat *coordinate `json:"lat"`
}

// Haversine distance
func (l1 location) distance(l2 location) float64 {
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
func (l1 location) bearing(l2 location) float64 {
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
	Id          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type Main struct {
	Temp        float64 `json:"temp"`
	FeelsLike   float64 `json:"feels_like"`
	TempMin     float64 `json:"temp_min"`
	TempMax     float64 `json:"temp_max"`
	Pressure    int     `json:"pressure"`
	SeaLevel    int     `json:"sea_level"`
	GroundLevel int     `json:"grnd_level"`
	Humidity    int     `json:"humidity"`
	TempKf      float64 `json:"temp_kf"`
}

type Wind struct {
	Speed float64 `json:"speed"`
	Deg   int     `json:"deg"`
	Gust  float64 `json:"gust"`
}

func (w Wind) scale() float64 {
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
func (r Rain) rainIntensity() int {
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

func (r Rain) rainText() string {

	switch r.rainIntensity() {
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
	All int `json:"all"`
}

type Sys struct {
	Pod string `json:"pod"`
}

type City struct {
	Id         int64    `json:"id"`
	Name       string   `json:"name"`
	Coord      location `json:"coord"`
	Country    string   `json:"country"`
	Population int      `json:"population"`
	Timezone   int      `json:"timezone"`
	Sunrise    int      `json:"sunrise"`
	Sunset     int      `json:"sunset"`
}

type ForecastEntry struct {
	Timestamp     int       `json:"dt"`
	Main          Main      `json:"main"`
	Weather       []Weather `json:"weather"`
	Clouds        Clouds    `json:"clouds"`
	Wind          Wind      `json:"wind"`
	Visibility    int       `json:"visibility"`
	Pop           float64   `json:"pop"`
	Rain          Rain      `json:"rain"`
	Sys           Sys       `json:"sys"`
	TimestampText string    `json:"dt_txt"`
}

type WeatherForecast struct {
	List    []ForecastEntry `json:"list"`
	COD     string          `json:"cod"`
	Message int             `json:"message"`
	Count   int             `json:"cnt"`
	City    City            `json:"city"`
}

func (w WeatherForecast) SunsetLocalTime() string {
	utcTime := time.Unix(int64(w.City.Sunset), 0).UTC()
	return utcTime.Add(time.Duration(w.City.Timezone) * time.Second).Format("15:04")

}

type WeatherSummary struct {
	// Temperature in Degree Celsius
	CurrentTemperature int `json:"temp_current"`
	FutureTemperature  int `json:"temp_future"`

	// Wind speed in km/h
	CurrentWindSpeed   int     `json:"wind_current"`
	FutureWindSpeed    int     `json:"wind_future"`
	CurrentWindDegrees int     `json:"wind_deg_current"`
	FutureWindDegrees  int     `json:"wind_deg_future"`
	CurrentWindGust    int     `json:"wind_gust_current"`
	FutureWindGust     int     `json:"wind_gust_future"`
	CurrentWindScale   float64 `json:"wind_scale_current"`
	FutureWindScale    float64 `json:"wind_scale_future"`
	CurrentRain        int     `json:"rain_current"`
	FutureRain         int     `json:"rain_future"`
	CurrentRainText    string  `json:"rain_current_text"`
	FutureRainText     string  `json:"rain_future_text"`

	// Local time
	SunsetTime string `json:"sunset"`
}

type OverpassSite struct {
	Bearing       float64
	Distance      float64
	DistanceText  string
	DistancePixel float64
	Name          string
	Website       string
}

type overpassSites []OverpassSite

func (sites *overpassSites) sortByDistance() {
	sort.Slice(*sites, func(i, j int) bool {
		return (*sites)[i].Distance < (*sites)[j].Distance
	})
}

func (sites *overpassSites) filterByBearing() {
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

	var filteredSites overpassSites

	for _, site := range *sites {
		angleFraction := int((site.Bearing + 180) / 30)
		if !angleFractions[angleFraction] {
			filteredSites = append(filteredSites, site)
			angleFractions[angleFraction] = true
		}
	}

	*sites = filteredSites
}

func newSite(siteLocation location, referenceLocation location) OverpassSite {
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
	distancePixel := minPixel + (maxPixel-minPixel)*distance/float64(maxOverpassDistance)

	return OverpassSite{
		Bearing:       referenceLocation.bearing(siteLocation),
		Distance:      distance,
		DistanceText:  distanceText,
		DistancePixel: distancePixel,
	}
}

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
