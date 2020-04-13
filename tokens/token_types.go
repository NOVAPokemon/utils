package tokens

import (
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/pokemons"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Types
type AuthToken struct {
	Id       primitive.ObjectID
	Username string
	jwt.StandardClaims
}

type TrainerStatsToken struct {
	TrainerStats utils.TrainerStats
	TrainerHash  []byte
	jwt.StandardClaims
}

type ItemsToken struct {
	Items     map[string]items.Item
	ItemsHash []byte
	jwt.StandardClaims
}

type PokemonToken struct {
	Pokemon     pokemons.Pokemon
	PokemonHash []byte
	jwt.StandardClaims
}
