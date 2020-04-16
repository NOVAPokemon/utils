package trades

import (
	"encoding/json"
	"github.com/NOVAPokemon/utils/messages"
	ws "github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/trades"
	log "github.com/sirupsen/logrus"
)

type Serializable interface {
	SerializeToWSMessage() *ws.Message
}

// Start
type StartMessage struct {
	messages.MessageWithId
}

func (sMsg StartMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(sMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: trades.START,
		MsgArgs: []string{string(msgJson)},
	}
}

// Trade
type TradeMessage struct {
	ItemId string
	messages.MessageWithId
}

func (tMsg TradeMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(tMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: trades.TRADE,
		MsgArgs: []string{string(msgJson)},
	}
}

// Accept
type AcceptMessage struct {
	messages.MessageWithId
}

func (aMsg AcceptMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(aMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: trades.ACCEPT,
		MsgArgs: []string{string(msgJson)},
	}
}

// Update
type UpdateMessage struct {
	TradeStarted  bool
	TradeFinished bool
	Players       [2]*trades.PlayerInfo
	messages.MessageWithId
}

func (uMsg UpdateMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(uMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: trades.UPDATE,
		MsgArgs: []string{string(msgJson)},
	}
}

func UpdateMessageFromTrade(trade *trades.TradeStatus) *UpdateMessage {
	players := [2]*trades.PlayerInfo{}
	players[0] = trades.PlayerToPlayerInfo(&trade.Players[0])
	players[1] = trades.PlayerToPlayerInfo(&trade.Players[1])
	return &UpdateMessage{
		TradeStarted:  trade.TradeStarted,
		TradeFinished: trade.TradeFinished,
		Players:       players,
	}
}

// SetToken
type SetTokenMessage struct {
	TokenString string
	messages.MessageWithId
}

func (sMsg SetTokenMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(sMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: trades.SETTOKEN,
		MsgArgs: []string{string(msgJson)},
	}
}

// Finish
type FinishMessage struct {
	Success bool
	messages.MessageWithId
}

func (fMsg FinishMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(fMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: trades.FINISH,
		MsgArgs: []string{string(msgJson)},
	}
}

// Error
type ErrorMessage struct {
	Info  string
	Fatal bool
	messages.MessageWithId
}

func (eMsg ErrorMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(eMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: trades.ERROR,
		MsgArgs: []string{string(msgJson)},
	}
}

type NoneMessage struct{}

func (nMsg NoneMessage) SerializeToWSMessage() *ws.Message {
	return &ws.Message{
		MsgType: trades.NONE,
		MsgArgs: nil,
	}
}

func Deserialize(msg *ws.Message) interface{} {
	switch msg.MsgType {
	case trades.START:
		var startMessage StartMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &startMessage)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &startMessage
	case trades.TRADE:
		var tradeMsg TradeMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &tradeMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &tradeMsg
	case trades.ACCEPT:
		var acceptMsg AcceptMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &acceptMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &acceptMsg
	case trades.UPDATE:
		var updateMsg UpdateMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &updateMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &updateMsg
	case trades.SETTOKEN:
		var setTokenMsg SetTokenMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &setTokenMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &setTokenMsg
	case trades.FINISH:
		var finishMsg FinishMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &finishMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &finishMsg
	case trades.ERROR:
		var errorMsg ErrorMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &errorMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &errorMsg
	default:
		return nil
	}
}
