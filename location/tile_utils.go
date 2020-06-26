package location

import (
	"github.com/NOVAPokemon/utils"
	"github.com/golang/geo/r2"
)

func IsWithinBounds(location utils.Location, topLeft utils.Location, botRight utils.Location) bool {
	if location.Longitude >= botRight.Longitude || location.Longitude <= topLeft.Longitude {
		return false
	}
	if location.Latitude <= botRight.Latitude || location.Latitude >= topLeft.Latitude {
		return false
	}
	return true
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
