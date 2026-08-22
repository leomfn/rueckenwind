package main

import (
	"log"
	"os"
	"strconv"
	"strings"
)

// Default settings
var (
	port                int64  = 80
	staticFilesDir      string = "./frontend/dist"
	maxOverpassDistance int64  = 25
	overpassUserAgent   string = "rueckenwind"
	rateLimitPerMinute  int64  = 60
	rateLimitBurst      int64  = 15
	trustProxyHeaders   bool   = false
	owmApiKey           string
	debug               bool = false
	domain              string
	trackingUrl         string
	trackingId          string
)

// Reads an environment variable that must be a positive integer, falling back
// to the given default when it is not set.
func positiveIntEnv(name string, fallback int64) int64 {
	value, exists := os.LookupEnv(name)
	if !exists {
		log.Printf("%s environment variable not set, using default value: %d", name, fallback)
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Fatalf("Environment variable %s must be an integer", name)
	}

	if parsed < 1 {
		log.Fatalf("Environment variable %s must be greater than zero", name)
	}

	return parsed
}

func init() {
	var (
		err    error
		exists bool
	)

	portEnv, exists := os.LookupEnv("PORT")
	if !exists {
		log.Printf("PORT environment variable not set, using default value: %d", port)
	} else {
		port, err = strconv.ParseInt(portEnv, 10, 64)

		if err != nil {
			log.Fatal("PORT environment variable must be an integer")
		}
	}

	staticfilesDirEnv, exists := os.LookupEnv("STATIC_FILES_DIR")
	if !exists {
		log.Printf("STATIC_FILES_DIR environment variable not set, using default value: %s", staticFilesDir)
	} else {
		staticFilesDir = staticfilesDirEnv
	}

	maxOverpassDistanceEnv, exists := os.LookupEnv("MAX_OVERPASS_DISTANCE")
	if !exists {
		log.Printf("MAX_OVERPASS_DISTANCE environment variable not set, using default value: %d", maxOverpassDistance)
	} else {
		maxOverpassDistance, err = strconv.ParseInt(maxOverpassDistanceEnv, 10, 64)

		if err != nil {
			log.Fatal("Environment variable MAX_OVERPASS_DISTANCE must be an integer")
		}
	}

	overpassUserAgentEnv, exists := os.LookupEnv("OVERPASS_USER_AGENT")
	if !exists {
		log.Printf("OVERPASS_USER_AGENT environment variable not set, using default value: %s", overpassUserAgent)
	} else {
		overpassUserAgent = overpassUserAgentEnv
	}

	rateLimitPerMinute = positiveIntEnv("RATE_LIMIT_PER_MINUTE", rateLimitPerMinute)
	rateLimitBurst = positiveIntEnv("RATE_LIMIT_BURST", rateLimitBurst)

	trustProxyHeaders = strings.ToLower(os.Getenv("TRUST_PROXY_HEADERS")) == "true"
	if trustProxyHeaders {
		log.Println("Trusting X-Forwarded-For for client identification")
	}

	owmApiKey, exists = os.LookupEnv("OPEN_WEATHER_MAP_API_KEY")

	if !exists {
		log.Fatal("Environment variable OPEN_WEATHER_MAP_API_KEY not found")
	}

	domain, exists = os.LookupEnv("DOMAIN")

	if !exists {
		log.Fatal("Environment variable DOMAIN not found")
	}

	// Disable tracking if variable is not set
	trackingUrl, exists = os.LookupEnv("TRACKING_URL")

	if exists && trackingUrl != "" {
		trackingId, exists = os.LookupEnv("TRACKING_ID")

		if !exists {
			log.Fatal("Environment variable TRACKING_ID not found")
		}
	}

	debugEnv := strings.ToLower(os.Getenv("DEBUG"))
	if debugEnv == "true" {
		debug = true
		log.Println("Running in Debug mode")
	}
}
