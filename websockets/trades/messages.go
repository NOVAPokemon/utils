package trades

import (
	"encoding/json"
	ws "github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrorParsing = ErrorMessage{
		Info:  "error parsing message",
		Fatal: false,
	}.SerializeToWSMessage()
)

// Message Types
const (
	Start = "START"

	Trade  = "TRADE"
	Accept = "ACCEPT"
	Update = "UPDATE"

	SetToken = "SETTOKEN"
	Finish   = "FINISH_TRADE"

	Error = "ERROR"
	None  = "NONE"
)

// Start
type StartMessage struct {
	ws.MessageWithId
}

func (sMsg StartMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(sMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: Start,
		MsgArgs: []string{string(msgJson)},
	}
}

// Trade
type TradeMessage struct {
	ItemId string
	ws.TrackedMessage
}

func NewTradeMessage(itemId string) TradeMessage {
	return TradeMessage{
		ItemId:         itemId,
		TrackedMessage: ws.NewTrackedMessage(primitive.NewObjectID()),
	}
}

func (tMsg TradeMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(tMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: Trade,
		MsgArgs: []string{string(msgJson)},
	}
}

// Accept
type AcceptMessage struct {
	ws.TrackedMessage
}

func NewAcceptMessage() AcceptMessage {
	return AcceptMessage{
		TrackedMessage: ws.NewTrackedMessage(primitive.NewObjectID()),
	}
}

func (aMsg AcceptMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(aMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: Accept,
		MsgArgs: []string{string(msgJson)},
	}
}

// Update
type UpdateMessage struct {
	TradeStarted  bool
	TradeFinished bool
	Players       [2]*PlayerInfo
	ws.TrackedMessage
}

func (uMsg UpdateMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(uMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: Update,
		MsgArgs: []string{string(msgJson)},
	}
}

func UpdateMessageFromTrade(trade *TradeStatus, trackedMsg ws.TrackedMessage) *UpdateMessage {
	players := [2]*PlayerInfo{}
	players[0] = PlayerToPlayerInfo(&trade.Players[0])
	players[1] = PlayerToPlayerInfo(&trade.Players[1])
	return &UpdateMessage{
		TradeStarted:   trade.TradeStarted,
		TradeFinished:  trade.TradeFinished,
		Players:        players,
		TrackedMessage: trackedMsg,
	}
}

// SetToken
type SetTokenMessage struct {
	TokenString string
	ws.MessageWithId
}

func (sMsg SetTokenMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(sMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: SetToken,
		MsgArgs: []string{string(msgJson)},
	}
}

// Finish
type FinishMessage struct {
	Success bool
	ws.MessageWithId
}

func (fMsg FinishMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(fMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: Finish,
		MsgArgs: []string{string(msgJson)},
	}
}

// Error
type ErrorMessage struct {
	Info  string
	Fatal bool
	ws.MessageWithId
}

func (eMsg ErrorMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(eMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: Error,
		MsgArgs: []string{string(msgJson)},
	}
}

type NoneMessage struct{}

func (nMsg NoneMessage) SerializeToWSMessage() *ws.Message {
	return &ws.Message{
		MsgType: None,
		MsgArgs: nil,
	}
}

func DeserializeTradeMessage(msg *ws.Message) ws.Serializable {
	switch msg.MsgType {
	case Start:
		var startMessage StartMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &startMessage)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &startMessage
	case Trade:
		var tradeMsg TradeMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &tradeMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &tradeMsg
	case Accept:
		var acceptMsg AcceptMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &acceptMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &acceptMsg
	case Update:
		var updateMsg UpdateMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &updateMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &updateMsg
	case SetToken:
		var setTokenMsg SetTokenMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &setTokenMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &setTokenMsg
	case Finish:
		var finishMsg FinishMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &finishMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &finishMsg
	case Error:
		var errorMsg ErrorMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &errorMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &errorMsg
	default:
		log.Error("invalid msg type")
		return nil
	}
}
