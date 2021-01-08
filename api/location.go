package api

import (
	"fmt"
)

const (
	UserLocationPath         = "/location/user_location"
	GymLocationPath          = "/location/gym_location"
	GetAllConfigsPath        = "/location/configs"
	SetServerConfigPath      = "/location/configs/%s"
	ForceLoadConfigPath      = "/location/config/reload"
	GetServerForLocationPath = "/location/server"
	GetActiveCells           = "/location/active/cells/%s"
	GetActivePokemons        = "/location/active/pokemons/%s"
	BeingRemoved             = "/location/removing_instance"

	ServerNamePathVar   = "serverName"
	LatitudeQueryParam  = "latitude"
	LongitudeQueryParam = "longitude"
)

var (
	GetAllConfigsRoute        = GetAllConfigsPath
	SetServerConfigRoute      = fmt.Sprintf(SetServerConfigPath, fmt.Sprintf("{%s}", ServerNamePathVar))
	GetServerForLocationRoute = GetServerForLocationPath

	GetActivePokemonsRoute = fmt.Sprintf(GetActivePokemons, fmt.Sprintf("{%s}", ServerNamePathVar))
	GetActiveCellsRoute    = fmt.Sprintf(GetActiveCells, fmt.Sprintf("{%s}", ServerNamePathVar))

	ForceLoadConfigRoute = ForceLoadConfigPath

	UserLocationRoute = UserLocationPath
	GymLocationRoute  = GymLocationPath

	BeingRemovedRoute = BeingRemoved
)

// var CatchWildPokemonRoute = CatchWildPokemonPath
