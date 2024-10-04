package main

import (
	"log"
	"testing"
)

func TestGetCampsites(t *testing.T) {
	lat := coordinate(52)
	lon := coordinate(10)
	location := location{Lon: &lon, Lat: &lat}

	campsites, err := getCampsites(location)
	if err != nil {
		t.Fatal(err)
	}

	log.Println(campsites)
}
