package models

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"
)

func TestLocation(t *testing.T) {
	t.Run("bearing", func(t *testing.T) {
		tests := []struct {
			l1, l2          Location
			expectedBearing float64
		}{
			{Location{10, 52}, Location{10, 0}, 180},
			{Location{10, 52}, Location{11, 52}, 89.61},
		}

		for i, test := range tests {
			t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
				t.Parallel()
				actualBearing := test.l1.bearing(test.l2)

				// round to 2 decimal places
				if math.Round(actualBearing*100)/100 != test.expectedBearing {
					t.Fatalf("expected bearing %f, but got %f", test.expectedBearing, actualBearing)
				}
			})
		}
	})

	t.Run("distance", func(t *testing.T) {
		tests := []struct {
			name             string
			l1, l2           Location
			expectedDistance float64
		}{
			{"same point", Location{10, 52}, Location{10, 52}, 0},
			// One degree of latitude is about 111.19 km anywhere on the globe.
			{"one degree of latitude", Location{10, 52}, Location{10, 53}, 111.19},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				actual := test.l1.distance(test.l2)

				if math.Abs(actual-test.expectedDistance) > 0.01 {
					t.Fatalf("expected distance %f, but got %f", test.expectedDistance, actual)
				}
			})
		}
	})
}

func TestOverpassSites(t *testing.T) {
	t.Run("SortByDistance", func(t *testing.T) {
		sites := OverpassSites{
			{Distance: 1.1},
			{Distance: 1},
			{Distance: 5},
			{Distance: 10},
			{Distance: 0.009},
			{Distance: 10},
			{Distance: 1000000},
		}

		sites.SortByDistance()

		for i := 0; i < len(sites)-1; i++ {
			lowerDistance := sites[i].Distance
			higherDistance := sites[i+1].Distance
			if lowerDistance > higherDistance {
				t.Fatalf("%f is not supposed to be greater than %f", lowerDistance, higherDistance)
			}
		}
	})

	t.Run("FilterByBearing", func(t *testing.T) {
		sites := OverpassSites{
			{Bearing: 0},
			{Bearing: 15},
			{Bearing: 29.9999},
			{Bearing: 30},
			{Bearing: 30.0001},
			{Bearing: 100},
			{Bearing: 359.9999},
		}

		sites.FilterByBearing()

		expectedLen := 4
		actualLen := len(sites)

		if actualLen != expectedLen {
			t.Fatalf("Expected %d results after filtering, but got %d", expectedLen, actualLen)
		}
	})

	// Bearings come from Location.bearing, which returns values in [-180, 180].
	t.Run("FilterByBearing covers the whole bearing range", func(t *testing.T) {
		sites := OverpassSites{}
		for bearing := -180.0; bearing <= 180.0; bearing += 1 {
			sites = append(sites, overpassSite{Bearing: bearing})
		}

		sites.FilterByBearing()

		if len(sites) != bearingBuckets {
			t.Fatalf("expected one POI per bucket (%d), but got %d", bearingBuckets, len(sites))
		}
	})

	t.Run("FilterByBearing keeps the first POI in each bucket", func(t *testing.T) {
		sites := OverpassSites{
			{Bearing: -180, Name: "first in bucket 0"},
			{Bearing: -175, Name: "also bucket 0"},
			{Bearing: -150, Name: "first in bucket 1"},
		}

		sites.FilterByBearing()

		if len(sites) != 2 {
			t.Fatalf("expected 2 results, but got %d", len(sites))
		}

		if sites[0].Name != "first in bucket 0" || sites[1].Name != "first in bucket 1" {
			t.Errorf("unexpected survivors: %q and %q", sites[0].Name, sites[1].Name)
		}
	})
}

// An empty result must reach the client as [] rather than null, which the
// frontend cannot iterate over.
func TestFilterByBearingEncodesEmptyResultAsList(t *testing.T) {
	sites := OverpassSites{}
	sites.FilterByBearing()

	encoded, err := json.Marshal(sites)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	if string(encoded) != "[]" {
		t.Errorf("expected [], but got %s", encoded)
	}
}

func TestBearingBucket(t *testing.T) {
	tests := []struct {
		bearing  float64
		expected int
	}{
		{-180, 0},
		{-150.0001, 0},
		{-150, 1},
		{0, 6},
		{149.9999, 10},
		{150, 11},
		{179.9999, 11},
		// Due south is the same direction as -180, so it folds back onto the
		// first bucket rather than falling off the end.
		{180, 0},
	}

	for _, test := range tests {
		if actual := bearingBucket(test.bearing); actual != test.expected {
			t.Errorf("expected bearing %v in bucket %d, but got %d", test.bearing, test.expected, actual)
		}
	}
}

func TestWindScale(t *testing.T) {
	// Speeds are in m/s, the scale saturates at 80 km/h.
	atCap := Wind{Speed: 80 / 3.6}.Scale()

	t.Run("is capped", func(t *testing.T) {
		for _, speed := range []float64{80 / 3.6, 50, 100, 1000} {
			if scale := (Wind{Speed: speed}).Scale(); scale > atCap {
				t.Errorf("expected %v m/s to be capped at %v, but got %v", speed, atCap, scale)
			}
		}
	})

	t.Run("increases with speed", func(t *testing.T) {
		previous := (Wind{Speed: 0}).Scale()

		for speed := 1.0; speed <= 22; speed++ {
			current := (Wind{Speed: speed}).Scale()

			if current <= previous {
				t.Fatalf("expected the scale to grow at %v m/s, but it went from %v to %v",
					speed, previous, current)
			}

			previous = current
		}
	})

	t.Run("stays within bounds", func(t *testing.T) {
		for _, speed := range []float64{0, 5, 20, 100} {
			scale := (Wind{Speed: speed}).Scale()

			if scale < 0.2 || scale > 1 {
				t.Errorf("expected a scale between 0.2 and 1 at %v m/s, but got %v", speed, scale)
			}
		}
	})
}

func TestRainIntensity(t *testing.T) {
	// Values are millimetres over three hours; the thresholds are on the hourly
	// rate. They sit well inside each band, because the exact boundaries are
	// subject to floating point rounding and are not meaningful to a reader.
	tests := []struct {
		threeHours   float64
		expected     int64
		expectedText string
	}{
		{0, 0, "dry"},
		{0.15, 1, "light"},
		{0.45, 2, "medium"},
		{1.2, 2, "medium"},
		{1.8, 3, "heavy"},
		{30, 3, "heavy"},
	}

	for _, test := range tests {
		rain := Rain{ThreeHours: test.threeHours}

		if actual := rain.RainIntensity(); actual != test.expected {
			t.Errorf("expected intensity %d for %v mm/3h, but got %d",
				test.expected, test.threeHours, actual)
		}

		if actual := rain.RainText(); actual != test.expectedText {
			t.Errorf("expected text %q for %v mm/3h, but got %q",
				test.expectedText, test.threeHours, actual)
		}
	}
}
