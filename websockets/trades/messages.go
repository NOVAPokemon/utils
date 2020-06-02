package trades

import (
	"encoding/json"

	ws "github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrorParsing = ws.ErrorMessage{
		Info:  "error parsing message",
		Fatal: false,
	}.SerializeToWSMessage()
)

// Message Types
const (
	Trade  = "TRADE"
	Accept = "ACCEPT"
	Update = "UPDATE"
)

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

func DeserializeTradeMessage(msg *ws.Message) (ws.Serializable, error) {
	switch msg.MsgType {
	case Trade:
		var tradeMsg TradeMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &tradeMsg)
		if err != nil {
			return nil, wrapDeserializeTradeMsgError(err, Trade)
		}

		return &tradeMsg, nil
	case Accept:
		var acceptMsg AcceptMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &acceptMsg)
		if err != nil {
			log.Error(err)
			return nil, wrapDeserializeTradeMsgError(err, Trade)
		}

		return &acceptMsg, nil
	case Update:
		var updateMsg UpdateMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &updateMsg)
		if err != nil {
			log.Error(err)
			return nil, wrapDeserializeTradeMsgError(err, Update)
		}

		return &updateMsg, nil
	default:
		deserializedMsg, err := ws.DeserializeMsg(msg)
		if err != nil {
			return nil, wrapDeserializeTradeMsgError(err, msg.MsgType)
		}

		return deserializedMsg, nil
	}
}
