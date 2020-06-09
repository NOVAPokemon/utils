package utils

import (
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/pokemons"
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
	Location Location
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

type Location struct {
	Latitude  float64
	Longitude float64
}

type UserLocation struct {
	Username string `json:"username" bson:"username,omitempty"`
	Location Location
}

type Gym struct {
	Name        string `json:"name" bson:"name,omitempty"`
	Location    Location
	RaidForming bool
	RaidBoss    *pokemons.Pokemon
}

type WildPokemon struct {
	Pokemon  pokemons.Pokemon
	Location Location
}

type LocationServerBoundary struct {
	TopLeftCorner  Location
	BotRightCorner Location
	ServerName     string
}

type GymWithServer struct {
	Gym        Gym    `json:"gym" bson:"gym,omitempty"`
	ServerName string `json:"servername" bson:"servername,omitempty"`
}
