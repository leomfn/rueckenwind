package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
)

// Index page
type getIndexHandler struct{}

func newGetIndexHandler() *getIndexHandler {
	return &getIndexHandler{}
}

func (h getIndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./templates/index.html"))
	tmpl.Execute(w, nil)
}

// Serve static files
type staticFilesHandler struct {
	directory http.Dir
}

func newStaticFilesHandler(directory string) *staticFilesHandler {
	return &staticFilesHandler{
		directory: http.Dir(directory),
	}
}

func (h *staticFilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	staticFileserver := http.FileServer(h.directory)
	http.StripPrefix("/static/", staticFileserver).ServeHTTP(w, r)
}

// About modal
type aboutHandler struct{}

func newAboutHandler() *aboutHandler {
	return &aboutHandler{}
}

func (h *aboutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: decide on setting cache control
	// w.Header().Set("Cache-Control", "max-age=86400") // 1 day

	tmpl := template.Must(template.ParseFiles("./templates/fragments/info-modal.html"))
	tmpl.Execute(w, nil)
}

// Error modal
type errorHandler struct{}

func newErrorHandler() *errorHandler {
	return &errorHandler{}
}

func (h *errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	// TODO: decide on setting cache control
	// w.Header().Set("Cache-Control", "max-age=86400") // 1 day

	tmpl := template.Must(template.ParseFiles("./templates/fragments/error-modal.html"))
	tmpl.Execute(w, data)
}

// Weather
type weatherHandler struct{}

func newWeatherHandler() *weatherHandler {
	return &weatherHandler{}
}

func (h *weatherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	weatherData, err := getWeather(*userLocation.Lat, *userLocation.Lon)
	if err != nil {
		http.Error(w, "Could not fetch weather data", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/fragments/weather.html"))
	tmpl.Execute(w, weatherData)
}

// POI sites
type sitesHandler struct{}

func newSitesHandler() *sitesHandler {
	return &sitesHandler{}
}

func (h *sitesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	overpassQuery := newOverpassQuery(userLocation)
	err := overpassQuery.execute()

	if err != nil {
		log.Println("Cloud not fetch sites data:", err)
		http.Error(w, "Error fetching sites", http.StatusInternalServerError)
		return
	}

	sitesData := struct {
		CampSites          overpassSites
		DrinkingWaterSites overpassSites
		CafeSites          overpassSites
	}{
		CampSites:          overpassQuery.campSites,
		DrinkingWaterSites: overpassQuery.drinkingWaterSites,
		CafeSites:          overpassQuery.cafeSites,
	}

	tmpl := template.Must(template.ParseFiles("./templates/fragments/sites.html"))
	tmpl.Execute(w, sitesData)
}
