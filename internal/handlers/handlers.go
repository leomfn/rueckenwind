package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"

	"github.com/leomfn/rueckenwind/internal/services"
)

// Healthcheck handler
type healthcheckHandler struct{}

func NewHealthcheckHandler() *healthcheckHandler {
	return &healthcheckHandler{}
}

type healthcheckResponse struct {
	Status string `json:"status"`
}

func (h *healthcheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := healthcheckResponse{
		Status: "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Index page
type getIndexHandler struct {
	directory string
}

func NewGetIndexHandler(directory string) *getIndexHandler {
	return &getIndexHandler{
		directory: directory,
	}
}

func (h getIndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.directory)
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
	http.StripPrefix("/assets/", staticFileserver).ServeHTTP(w, r)
}

// General handlers

// General POST handler that reads application/json data
// TODO: generalize handlers that read json
// type postHandler struct {
// 	data interface{}
// }

// func (h *postHandler) readJSONPayload(w http.ResponseWriter, r *http.Request) error {
// 	err := json.NewDecoder(r.Body).Decode(&h.data)

// 	if err != nil {
// 		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
// 		return nil
// 	}

// 	return err
// }

// Extracts the location coordinates from the request body. Handlers are shared
// between concurrent requests, so the coordinates are returned rather than
// stored on the handler. The error returned can be used as an error message to
// the client.
func extractLocation(r *http.Request) (coordinates, error) {
	var coordinatesBody coordinates

	if err := json.NewDecoder(r.Body).Decode(&coordinatesBody); err != nil {
		return coordinates{}, errors.New("invalid request body")
	}

	if err := coordinatesBody.validate(); err != nil {
		return coordinates{}, err
	}

	return coordinatesBody, nil
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

type coordinates struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

// Rejects coordinates that are not finite or outside the valid range, so that
// they can never be written into an upstream query. Missing fields decode to
// zero, which is a valid location, so they are accepted.
func (c coordinates) validate() error {
	if math.IsNaN(c.Lat) || math.IsNaN(c.Lon) || math.IsInf(c.Lat, 0) || math.IsInf(c.Lon, 0) {
		return errors.New("coordinates must be finite numbers")
	}

	if c.Lat < -90 || c.Lat > 90 {
		return errors.New("latitude must be between -90 and 90")
	}

	if c.Lon < -180 || c.Lon > 180 {
		return errors.New("longitude must be between -180 and 180")
	}

	return nil
}

func (h *weatherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	location, err := extractLocation(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	weatherData, err := h.service.GetWeatherForecast(r.Context(), location.Lon, location.Lat)
	if err != nil {
		http.Error(w, "Could not fetch weather data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(weatherData)
}

// POI sites
type poiData struct {
	coordinates
	Category string `json:"category"`
}

type poiHandler struct {
	service services.PoiService
}

func NewPoiHandler(maxDistance int64, userAgent string) *poiHandler {
	return &poiHandler{
		service: services.NewOverpassPoiService(maxDistance, userAgent),
	}
}

func (h *poiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var data poiData

	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := data.validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	poiResults, err := h.service.GetPois(r.Context(), data.Category, data.Lon, data.Lat)

	if errors.Is(err, services.ErrUnknownCategory) {
		http.Error(w, "unknown category", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Println("Could not fetch sites data:", err)
		http.Error(w, "Error fetching sites", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(poiResults)
}
