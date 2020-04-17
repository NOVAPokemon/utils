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
	Id               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username         string             `json:"username" bson:"username,omitempty"`
	Type             string
	Content          []byte
	TimestampEmitted int64
}

type Location struct {
	Latitude  float64
	Longitude float64
}

type UserLocation struct {
	Username string `json:"username" bson:"username,omitempty"`
	Location Location
}

// Location parameters describe how the user will move. The moving probability indicates
// how likely is the user to move between location updates
// MaxMovingSpeed should be in meters per second
type LocationParameters struct {
	StartingLocation     Location
	MaxMovingSpeed       float64
	MovingProbability    float64
	MaxDistanceFromStart int
}

type Gym struct {
	Name        string
	Location    Location
	RaidForming bool
	RaidBoss    *pokemons.Pokemon
}
