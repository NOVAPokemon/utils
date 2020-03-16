package utils

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	// user info
	Id           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TrainerId    primitive.ObjectID
	Username     string
	PasswordHash [] byte
	Salt         [] byte // optional

}

type Trainer struct {
	// game info
	Bag      primitive.ObjectID
	Pokemons [] primitive.ObjectID
	Level    int
	Coins    int
}

type Bag struct {
	Id    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Items [] primitive.ObjectID
}

type Item struct {
	Id    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Price int
}

type Pokemon struct {
	Id      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Owner   primitive.ObjectID
	Species string
	Level   int
	HP      int
	Damage  int
}