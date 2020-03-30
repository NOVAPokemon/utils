package cookies

import (
	"github.com/NOVAPokemon/utils"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Types
type UserToken struct {
	Id       primitive.ObjectID
	Username string
	jwt.StandardClaims
}

type TrainerStatsToken struct {
	TrainerStats utils.TrainerStats
	TrainerHash  []byte
	Claims       jwt.StandardClaims
}

type ItemsToken struct {
	Items     []utils.Item
	ItemsHash []byte
	Claims    jwt.StandardClaims
}

type PokemonsToken struct {
	Pokemons      map[string][]utils.Pokemon
	PokemonHashes map[string][]byte
	Claims        jwt.StandardClaims
}
