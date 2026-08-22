package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/leomfn/rueckenwind/internal/models"
)

// Records the coordinates it was called with, so a handler that leaks state
// between concurrent requests can be detected.
type stubWeatherService struct{}

func (stubWeatherService) GetWeatherForecast(ctx context.Context, lon float64, lat float64) (models.WeatherSummary, error) {
	// Echo the coordinates back so the caller can verify it got its own answer.
	return models.WeatherSummary{
		CurrentTemperature: int64(lat),
		FutureTemperature:  int64(lon),
	}, nil
}

// Handlers are registered once and serve every request, so per-request data
// must never be stored on the handler itself. Run with -race.
func TestWeatherHandlerIsRequestScoped(t *testing.T) {
	handler := &weatherHandler{service: stubWeatherService{}}

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Go(func() {
			// Distinct but valid coordinates, so each request can be told apart.
			lon, lat := float64(i), float64(i+20)

			body := fmt.Sprintf(`{"lon":%v,"lat":%v}`, lon, lat)
			request := httptest.NewRequest(http.MethodPost, "/data/weather", strings.NewReader(body))
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			var summary models.WeatherSummary
			if err := json.NewDecoder(recorder.Body).Decode(&summary); err != nil {
				t.Errorf("could not decode response: %v", err)
				return
			}

			if summary.FutureTemperature != int64(lon) || summary.CurrentTemperature != int64(lat) {
				t.Errorf("request for (%v, %v) was answered with (%v, %v)",
					lon, lat, summary.FutureTemperature, summary.CurrentTemperature)
			}
		})
	}
	wg.Wait()
}

func TestExtractLocation(t *testing.T) {
	t.Run("valid body", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"lon":10.5,"lat":52.5}`))

		location, err := extractLocation(request)
		if err != nil {
			t.Fatalf("expected no error, but got %v", err)
		}

		if location.Lon != 10.5 || location.Lat != 52.5 {
			t.Fatalf("expected (10.5, 52.5), but got (%v, %v)", location.Lon, location.Lat)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`not json`))

		if _, err := extractLocation(request); err == nil {
			t.Fatal("expected an error for a malformed body, but got none")
		}
	})

	t.Run("out of range coordinates are rejected", func(t *testing.T) {
		bodies := []string{
			`{"lon":10,"lat":91}`,
			`{"lon":10,"lat":-91}`,
			`{"lon":181,"lat":52}`,
			`{"lon":-181,"lat":52}`,
			`{"lon":1e300,"lat":1e300}`,
		}

		for _, body := range bodies {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

			if _, err := extractLocation(request); err == nil {
				t.Errorf("expected %s to be rejected, but it was accepted", body)
			}
		}
	})

	t.Run("edge coordinates are accepted", func(t *testing.T) {
		bodies := []string{
			`{"lon":180,"lat":90}`,
			`{"lon":-180,"lat":-90}`,
			`{"lon":0,"lat":0}`,
		}

		for _, body := range bodies {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

			if _, err := extractLocation(request); err != nil {
				t.Errorf("expected %s to be accepted, but got %v", body, err)
			}
		}
	})
}

func TestCoordinatesValidateRejectsNonFinite(t *testing.T) {
	tests := []coordinates{
		{Lon: math.NaN(), Lat: 52},
		{Lon: 10, Lat: math.NaN()},
		{Lon: math.Inf(1), Lat: 52},
		{Lon: 10, Lat: math.Inf(-1)},
	}

	for _, test := range tests {
		if err := test.validate(); err == nil {
			t.Errorf("expected (%v, %v) to be rejected, but it was accepted", test.Lon, test.Lat)
		}
	}
}
