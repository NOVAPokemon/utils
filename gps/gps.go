package gps

import (
	"math"

	"github.com/golang/geo/s2"
)

// Earth radius in meters
const earthRadius = 6378000

// Add distance in meters to location. Both dLat and dLong are assumed to be in meters.
func CalcLocationPlusDistanceTraveled(location s2.LatLng, dLat, dLong float64) s2.LatLng {
	newLat := location.Lat.Degrees() + (dLat/earthRadius)*(180.0/math.Pi)
	newLong := location.Lng.Degrees() +
		(dLong/earthRadius)*(180.0/math.Pi)/math.Cos(location.Lat.Degrees()*math.Pi/180)

	return s2.LatLngFromDegrees(newLat, newLong)
}
