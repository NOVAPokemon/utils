package websockets

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
)

type Serializable interface {
	GetId() primitive.ObjectID
	SerializeToWSMessage() *Message
}

type Message struct {
	MsgType string
	MsgArgs []string
}

func (msg Message) Serialize() string {
	builder := strings.Builder{}

	builder.WriteString(msg.MsgType)
	builder.WriteString(" ")

	for _, arg := range msg.MsgArgs {
		builder.WriteString(arg)
		builder.WriteString(" ")
	}

	return builder.String()
}

type MessageWithId struct {
	Id primitive.ObjectID
}

func (msgWithId MessageWithId) GetId() primitive.ObjectID {
	return msgWithId.Id
}

type GenericMsg struct {
	MsgType int
	Data    []byte
}
