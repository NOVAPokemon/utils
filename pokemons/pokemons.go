package pokemons

import "go.mongodb.org/mongo-driver/bson/primitive"

type Pokemon struct {
	Id      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Species string
	XP      float64
	Level   int
	HP      int
	MaxHP   int
	Damage  int
}