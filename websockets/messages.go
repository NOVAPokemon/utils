package websockets

import (
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
	"time"
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

func NewMessageWithId(id primitive.ObjectID) MessageWithId {
	return MessageWithId{Id: id}
}

func (msgWithId MessageWithId) GetId() primitive.ObjectID {
	return msgWithId.Id
}

type Trackable interface {
	Emit(emitTimestamp int64)
	Receive(receiveTimestamp int64)
	LogEmit()
	LogReceive()
}

type TrackedMessage struct {
	TimeEmitted  int64
	TimeReceived int64
	MessageWithId
}

func NewTrackedMessage(id primitive.ObjectID) TrackedMessage {
	return TrackedMessage{
		TimeEmitted:   MakeTimestamp(),
		MessageWithId: NewMessageWithId(id),
	}
}

func (msg *TrackedMessage) Emit(emitTimestamp int64) {
	msg.TimeEmitted = emitTimestamp
}

func (msg *TrackedMessage) Receive(receiveTimestamp int64) {
	msg.TimeReceived = receiveTimestamp
}

func (msg *TrackedMessage) LogEmit(msgType string) {
	if msg.TimeEmitted == 0 {
		log.Error("tried logging before setting emitted timestamp")
		return
	}

	log.Infof("[EMIT] %s %s %d", msgType, msg.Id.Hex(), msg.TimeEmitted)
}

func (msg *TrackedMessage) LogReceive(msgType string) {
	if msg.TimeReceived == 0 {
		log.Error("tried logging before setting received timestamp")
	}

	log.Infof("[RECEIVE] %s %s %d", msgType, msg.Id.Hex(), msg.TimeReceived)
}

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

type GenericMsg struct {
	MsgType int
	Data    []byte
}
