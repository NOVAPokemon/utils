package tokens

import (
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/pokemons"
	"github.com/dgrijalva/jwt-go"
)

// Types
type AuthToken struct {
	Id       string
	Username string
	jwt.StandardClaims
}

type TrainerStatsToken struct {
	TrainerStats utils.TrainerStats
	TrainerHash  string
	jwt.StandardClaims
}

type ItemsToken struct {
	Items     map[string]items.Item
	ItemsHash string
	jwt.StandardClaims
}

type PokemonToken struct {
	Pokemon     pokemons.Pokemon
	PokemonHash string
	jwt.StandardClaims
}
