package handlers

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/leomfn/rueckenwind/internal/services"
	"github.com/leomfn/rueckenwind/models"
)

// Index page
type getIndexHandler struct{}

func NewGetIndexHandler() *getIndexHandler {
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

func NewStaticFilesHandler(directory string) *staticFilesHandler {
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

func NewAboutHandler() *aboutHandler {
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

func NewErrorHandler() *errorHandler {
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
type weatherHandler struct {
	service services.WeatherService
}

func NewWeatherHandler(apiKey string) *weatherHandler {
	return &weatherHandler{
		service: services.NewOpenWeatherService(apiKey),
	}
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

	lonCoord := models.Coordinate(lon)
	latCoord := models.Coordinate(lat)
	userLocation := models.Location{Lon: lonCoord, Lat: latCoord}

	weatherData, err := h.service.GetWeatherForecast(float64(userLocation.Lon), float64(userLocation.Lat))
	if err != nil {
		http.Error(w, "Could not fetch weather data", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("./templates/fragments/weather.html"))
	tmpl.Execute(w, weatherData)
}

// General handler to facilitate location handling. It is not a http.Handler
// itself, but can be embedded in other handlers that expect location data to be
// sent via a form in a POST request.
type locationHandler struct {
	lon, lat float64
}

// Extracts the location coordinates from the request and stores them in the
// handler. The error returned can be used as an error message to the client.
func (h *locationHandler) extractLocation(r *http.Request) error {
	// TODO: Add input validation
	formLon := r.PostFormValue("lon")
	formLat := r.PostFormValue("lat")

	if formLon == "" || formLat == "" {
		return errors.New("lon and lat must be provided")
	}

	var lonErr, latErr error

	h.lon, lonErr = strconv.ParseFloat(formLon, 64)
	h.lat, latErr = strconv.ParseFloat(formLat, 64)

	if lonErr != nil || latErr != nil {
		return errors.New("lon and lat must be numbers")
	}

	return nil
}

// POI sites
type poiHandler struct {
	locationHandler
	service services.PoiService
}

func NewPoiHandler(maxDistance int64) *poiHandler {
	return &poiHandler{
		service: services.NewOverpassPoiService(maxDistance),
	}
}

func (h *poiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.extractLocation(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	poiCategory := r.PostFormValue("category")

	var poiResults models.OverpassSites

	switch poiCategory {
	case "camping":
		poiResults, err = h.service.GetCampingPois(h.lon, h.lat)
	case "drinking-water":
		poiResults, err = h.service.GetDrinkingWaterPois(h.lon, h.lat)
	case "cafe":
		poiResults, err = h.service.GetCafePois(h.lon, h.lat)
	default:
		http.Error(w, "unknown category", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Println("Cloud not fetch sites data:", err)
		http.Error(w, "Error fetching sites", http.StatusInternalServerError)
		return
	}

	sitesData := struct {
		Pois     models.OverpassSites
		Category string
	}{
		Pois:     poiResults,
		Category: poiCategory,
	}

	tmpl := template.Must(template.ParseFiles("./templates/fragments/pois.html"))
	tmpl.Execute(w, sitesData)
}
