package api

import (
	"fmt"
)

const UserLocationPath = "/location/user_location"
const GymLocationPath = "/location/gym_location"
const CatchWildPokemonPath = "/location/catch"

const GetAllConfigsPath = "/location/configs"
const SetServerConfigPath = "/location/configs/%s"
const ForceLoadConfigPath = "/location/config/reload"
const GetServerForLocationPath = "/location/server"

const ServerNamePathVar = "serverName"
const LatitudeQueryParam = "latitude"
const LongitudeQueryParam = "longitude"

var GetAllConfigsRoute = GetAllConfigsPath
var SetServerConfigRoute = fmt.Sprintf(SetServerConfigPath, ServerNamePathVar)
var GetServerForLocationRoute = GetServerForLocationPath

var ForceLoadConfigRoute = ForceLoadConfigPath

var UserLocationRoute = UserLocationPath
var GymLocationRoute = GymLocationPath
//var CatchWildPokemonRoute = CatchWildPokemonPath
