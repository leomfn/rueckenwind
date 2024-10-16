package models

import (
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
}

func TestPois(t *testing.T) {
	t.Run("sortByDistance", func(t *testing.T) {
		pois := Pois{}
		pois.pois = []poi{
			{distance: 1.1},
			{distance: 1},
			{distance: 5},
			{distance: 10},
			{distance: 0.009},
			{distance: 10},
			{distance: 1000000},
		}

		pois.sortByDistance()

		for i := 0; i < len(pois.pois)-1; i++ {
			lowerDistance := pois.pois[i].distance
			higherDistance := pois.pois[i+1].distance
			if lowerDistance > higherDistance {
				t.Fatalf("%f is not supposed to be greater than %f", lowerDistance, higherDistance)
			}
		}
	})

	t.Run("filterByBearing", func(t *testing.T) {
		pois := Pois{}
		pois.pois = []poi{
			{bearing: 0},
			{bearing: 15},
			{bearing: 29.9999},
			{bearing: 30},
			{bearing: 30.0001},
			{bearing: 100},
			{bearing: 359.9999},
		}

		pois.filterByBearing()

		expectedLen := 4
		actualLen := len(pois.pois)

		if actualLen != expectedLen {
			t.Fatalf("Expected %d results after filtering, but got %d", expectedLen, actualLen)
		}
	})
}
