package location

import (
	"github.com/NOVAPokemon/utils"
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

func GetTileBoundsFromTileNr(tileNr, lineSize int, tileSideLength float64) (topLeft utils.Location,
	botRight utils.Location) {
	topLeft = utils.Location{
		Longitude: (float64(tileNr%lineSize))*tileSideLength - 180.0,
		Latitude:  180 - float64(tileNr)/float64(lineSize)*tileSideLength,
	}

	botRight = utils.Location{
		Latitude:  topLeft.Latitude - tileSideLength,
		Longitude: topLeft.Longitude + tileSideLength,
	}

	return topLeft, botRight
}

func GetTileCenterLocationFromTileNr(tileNr, lineSize int, tileSideLength float64) utils.Location {
	topLeft, _ := GetTileBoundsFromTileNr(tileNr, lineSize, tileSideLength)

	return utils.Location{
		Latitude:  topLeft.Latitude - (tileSideLength / 2),
		Longitude: topLeft.Longitude + (tileSideLength / 2),
	}
}
