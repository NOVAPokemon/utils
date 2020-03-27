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

	Level    int
	Coins    int
}

type Item struct {
	Id   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name string
}

type Pokemon struct {
	Id      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Species string
	Level   int
	HP      int
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

type UserJSON struct{
	Username     string `json:"username"`
	Password string `json:"password"`
}
