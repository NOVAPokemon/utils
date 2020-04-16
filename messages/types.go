package messages

import (
	ws "github.com/NOVAPokemon/utils/websockets"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Serializable interface {
	GetId() primitive.ObjectID
	SerializeToWSMessage() *ws.Message
}

type MessageWithId struct {
	Id primitive.ObjectID
}

func (msgWithId MessageWithId) GetId() primitive.ObjectID{
	return msgWithId.Id
}
