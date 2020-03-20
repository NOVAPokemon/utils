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
	Bag      primitive.ObjectID
	Pokemons []primitive.ObjectID
	Level    int
	Coins    int
}

type Bag struct {
	Id    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Items []primitive.ObjectID
}

type Item struct {
	Id   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name string
}

type Pokemon struct {
	Id      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Owner   primitive.ObjectID
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
