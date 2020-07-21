package utils

import (
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/pokemons"
	"github.com/golang/geo/s2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	// user info
	Username     string `json:"username" bson:"username,omitempty"`
	PasswordHash []byte
}

type Trainer struct {
	// game info
	Username string `json:"username" bson:"username,omitempty"`
	Pokemons map[string]pokemons.Pokemon
	Items    map[string]items.Item
	Stats    TrainerStats
	Location s2.LatLng
}

type TransactionTemplate struct {
	Name  string
	Coins int
	Price int
}

type TransactionRecord struct {
	Id           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	User         string
	TemplateName string
}

type TrainerStats struct {
	XP    float64
	Level int
	Coins int
}

type Battle struct {
	Trainer1 string
	Trainer2 string
	Winner   string
	// TODO add more stuff? perharps a log of a battle
	// logAddr url.URL
}

type Lobby struct {
	Id       primitive.ObjectID
	Username string
}

type UserJSON struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Notification struct {
	Id       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username string             `json:"username" bson:"username,omitempty"`
	Type     string
	Content  []byte
}

type Gym struct {
	Name        string `json:"name" bson:"name,omitempty"`
	Location    s2.LatLng
	RaidForming bool
	RaidBoss    *pokemons.Pokemon
}

type WildPokemonWithServer struct {
	Pokemon  pokemons.Pokemon
	Location s2.LatLng
	Server   string
}

type LocationServerCells struct {
	CellIdsStrings []string `json:"cells"`
	ServerName     string
}

type GymWithServer struct {
	Gym        Gym    `json:"gym" bson:"gym,omitempty"`
	ServerName string `json:"servername" bson:"servername,omitempty"`
}

type ClientDelays struct {
	Default     float64            `json:"default"`
	Multipliers map[string]float64 `json:"multipliers"`
}

