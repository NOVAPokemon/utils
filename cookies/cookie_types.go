package cookies

import (
	"github.com/NOVAPokemon/utils"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Types
type AuthToken struct {
	Id       primitive.ObjectID
	Username string
	Claims   jwt.StandardClaims
}

type TrainerStatsToken struct {
	TrainerStats utils.TrainerStats
	TrainerHash  []byte
	Claims       jwt.StandardClaims
}

type ItemsToken struct {
	Items     map[string]utils.Item
	ItemsHash []byte
	Claims    jwt.StandardClaims
}

type PokemonsToken struct {
	Pokemons      map[string][]utils.Pokemon
	PokemonHashes map[string][]byte
	Claims        jwt.StandardClaims
}
