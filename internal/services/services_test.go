package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestWeatherService(url string) *openWeatherService {
	return &openWeatherService{
		client:           &http.Client{Timeout: weatherRequestTimeout},
		forecastUrl:      url,
		apiKey:           "test-key",
		maxForecastCount: 2,
	}
}

// A truncated or error response from openweather must be reported as an error
// rather than being indexed into.
func TestGetWeatherForecastRejectsUnusableResponses(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{"empty forecast list", http.StatusOK, `{"cod":"200","list":[]}`},
		{"single forecast entry", http.StatusOK, `{"cod":"200","list":[{"dt":1}]}`},
		{"missing forecast list", http.StatusOK, `{"cod":"200"}`},
		{"invalid api key", http.StatusUnauthorized, `{"cod":401,"message":"Invalid API key"}`},
		{"rate limited", http.StatusTooManyRequests, `{"cod":429,"message":"limit reached"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.status)
				w.Write([]byte(test.body))
			}))
			defer server.Close()

			service := newTestWeatherService(server.URL)

			if _, err := service.GetWeatherForecast(context.Background(), 10, 52); err == nil {
				t.Fatal("expected an error, but got none")
			}
		})
	}
}

func TestGetWeatherForecastSummarizesTwoEntries(t *testing.T) {
	body := `{"cod":"200","list":[
		{"main":{"temp":12.4},"wind":{"speed":5,"deg":180,"gust":7},"rain":{"3h":0}},
		{"main":{"temp":15.6},"wind":{"speed":10,"deg":200,"gust":12},"rain":{"3h":3}}
	],"city":{"sunset":1700000000,"timezone":0}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	}))
	defer server.Close()

	service := newTestWeatherService(server.URL)

	summary, err := service.GetWeatherForecast(context.Background(), 10, 52)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	if summary.CurrentTemperature != 12 || summary.FutureTemperature != 16 {
		t.Errorf("expected temperatures (12, 16), but got (%d, %d)",
			summary.CurrentTemperature, summary.FutureTemperature)
	}

	// 5 m/s is 18 km/h
	if summary.CurrentWindSpeed != 18 {
		t.Errorf("expected current wind speed 18 km/h, but got %d", summary.CurrentWindSpeed)
	}

	if summary.CurrentRainText != "dry" || summary.FutureRainText != "heavy" {
		t.Errorf("expected rain (dry, heavy), but got (%s, %s)",
			summary.CurrentRainText, summary.FutureRainText)
	}
}

// A request that is cancelled by the client must not leave the upstream call
// running.
func TestGetWeatherForecastHonoursContextCancellation(t *testing.T) {
	blocked := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocked
	}))
	defer server.Close()
	defer close(blocked)

	service := newTestWeatherService(server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := service.GetWeatherForecast(ctx, 10, 52); err == nil {
		t.Fatal("expected an error for a cancelled context, but got none")
	}
}
