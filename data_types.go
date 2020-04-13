package utils

import (
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
	Pokemons map[string]Pokemon
	Items    map[string]Item
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

type Pokemon struct {
	Id      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Species string
	XP      float64
	Level   int
	HP      int
	MaxHP   int
	Damage  int
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
	//TODO there are a bunch of possible fields,
	RegionName string // e.g. /Europe/England/London/Greenwich
	Latitude   float64
	Longitude  float64
}
