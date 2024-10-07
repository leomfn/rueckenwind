package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	maxOverpassDistance int64
	owmApiKey           string
	domain              string
	trackingURL         string
)

func init() {
	// Load environment variables
	maxOverpassDistanceEnv, exists := os.LookupEnv("MAX_OVERPASS_DISTANCE")
	if !exists {
		maxOverpassDistance = 25
	} else {
		var err error
		maxOverpassDistance, err = strconv.ParseInt(maxOverpassDistanceEnv, 10, 64)

		if err != nil {
			log.Fatal("Environment variable MAX_OVERPASS_DISTANCE must be an integer.")
		}
	}

	owmApiKey, exists = os.LookupEnv("OPEN_WEATHER_MAP_API_KEY")

	if !exists {
		log.Fatal("Environment variable OPEN_WEATHER_MAP_API_KEY not found.")
	}

	domain, exists = os.LookupEnv("DOMAIN")

	if !exists {
		log.Fatal("Environment variable DOMAIN not found.")
	}

	trackingURL, exists = os.LookupEnv("TRACKING_URL")

	if !exists {
		log.Fatal("Environment variable TRACKING_URL not found.")
	}
}

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

type locationInfo struct {
	WeatherSummary
	Campsites []Campsite
}

type Campsite struct {
	Bearing       float64
	Distance      float64
	DistanceText  string
	DistancePixel float64
	Name          string
	Website       string
}

func newCampsite(campsiteLocation location, referenceLocation location) Campsite {
	distance := referenceLocation.distance(campsiteLocation)
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

	return Campsite{
		Bearing:       referenceLocation.bearing(campsiteLocation),
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

func getCampsites(referenceLocation location) ([]Campsite, error) {
	// TODO: maybe useful for debugging/testing:

	// lon1 := coordinate(11.0)
	// lat1 := coordinate(50.0)
	// lon2 := coordinate(13.0)
	// lat2 := coordinate(53.0)
	// lon3 := coordinate(9.708282)
	// lat3 := coordinate(52.374027)

	// camp1 := newCampsite(location{Lon: &lon1, Lat: &lat1}, referenceLocation)
	// camp2 := newCampsite(location{Lon: &lon2, Lat: &lat2}, referenceLocation)
	// camp3 := newCampsite(location{Lon: &lon3, Lat: &lat3}, referenceLocation)

	// return []Campsite{camp1, camp2, camp3}, nil

	overpassURL := "https://overpass-api.de/api/interpreter"

	query := fmt.Sprintf(`[out:json];nwr["tourism"="camp_site"]["tent"!="no"](around:%d,%v,%v);out geom;`, maxOverpassDistance*1000, *referenceLocation.Lat, *referenceLocation.Lon)

	resp, err := http.Post(overpassURL, "text/plain", bytes.NewBuffer([]byte(query)))

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

	campsites := []Campsite{}

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

		campsite := newCampsite(location{Lon: &lon, Lat: &lat}, referenceLocation)
		campsite.Name = element.Tags.Name
		campsite.Website = element.Tags.Website

		campsites = append(campsites, campsite)
	}

	return campsites, nil
}

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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/index.html"))
	data := struct {
		Color   string
		Message string
	}{
		Color:   "blue",
		Message: "Welcome",
	}
	tmpl.Execute(w, data)
}

func positionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	formLon := r.PostFormValue("lon")
	formLat := r.PostFormValue("lat")

	if formLon == "" || formLat == "" {
		http.Error(w, "lon and lat must be provided", http.StatusBadRequest)
		return
	}

	lon, lonErr := strconv.ParseFloat(formLon, 64)
	lat, latErr := strconv.ParseFloat(formLat, 64)

	if lonErr != nil || latErr != nil {
		http.Error(w, "lon and lat must be numbers", http.StatusBadRequest)
		return
	}

	lonCoord := coordinate(lon)
	latCoord := coordinate(lat)
	userLocation := location{Lon: &lonCoord, Lat: &latCoord}

	weatherSummary, err := getWeather(*userLocation.Lat, *userLocation.Lon)
	if err != nil {
		http.Error(w, "Could not fetch weather data", http.StatusInternalServerError)
		return
	}

	campsites, err := getCampsites(userLocation)
	if err != nil {
		http.Error(w, "Could not fetch campsite data", http.StatusInternalServerError)
		return
	}
	locationInfo := locationInfo{WeatherSummary: weatherSummary, Campsites: campsites}

	tmpl := template.Must(template.ParseFiles("./templates/fragments/position.html"))
	tmpl.Execute(w, locationInfo)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Cache-Control", "max-age=86400") // 1 day

	tmpl := template.Must(template.ParseFiles("./templates/fragments/info-modal.html"))
	tmpl.Execute(w, nil)
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	errorType := query.Get("type")
	if errorType == "" {
		http.Error(w, "Missing query parameter 'type'", http.StatusBadRequest)
		return
	}

	var errorMessage, errorTitle string

	switch errorType {
	case "location":
		errorMessage = "This site doesn't work without location permission."
		errorTitle = "Error"
	case "orientation":
		errorMessage = "Active user input is necessary to access your iPhone's orientation sensors for the compass to work correctly. Touch the compass to confirm."
		errorTitle = "iOS Information"
	default:
		http.Error(w, "Unknown error type", http.StatusBadRequest)
		return
	}

	data := struct {
		ErrorTitle   string
		ErrorType    string
		ErrorMessage string
	}{
		ErrorTitle:   errorTitle,
		ErrorType:    errorType,
		ErrorMessage: errorMessage,
	}

	// w.Header().Set("Cache-Control", "max-age=86400") // 1 day

	tmpl := template.Must(template.ParseFiles("./templates/fragments/error-modal.html"))
	tmpl.Execute(w, data)
}

type trackingBody struct {
	Name     string `json:"name"`
	Url      string `json:"url"`
	Domain   string `json:"domain"`
	Referrer string `json:"referrer"`
}

func trackingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := trackingBody{
			Name:     "pageview",
			Url:      r.URL.Path,
			Domain:   domain,
			Referrer: r.Referer(),
		}

		bodyJSON, err := json.Marshal(body)

		if err != nil {
			log.Println("Error marhsalling tracking body JSON")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		trackingRequest, err := http.NewRequest("POST", trackingURL, bytes.NewBuffer(bodyJSON))

		if err != nil {
			log.Println("Failed to create POST request to tracking server:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		trackingRequest.Header.Add("Content-Type", "application/json")
		trackingRequest.Header.Add("User-Agent", r.Header.Get("User-Agent"))
		trackingRequest.Header.Add("X-Forwarded-For", r.Header.Get("X-Forwarded-For"))

		client := &http.Client{}

		trackingResponse, err := client.Do(trackingRequest)

		if err != nil || trackingResponse.StatusCode != 202 {
			log.Println("Failed sending POST request to tracking server:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	staticFileserver := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", staticFileserver))

	http.Handle("/", trackingMiddleware(http.HandlerFunc(indexHandler)))
	http.HandleFunc("/position", positionHandler)
	http.Handle("/about", trackingMiddleware(http.HandlerFunc(aboutHandler)))
	http.HandleFunc("/error", errorHandler)

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
