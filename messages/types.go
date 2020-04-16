package messages

import "go.mongodb.org/mongo-driver/bson/primitive"

type MessageWithId struct {
	Id primitive.ObjectID
}
