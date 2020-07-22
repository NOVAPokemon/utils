package api

import (
	"fmt"
)

const UserLocationPath = "/location/user_location"
const GymLocationPath = "/location/gym_location"

const GetAllConfigsPath = "/location/configs"
const SetServerConfigPath = "/location/configs/%s"
const ForceLoadConfigPath = "/location/config/reload"
const GetServerForLocationPath = "/location/server"
const GetActiveCells = "/location/active/cells/%s"
const GetActivePokemons = "/location/active/pokemons/%s"

const ServerNamePathVar = "serverName"
const LatitudeQueryParam = "latitude"
const LongitudeQueryParam = "longitude"

var GetAllConfigsRoute = GetAllConfigsPath
var SetServerConfigRoute = fmt.Sprintf(SetServerConfigPath, fmt.Sprintf("{%s}", ServerNamePathVar))
var GetServerForLocationRoute = GetServerForLocationPath

var GetActivePokemonsRoute = fmt.Sprintf(GetActivePokemons, fmt.Sprintf("{%s}", ServerNamePathVar))
var GetActiveCellsRoute = fmt.Sprintf(GetActiveCells, fmt.Sprintf("{%s}", ServerNamePathVar))

var ForceLoadConfigRoute = ForceLoadConfigPath

var UserLocationRoute = UserLocationPath
var GymLocationRoute = GymLocationPath

// var CatchWildPokemonRoute = CatchWildPokemonPath
