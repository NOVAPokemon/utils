package location

import (
	"fmt"

	"github.com/NOVAPokemon/utils"
	"github.com/golang/geo/r2"
)

func IsWithinBounds(location utils.Location, topLeft utils.Location, botRight utils.Location) bool {
	topLeftPoint := r2.Point{
		X: topLeft.Longitude,
		Y: topLeft.Latitude,
	}

	botRightPoint := r2.Point{
		X: botRight.Longitude,
		Y: botRight.Latitude,
	}

	boundsRect := r2.RectFromPoints(topLeftPoint, botRightPoint)

	locPoint := r2.Point{
		X: location.Longitude,
		Y: location.Latitude,
	}

	fmt.Printf("checking if {%f, %f} in {%f, %f}{%f, %f}\n", location.Longitude, location.Latitude,
		topLeft.Longitude, topLeft.Latitude, botRight.Longitude, botRight.Latitude)

	return boundsRect.ContainsPoint(locPoint)
}

func LocationsToRect(topLeft, botRight utils.Location) r2.Rect {
	topLeftPoint := r2.Point{
		X: topLeft.Longitude,
		Y: topLeft.Latitude,
	}
	bottomRightPoint := r2.Point{
		X: botRight.Longitude,
		Y: botRight.Latitude,
	}

	return r2.RectFromPoints(topLeftPoint, bottomRightPoint)
}
