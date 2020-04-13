package items

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	HealName = "potion"
	ReviveName = "revive"
	PokeBallName = "poke-ball"
	MasterBallName = "master-ball"
)

type Item struct {
	Id     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name   string
	Effect Effect
}
