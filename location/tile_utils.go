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

func BoundaryToRect(boundary utils.Boundary) r2.Rect {
	topLeftPoint := r2.Point{
		X: boundary.TopLeft.Longitude,
		Y: boundary.TopLeft.Latitude,
	}
	bottomRightPoint := r2.Point{
		X: boundary.BottomRight.Longitude,
		Y: boundary.BottomRight.Latitude,
	}

	return r2.RectFromPoints(topLeftPoint, bottomRightPoint)
}