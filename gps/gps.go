package gps

import (
	"github.com/NOVAPokemon/utils"
	"math"
)

// Earth radius in meters
const earthRadius = 6378000

// Add distance in meters to location. Both dLat and dLong are assumed to be in meters.
func CalcLocationPlusDistanceTraveled(location utils.Location, dLat, dLong float64) utils.Location {
	newLat := location.Latitude + (dLat/earthRadius)*(180.0/math.Pi)
	newLong := location.Longitude + (dLong/earthRadius)*(180.0/math.Pi)/math.Cos(location.Latitude*math.Pi/180)

	return utils.Location{
		Latitude:  newLat,
		Longitude: newLong,
	}
}

// Using distance in straight line not contemplating earth's curvature
func CalcDistanceBetweenLocations(locationA, locationB utils.Location) float64 {
	dLat := locationB.Latitude - locationA.Latitude
	dLong := locationB.Longitude - locationA.Longitude

	return math.Sqrt(math.Pow(dLat, 2)+math.Pow(dLong, 2)) * (math.Pi/180.0) * earthRadius
}
