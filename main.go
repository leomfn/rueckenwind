package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
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
	CurrentTemperature int `json:"temp_current"`
	FutureTemperature  int `json:"temp_future"`

	// Wind speed in km/h
	CurrentWindSpeed   int `json:"wind_current"`
	FutureWindSpeed    int `json:"wind_future"`
	CurrentWindDegrees int `json:"wind_deg_current"`
	FutureWindDegrees  int `json:"wind_deg_future"`
	CurrentWindGust    int `json:"wind_gust_current"`
	FutureWindGust     int `json:"wind_gust_future"`
	CurrentRain        int `json:"rain_current"`
	FutureRain         int `json:"rain_future"`

	// Local time
	SunsetTime string `json:"sunset"`
}

func getWeather(lat float64, lon float64) (WeatherSummary, error) {

	owmApiKey, exists := os.LookupEnv("OPEN_WEATHER_MAP_API_KEY")

	if !exists {
		log.Fatal("Environment variable OPEN_WEATHER_MAP_API_KEY not found.")
	}

	// Request weather forecast for next 12 hours in 3-hour blocks (4 items in total)
	owmUrl := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?lat=%f&lon=%f&appid=%s&units=metric&cnt=2", lat, lon, owmApiKey)

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
		CurrentRain:        currentWeather.Rain.rainIntensity(),
		FutureRain:         nextWeather.Rain.rainIntensity(),
		SunsetTime:         weatherForecast.SunsetLocalTime(),
	}

	return weatherSummary, nil
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	var ipAddress string
	var username *string

	if os.Getenv("PROXY") == "true" {
		ipAddress = r.Header.Get("X-Forwarded-For")
	} else {
		ipAddress = r.RemoteAddr
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		username = nil
	} else {
		b64Credentials := strings.TrimPrefix(authHeader, "Basic ")
		decodedCredentials, err := base64.StdEncoding.DecodeString(b64Credentials)

		if err != nil {
			log.Println("Could not decode auth credentials")
			http.Error(w, "Failed to process credentials", http.StatusUnauthorized)
			return
		}

		credentials := string(decodedCredentials)
		usernamePassword := strings.SplitN(credentials, ":", 2)
		if len(usernamePassword) != 2 {
			log.Printf("Splitting decoded credentials from auth header resulted in %d instead of 2 elements\n", len(usernamePassword))
			http.Error(w, "Failed to process credentials", http.StatusUnauthorized)
			return
		}

		username = &usernamePassword[0]
	}

	log.Printf("%s %s %s - %s", ipAddress, r.Method, r.URL, r.UserAgent())
	t.addEntry(ipAddress, username)

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(weatherSummary)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func siteStatsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("stats.html")

	if err != nil {
		http.Error(w, "Could not load stats page", http.StatusInternalServerError)
	}

	var todayVisitors, todayRequests, last7DaysVisitors, last7DaysRequests, last30DaysVisitors, last30DaysRequests int

	todayVisitors, todayRequests = t.getStats(0)
	last7DaysVisitors, last7DaysRequests = t.getStats(7)
	last30DaysVisitors, last30DaysRequests = t.getStats(30)

	statsData := struct {
		TodayVisitors      int
		TodayRequests      int
		Last7DaysVisitors  int
		Last7DaysRequests  int
		Last30DaysVisitors int
		Last30DaysRequests int
	}{
		todayVisitors,
		todayRequests,
		last7DaysVisitors,
		last7DaysRequests,
		last30DaysVisitors,
		last30DaysRequests,
	}

	tmpl.Execute(w, statsData)
}

var t *tracker

type tracker struct {
	db *sql.DB
}

func newTracker() *tracker {
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		// Make data directory if it doesn't exist yet
		err := os.Mkdir("./data", 0700)
		if err != nil {
			log.Fatal("Could not create data directory:", err)
		}
	}

	db, err := sql.Open("sqlite3", "data/stats.sqlite")
	if err != nil {
		log.Fatal("Error opening database file:", err)
	}

	createTableStatement := `
		CREATE TABLE IF NOT EXISTS stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			ip_address TEXT NOT NULL,
			username TEXT
		);
	`

	if _, err := db.Exec(createTableStatement); err != nil {
		log.Fatal("Error creating stats table:", err)
	}

	return &tracker{db: db}
}

func (tracker *tracker) addEntry(ip string, username *string) {
	stmt, err := tracker.db.Prepare("INSERT INTO stats (ip_address, username) VALUES (?, ?)")
	if err != nil {
		log.Fatal("Error writing data to stats table:", err)
	}

	if _, err := stmt.Exec(ip, &username); err != nil {
		log.Fatal(err)
	}
}

func (tracker *tracker) clean() {
	query := `
		DELETE
		FROM stats
		WHERE date(timestamp) < date(datetime('now', '-30 days'));`

	_, err := tracker.db.Exec(query)
	if err != nil {
		log.Println("Error cleaning up old entries in tracking database:", err)
	}

	log.Println("Cleaned up old entries in tracking database")
}

func (tracker *tracker) getStats(n int) (int, int) {
	query := `
	SELECT COUNT(DISTINCT ip_address) as visitors, COUNT(*) as requests 
	FROM stats
	WHERE date(timestamp) >= date(datetime('now', '-' || ? || ' days'));`

	row := tracker.db.QueryRow(query, n)
	var visitors, requests int
	err := row.Scan(&visitors, &requests)
	if err != nil {
		log.Fatal(err)
	}

	return visitors, requests
}

func main() {
	t = newTracker()

	defer t.db.Close()

	go func() {
		for {
			t.clean()
			time.Sleep(24 * time.Hour)
		}
	}()

	http.HandleFunc("/hello", statusHandler)
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/weather", weatherHandler)
	http.HandleFunc("/stats", siteStatsHandler)

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
