package api

import (
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/pokemons"
)

const UserLocationPath = "/location/user_location"
const GymLocationPath = "/location/gym_location"
const CatchWildPokemonPath = "/location/catch"
const SetRegionAreaPath = "/location/region"

var SetRegionAreaRoute = SetRegionAreaPath
var UserLocationRoute = UserLocationPath
var GymLocationRoute = GymLocationPath
var CatchWildPokemonRoute = CatchWildPokemonPath

type CatchWildPokemonRequest struct {
	Pokeball items.Item
	Pokemon  pokemons.Pokemon
}
