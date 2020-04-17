package trades

import (
	"encoding/json"
	ws "github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
)

var (
	ErrorParsing = ErrorMessage{
		Info:  "error parsing message",
		Fatal: false,
	}.SerializeToWSMessage()

	NoneMessageConst = NoneMessage{}.SerializeToWSMessage()
)

// Message Types
const (
	START = "START"

	TRADE  = "TRADE"
	ACCEPT = "ACCEPT"

	UPDATE = "UPDATE"

	SETTOKEN = "SETTOKEN"

	FINISH = "FINISH_TRADE"

	// Error
	ERROR = "ERROR"

	NONE = "NONE"
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
		MsgType: START,
		MsgArgs: []string{string(msgJson)},
	}
}

// Trade
type TradeMessage struct {
	ItemId string
	ws.MessageWithId
}

func (tMsg TradeMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(tMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: TRADE,
		MsgArgs: []string{string(msgJson)},
	}
}

// Accept
type AcceptMessage struct {
	ws.MessageWithId
}

func (aMsg AcceptMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(aMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: ACCEPT,
		MsgArgs: []string{string(msgJson)},
	}
}

// Update
type UpdateMessage struct {
	TradeStarted  bool
	TradeFinished bool
	Players       [2]*PlayerInfo
	ws.MessageWithId
}

func (uMsg UpdateMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(uMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: UPDATE,
		MsgArgs: []string{string(msgJson)},
	}
}

func UpdateMessageFromTrade(trade *TradeStatus) *UpdateMessage {
	players := [2]*PlayerInfo{}
	players[0] = PlayerToPlayerInfo(&trade.Players[0])
	players[1] = PlayerToPlayerInfo(&trade.Players[1])
	return &UpdateMessage{
		TradeStarted:  trade.TradeStarted,
		TradeFinished: trade.TradeFinished,
		Players:       players,
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
		MsgType: SETTOKEN,
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
		MsgType: FINISH,
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
		MsgType: ERROR,
		MsgArgs: []string{string(msgJson)},
	}
}

type NoneMessage struct{}

func (nMsg NoneMessage) SerializeToWSMessage() *ws.Message {
	return &ws.Message{
		MsgType: NONE,
		MsgArgs: nil,
	}
}

func Deserialize(msg *ws.Message) ws.Serializable {
	switch msg.MsgType {
	case START:
		var startMessage StartMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &startMessage)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &startMessage
	case TRADE:
		var tradeMsg TradeMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &tradeMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &tradeMsg
	case ACCEPT:
		var acceptMsg AcceptMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &acceptMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &acceptMsg
	case UPDATE:
		var updateMsg UpdateMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &updateMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &updateMsg
	case SETTOKEN:
		var setTokenMsg SetTokenMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &setTokenMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &setTokenMsg
	case FINISH:
		var finishMsg FinishMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &finishMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &finishMsg
	case ERROR:
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
