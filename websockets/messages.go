package websockets

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	jsonbytes, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return string(jsonbytes)
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
		log.Error(ErrorTooEarlyToLogEmit)
		return
	}

	log.Infof("[EMIT] %s %s %d", msgType, msg.Id.Hex(), msg.TimeEmitted)
}

func (msg *TrackedMessage) LogReceive(msgType string) {
	if msg.TimeReceived == 0 {
		log.Error(ErrorTooEarlyToLogReceive)
	}

	log.Infof("[RECEIVE] %s %s %d", msgType, msg.Id.Hex(), msg.TimeReceived)
}

func (msg *TrackedMessage) TimeTook() (int64, bool) {
	if msg.TimeEmitted == 0 {
		log.Error(ErrorMsgWasNotEmmitted)
		return 0, false
	}

	if msg.TimeReceived == 0 {
		log.Error(ErrorMsgWasNotEmmitted)
		return 0, false
	}

	return msg.TimeReceived - msg.TimeEmitted, true
}

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

// Basic messages
const (
	Tagged   = "TAGGED"
	Start    = "START"
	Reject   = "REJECT"
	SetToken = "SETTOKEN"
	Finish   = "FINISH"
	Error    = "ERROR"
)

type TaggedMessage struct {
	LocationTag string
	IsClient    bool
	MsgBytes    []byte
}

func (tMsg TaggedMessage) SerializeToWSMessage() *Message {
	msgJson, err := json.Marshal(tMsg)
	if err != nil {
		log.Error(WrapSerializeToWSMessageError(err, Tagged))
		return nil
	}

	return &Message{
		MsgType: Tagged,
		MsgArgs: []string{string(msgJson)},
	}
}

func DeserializeTaggedMessage(msgBytes []byte) (*TaggedMessage, error) {
	var taggedMessage TaggedMessage

	err := json.Unmarshal(msgBytes, &taggedMessage)
	if err != nil {
		return nil, wrapDeserializeMsgError(err, Tagged)
	}

	return &taggedMessage, nil
}

type StartMessage struct {
	MessageWithId
}

func (sMsg StartMessage) SerializeToWSMessage() *Message {
	msgJson, err := json.Marshal(sMsg)
	if err != nil {
		log.Error(WrapSerializeToWSMessageError(err, Start))
		log.Error(err)
		return nil
	}

	return &Message{
		MsgType: Start,
		MsgArgs: []string{string(msgJson)},
	}
}

type RejectMessage struct {
	MessageWithId
}

func (rMsg RejectMessage) SerializeToWSMessage() *Message {
	msgJson, err := json.Marshal(rMsg)
	if err != nil {
		log.Error(WrapSerializeToWSMessageError(err, Reject))
		log.Error(err)
		return nil
	}

	return &Message{
		MsgType: Reject,
		MsgArgs: []string{string(msgJson)},
	}
}

type FinishMessage struct {
	Success bool
	MessageWithId
}

func (fMsg FinishMessage) SerializeToWSMessage() *Message {
	msgJson, err := json.Marshal(fMsg)
	if err != nil {
		log.Error(WrapSerializeToWSMessageError(err, Finish))
		return nil
	}

	return &Message{
		MsgType: Finish,
		MsgArgs: []string{string(msgJson)},
	}
}

type SetTokenMessage struct {
	TokenField   string
	TokensString []string
	MessageWithId
}

func (sMsg SetTokenMessage) SerializeToWSMessage() *Message {
	msgJson, err := json.Marshal(sMsg)
	if err != nil {
		log.Error(WrapSerializeToWSMessageError(err, SetToken))
		return nil
	}

	return &Message{
		MsgType: SetToken,
		MsgArgs: []string{string(msgJson)},
	}
}

type ErrorMessage struct {
	Info  string
	Fatal bool
	MessageWithId
}

func (eMsg ErrorMessage) SerializeToWSMessage() *Message {

	msgJson, err := json.Marshal(eMsg)
	if err != nil {
		log.Error(WrapSerializeToWSMessageError(err, Error))
		return nil
	}

	return &Message{
		MsgType: Error,
		MsgArgs: []string{string(msgJson)},
	}
}

func DeserializeMsg(msg *Message) (Serializable, error) {
	switch msg.MsgType {
	case Start:
		var startMsg StartMessage

		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &startMsg)
		if err != nil {
			return nil, wrapDeserializeMsgError(err, Start)
		}

		return &startMsg, nil
	case Reject:
		var rejectMessage RejectMessage

		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &rejectMessage)
		if err != nil {
			return nil, wrapDeserializeMsgError(err, Reject)
		}

		return &rejectMessage, nil
	case Finish:
		var finishMsg FinishMessage

		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &finishMsg)
		if err != nil {
			return nil, wrapDeserializeMsgError(err, Finish)
		}

		return &finishMsg, nil
	case Error:
		var errMsg ErrorMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &errMsg)
		if err != nil {
			return nil, wrapDeserializeMsgError(err, Error)
		}
		return &errMsg, nil
	case SetToken:
		var setTokenMessage SetTokenMessage

		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &setTokenMessage)
		if err != nil {
			return nil, wrapDeserializeMsgError(err, SetToken)
		}

		return &setTokenMessage, nil
	default:
		return nil, wrapDeserializeMsgError(ErrorInvalidMessageType, msg.MsgType)
	}
}

type GenericMsg struct {
	MsgType int
	Data    []byte
}
