package api

import (
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/pokemons"
)

const UserLocationPath = "/user_location"
const GymLocationPath = "/gym_location"
const CatchWildPokemonPath = "/catch"
const SetRegionAreaPath = "/region"

var SetRegionAreaRoute = SetRegionAreaPath
var UserLocationRoute = UserLocationPath
var GymLocationRoute = GymLocationPath
var CatchWildPokemonRoute = CatchWildPokemonPath

type CatchWildPokemonRequest struct {
	Pokeball items.Item
	Pokemon  pokemons.Pokemon
}
