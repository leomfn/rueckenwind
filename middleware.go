package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
)

type middlewareFunc func(http.Handler) http.Handler

// Logging
func loggingMiddleware(next http.Handler) http.Handler {
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
func sameSiteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refDomain := domain
		if debug {
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
type trackingBody struct {
	Name     string `json:"name"`
	Url      string `json:"url"`
	Domain   string `json:"domain"`
	Referrer string `json:"referrer"`
}

func trackingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if debug {
			next.ServeHTTP(w, r)
			return
		}

		body := trackingBody{
			Name:     "pageview",
			Url:      r.URL.Path,
			Domain:   domain,
			Referrer: r.Referer(),
		}

		bodyJSON, err := json.Marshal(body)

		if err != nil {
			log.Println("Error marhsalling tracking body JSON")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		trackingRequest, err := http.NewRequest("POST", trackingURL, bytes.NewBuffer(bodyJSON))

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
