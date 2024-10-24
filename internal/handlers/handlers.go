package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/leomfn/rueckenwind/internal/models"
	"github.com/leomfn/rueckenwind/internal/services"
)

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
	var coordinatesBody coordinates

	err := json.NewDecoder(r.Body).Decode(&coordinatesBody)

	if err != nil {
		return errors.New("invalid request body")
	}

	h.lon = coordinatesBody.Lon
	h.lat = coordinatesBody.Lat

	return nil
}

// Weather
type weatherHandler struct {
	locationHandler
	service services.WeatherService
}

type WeatherBody struct {
	coordinates
	Category string `json:"category"`
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

func (h *weatherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.extractLocation(r)
	if err != nil {
		http.Error(w, "Could not read location", http.StatusBadRequest)
		return
	}

	userLocation := models.Location{Lon: models.Coordinate(h.lon), Lat: models.Coordinate(h.lat)}

	weatherData, err := h.service.GetWeatherForecast(float64(userLocation.Lon), float64(userLocation.Lat))
	if err != nil {
		http.Error(w, "Could not fetch weather data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(weatherData)
}

// POI sites
type poiData struct {
	Lon      float64 `json:"lon"`
	Lat      float64 `json:"lat"`
	Category string  `json:"category"`
}

type poiHandler struct {
	service services.PoiService
}

func NewPoiHandler(maxDistance int64) *poiHandler {
	return &poiHandler{
		service: services.NewOverpassPoiService(maxDistance),
	}
}

func (h *poiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var data poiData

	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	var poiResults models.OverpassSites

	switch data.Category {
	case "camping":
		poiResults, err = h.service.GetCampingPois(data.Lon, data.Lat)
	case "water":
		poiResults, err = h.service.GetDrinkingWaterPois(data.Lon, data.Lat)
	case "cafe":
		poiResults, err = h.service.GetCafePois(data.Lon, data.Lat)
	case "observation":
		poiResults, err = h.service.GetObservationPois(data.Lon, data.Lat)
	default:
		http.Error(w, "unknown category", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Println("Cloud not fetch sites data:", err)
		http.Error(w, "Error fetching sites", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(poiResults)
}
