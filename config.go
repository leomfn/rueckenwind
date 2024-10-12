package main

import (
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	staticFilesDir      string
	maxOverpassDistance int64
	owmApiKey           string
	domain              string
	trackingURL         string
	debug               bool
)

func init() {
	var (
		err    error
		exists bool
	)

	// Load environment variables
	staticFilesDir, exists = os.LookupEnv("STATIC_FILES_DIR")
	if !exists {
		log.Fatal("Environment variable STATIC_FILES_DIR not found")
	}

	maxOverpassDistanceEnv, exists := os.LookupEnv("MAX_OVERPASS_DISTANCE")
	if !exists {
		maxOverpassDistance = 25
	} else {
		maxOverpassDistance, err = strconv.ParseInt(maxOverpassDistanceEnv, 10, 64)

		if err != nil {
			log.Fatal("Environment variable MAX_OVERPASS_DISTANCE must be an integer")
		}
	}

	owmApiKey, exists = os.LookupEnv("OPEN_WEATHER_MAP_API_KEY")

	if !exists {
		log.Fatal("Environment variable OPEN_WEATHER_MAP_API_KEY not found")
	}

	domain, exists = os.LookupEnv("DOMAIN")

	if !exists {
		log.Fatal("Environment variable DOMAIN not found")
	}

	trackingURL, exists = os.LookupEnv("TRACKING_URL")

	if !exists {
		log.Fatal("Environment variable TRACKING_URL not found")
	}

	debugEnv := strings.ToLower(os.Getenv("DEBUG"))
	if debugEnv == "true" {
		debug = true
		log.Println("Running in Debug mode")
	}
}
