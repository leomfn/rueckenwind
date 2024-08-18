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

type Page struct {
	Title string
	Body  []byte
}

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
	Humidity    int     `json:"humidity"`
	SeaLevel    int     `json:"sea_level"`
	GroundLevel int     `json:"grnd_level"`
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
	OneHour float64 `json:"1h"`
}

type Clouds struct {
	All int `json:"all"`
}

type Sys struct {
	Type    int    `json:"type" `
	Id      int    `json:"id"`
	Country string `json:"country"`
	Sunrise int    `json:"sunrise"`
	Sunset  int    `json:"sunset"`
}

type CurrentWeather struct {
	Coord      Coord     `json:"coord"`
	Weather    []Weather `json:"weather"`
	Base       string    `json:"base"`
	Main       Main      `json:"main"`
	Visibility int       `json:"visibility"`
	Wind       Wind      `json:"wind"`
	Rain       Rain      `json:"rain"`
	Clouds     Clouds    `json:"clouds"`
	Timestamp  int       `json:"dt"`
	Sys        Sys       `json:"sys"`
	Timezone   int       `json:"timezone"`
	Id         int       `json:"id"`
	Name       string    `json:"name"`
	COD        int       `json:"cod"`
}

func (w CurrentWeather) SunsetLocalTime() string {
	utcTime := time.Unix(int64(w.Sys.Sunset), 0).UTC()
	return utcTime.Add(time.Duration(w.Timezone) * time.Second).Format("15:04")

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

	owmUrl := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=metric", lat, lon, owmApiKey)

	resp, err := http.Get(owmUrl)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	var currentWeather CurrentWeather

	if err := json.NewDecoder(resp.Body).Decode(&currentWeather); err != nil {
		log.Println(err)
	}

	weatherSummary := WeatherSummary{
		CurrentTemperature:   int(math.Round(currentWeather.Main.Temp)),
		CurrentWindSpeed:     int(math.Round(currentWeather.Wind.Speed * 3.6)),
		CurrentWindDirection: currentWeather.Wind.directionString(),
		CurrentWindDegrees:   currentWeather.Wind.Deg,
		SunsetTime:           currentWeather.SunsetLocalTime(),
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
