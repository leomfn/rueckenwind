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
	owmApiKey           string
	debug               bool = false
	domain              string
	trackingUrl         string
)

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

	owmApiKey, exists = os.LookupEnv("OPEN_WEATHER_MAP_API_KEY")

	if !exists {
		log.Fatal("Environment variable OPEN_WEATHER_MAP_API_KEY not found")
	}

	domain, exists = os.LookupEnv("DOMAIN")

	if !exists {
		log.Fatal("Environment variable DOMAIN not found")
	}

	// Disable tracking if variable is not set
	trackingUrl = os.Getenv("TRACKING_URL")

	debugEnv := strings.ToLower(os.Getenv("DEBUG"))
	if debugEnv == "true" {
		debug = true
		log.Println("Running in Debug mode")
	}
}
