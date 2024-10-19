package middleware

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
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
	debug       bool
}

func NewTrackingMiddleware(domain string, trackingURL string, debug bool) Middleware {
	return &trackingMiddleware{
		domain:      domain,
		trackingURL: trackingURL,
		debug:       debug,
	}
}

// TODO: is this really necessary here?
type trackingBody struct {
	Name     string `json:"name"`
	Url      string `json:"url"`
	Domain   string `json:"domain"`
	Referrer string `json:"referrer"`
}

func (m *trackingMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.debug {
			next.ServeHTTP(w, r)
			return
		}

		body := trackingBody{
			Name:     "pageview",
			Url:      r.URL.Path,
			Domain:   m.domain,
			Referrer: r.Referer(),
		}

		bodyJSON, err := json.Marshal(body)

		if err != nil {
			log.Println("Error marhsalling tracking body JSON")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		trackingRequest, err := http.NewRequest("POST", m.trackingURL, bytes.NewBuffer(bodyJSON))

		if err != nil {
			log.Println("Failed to create POST request to tracking server:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		trackingRequest.Header.Add("Content-Type", "application/json")
		trackingRequest.Header.Add("User-Agent", r.Header.Get("User-Agent"))
		trackingRequest.Header.Add("X-Forwarded-For", r.Header.Get("X-Forwarded-For"))

		client := &http.Client{}

		trackingResponse, err := client.Do(trackingRequest)

		if err != nil || trackingResponse.StatusCode != 202 {
			log.Println("Failed sending POST request to tracking server:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r)
	})
}
