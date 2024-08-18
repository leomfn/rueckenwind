package main

import (
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

type Coord struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
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

func (w Wind) directionString() string {
	switch {
	case w.Deg >= 360-45/2 || w.Deg < 45/2:
		return "N"
	case w.Deg >= 45/2 && w.Deg < 45+45/2:
		return "NE"
	case w.Deg >= 45+45/2 && w.Deg < 90+45/2:
		return "E"
	case w.Deg >= 90+45/2 && w.Deg < 135+45/2:
		return "SE"
	case w.Deg >= 135+45/2 && w.Deg < 180+45/2:
		return "S"
	case w.Deg >= 180+45/2 && w.Deg < 225+45/2:
		return "SW"
	case w.Deg >= 225+45/2 && w.Deg < 270+45/2:
		return "W"
	case w.Deg >= 270+45/2 && w.Deg < 315+45/2:
		return "NW"
	}

	return ""
}

type Rain struct {
	ThreeHours float64 `json:"3h"`
}

type Clouds struct {
	All int `json:"all"`
}

type Sys struct {
	Pod string `json:"pod"`
}

type City struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	Coord      Coord  `json:"coord"`
	Country    string `json:"country"`
	Population int    `json:"population"`
	Timezone   int    `json:"timezone"`
	Sunrise    int    `json:"sunrise"`
	Sunset     int    `json:"sunset"`
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
	CurrentTemperature int `json:"temp_cur"`

	// Wind speed in km/h
	CurrentWindSpeed     int    `json:"wind_cur"`
	CurrentWindDirection string `json:"wind_dir"`
	CurrentWindDegrees   int    `json:"wind_deg"`

	// Local time
	SunsetTime string `json:"sunset"`
}

func getWeather(lat float64, lon float64) (WeatherSummary, error) {

	owmApiKey, exists := os.LookupEnv("OPEN_WEATHER_MAP_API_KEY")

	if !exists {
		log.Fatal("Environment variable OPEN_WEATHER_MAP_API_KEY not found.")
	}

	// Request weather forecast for next 12 hours in 3-hour blocks (4 items in total)
	owmUrl := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?lat=%f&lon=%f&appid=%s&units=metric&cnt=4", lat, lon, owmApiKey)

	resp, err := http.Get(owmUrl)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	var weatherForecast WeatherForecast

	if err := json.NewDecoder(resp.Body).Decode(&weatherForecast); err != nil {
		log.Println(err)
	}

	currentWeather := weatherForecast.List[0]

	weatherSummary := WeatherSummary{
		CurrentTemperature:   int(math.Round(currentWeather.Main.Temp)),
		CurrentWindSpeed:     int(math.Round(currentWeather.Wind.Speed * 3.6)),
		CurrentWindDirection: currentWeather.Wind.directionString(),
		CurrentWindDegrees:   currentWeather.Wind.Deg,
		SunsetTime:           weatherForecast.SunsetLocalTime(),
	}

	return weatherSummary, nil
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	latStr := query.Get("lat")
	lonStr := query.Get("lon")
	if latStr == "" || lonStr == "" {
		http.Error(w, "Missing query parameters: both 'lat' and 'lon' are required.", http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "Invalid latitude value.", http.StatusBadRequest)
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		http.Error(w, "Invalid longitude value.", http.StatusBadRequest)
		return
	}

	weatherSummary, err := getWeather(lat, lon)

	if err != nil {
		http.Error(w, "Could not get current weather", http.StatusInternalServerError)
	}

	t, err := template.ParseFiles("results.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Could not load results", http.StatusInternalServerError)
	}

	if err := t.Execute(w, weatherSummary); err != nil {
		log.Println(err)
		http.Error(w, "Could not load results", http.StatusInternalServerError)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func main() {
	http.HandleFunc("/hello", statusHandler)
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/weather", weatherHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
