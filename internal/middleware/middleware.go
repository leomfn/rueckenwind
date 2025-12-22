package middleware

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Middleware interface {
	MiddlewareFunc(http.Handler) http.Handler
}

// Logging
type loggingMiddleware struct{}

func NewLoggingMiddleware() Middleware {
	return &loggingMiddleware{}
}

func (m *loggingMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// Same site protection
//
// Prevent simple requests to endpoints which are meant to be requested from the
// specified domain only. This is not a real security measure, because this can
// be easily circumvented by spoofing the header. Also, this might be
// problematic for privacy-focused browsers or proxies, that remove the Referer
// header.
//
// TODO: this must be tested in live environment
type sameSiteMiddleware struct {
	debug  bool
	domain string
}

func NewSameSiteMiddleware(domain string, debug bool) Middleware {
	return &sameSiteMiddleware{
		debug:  debug,
		domain: domain,
	}
}

func (m *sameSiteMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.debug {
			next.ServeHTTP(w, r)
			return
		}

		refDomain := m.domain
		if m.debug {
			refDomain = "localhost"
		}

		refHeader := r.Referer()

		requestReferrerURL, err := url.Parse(refHeader)

		if err != nil || requestReferrerURL.Hostname() != refDomain {
			log.Printf("Access to %s blocked, invalid referrer '%s'", r.URL.Path, refHeader)
			http.Error(w, "Invalid Referer", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Tracking
type trackingMiddleware struct {
	domain      string
	trackingURL string
	websiteId   string
	debug       bool
}

func NewTrackingMiddleware(domain string, trackingURL string, websiteId string, debug bool) Middleware {
	return &trackingMiddleware{
		domain:      domain,
		trackingURL: trackingURL,
		websiteId:   websiteId,
		debug:       debug,
	}
}

type umamiRequest struct {
	Type    string       `json:"type"`
	Payload umamiPayload `json:"payload"`
}

type umamiPayload struct {
	Website  string `json:"website"`
	Hostname string `json:"hostname"`
	URL      string `json:"url"`
	Title    string `json:"title,omitempty"`
	Referrer string `json:"referrer,omitempty"`
	Language string `json:"language,omitempty"`
	Screen   string `json:"screen,omitempty"`
}

func newTrackingBody(r *http.Request, websiteID, hostname string) umamiRequest {
	language := ""
	if acceptLang := r.Header.Get("Accept-Language"); acceptLang != "" {
		if idx := strings.Index(acceptLang, ","); idx != -1 {
			language = acceptLang[:idx]
		} else {
			language = acceptLang
		}
	}

	return umamiRequest{
		Type: "event",
		Payload: umamiPayload{
			Website:  websiteID,
			Hostname: hostname,
			URL:      r.URL.Path,
			Referrer: r.Referer(),
			Language: language,
		},
	}
}

func (m *trackingMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.debug || m.trackingURL == "" {
			next.ServeHTTP(w, r)
			return
		}

		trackingBody := newTrackingBody(r, m.websiteId, m.domain)

		bodyJSON, err := json.Marshal(trackingBody)
		if err != nil {
			log.Println("Error marshalling tracking body JSON:", err)
			next.ServeHTTP(w, r)
			return
		}

		trackingRequest, err := http.NewRequest("POST", m.trackingURL, bytes.NewBuffer(bodyJSON))
		if err != nil {
			log.Println("Failed to create POST request to tracking server:", err)
			next.ServeHTTP(w, r)
			return
		}

		trackingRequest.Header.Add("Content-Type", "application/json")
		trackingRequest.Header.Add("User-Agent", r.Header.Get("User-Agent"))
		trackingRequest.Header.Add("X-Forwarded-For", r.Header.Get("X-Forwarded-For"))

		client := &http.Client{
			Timeout: 5 * time.Second, // Add timeout to avoid hanging
		}

		// Send tracking request asynchronously to not block the main request
		go func() {
			trackingResponse, err := client.Do(trackingRequest)
			if err != nil {
				log.Println("Failed sending POST request to tracking server:", err)
				return
			}
			defer trackingResponse.Body.Close()

			if trackingResponse.StatusCode < 200 || trackingResponse.StatusCode >= 300 {
				log.Printf("Tracking server returned status code: %d", trackingResponse.StatusCode)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
