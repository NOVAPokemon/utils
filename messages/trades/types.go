package trades

import (
	ws "github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/trades"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type Serializable interface {
	SerializeToWSMessage() *ws.Message
}

// Start
type StartMessage struct{}

func (sMsg StartMessage) SerializeToWSMessage() *ws.Message {
	return &ws.Message{
		MsgType: trades.START,
		MsgArgs: nil,
	}
}

// Trade
type TradeMessage struct {
	Items string
}

func (tMsg TradeMessage) SerializeToWSMessage() *ws.Message {
	return &ws.Message{
		MsgType: trades.TRADE,
		MsgArgs: []string{tMsg.Items},
	}
}

// Accept
type AcceptMessage struct{}

func (aMsg AcceptMessage) SerializeToWSMessage() *ws.Message {
	return &ws.Message{
		MsgType: trades.ACCEPT,
		MsgArgs: nil,
	}
}

// Update
type UpdateMessage struct {
	TradeStarted  bool
	TradeFinished bool
	Players       [2]*trades.PlayerInfo
}

func (uMsg UpdateMessage) SerializeToWSMessage() *ws.Message {
	msgArgs := []string{
		strconv.FormatBool(uMsg.TradeStarted),
		strconv.FormatBool(uMsg.TradeFinished),
		strconv.FormatBool(uMsg.Players[0].Accepted),
		strconv.Itoa(len(uMsg.Players[0].Items)),
	}

	msgArgs = append(msgArgs, uMsg.Players[0].Items...)

	msgArgs = append(msgArgs, strconv.FormatBool(uMsg.Players[1].Accepted), strconv.Itoa(len(uMsg.Players[1].Items)))
	msgArgs = append(msgArgs, uMsg.Players[1].Items...)

	return &ws.Message{
		MsgType: trades.UPDATE,
		MsgArgs: msgArgs,
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
}

func (sMsg SetTokenMessage) SerializeToWSMessage() *ws.Message {
	return &ws.Message{
		MsgType: trades.SETTOKEN,
		MsgArgs: []string{sMsg.TokenString},
	}
}

// Finish
type FinishMessage struct {
	Success bool
}

func (fMsg FinishMessage) SerializeToWSMessage() *ws.Message {
	return &ws.Message{
		MsgType: trades.FINISH,
		MsgArgs: []string{strconv.FormatBool(fMsg.Success)},
	}
}

// Error
type ErrorMessage struct {
	Info  string
	Fatal bool
}

func (eMsg ErrorMessage) SerializeToWSMessage() *ws.Message {
	return &ws.Message{
		MsgType: trades.ERROR,
		MsgArgs: []string{eMsg.Info, strconv.FormatBool(eMsg.Fatal)},
	}
}

type NoneMessage struct {}

func (nMsg NoneMessage) SerializeToWSMessage() *ws.Message {
	return &ws.Message{
		MsgType: trades.NONE,
		MsgArgs: nil,
	}
}

func Deserialize(msg *ws.Message) interface{} {
	switch msg.MsgType {
	case trades.START:
		return &StartMessage{}
	case trades.TRADE:
		return &TradeMessage{Items: msg.MsgArgs[0]}
	case trades.ACCEPT:
		return &AcceptMessage{}
	case trades.UPDATE:
		index := 0
		started, err := strconv.ParseBool(msg.MsgArgs[index])
		if err != nil {
			handleDeserializingError(err)
			return nil
		}
		index++

		finished, err := strconv.ParseBool(msg.MsgArgs[index])
		if err != nil {
			handleDeserializingError(err)
			return nil
		}
		index++

		finishedAt, player1 := parsePlayer(msg.MsgArgs, index)
		if player1 == nil {
			return nil
		}

		_, player2 := parsePlayer(msg.MsgArgs, finishedAt)
		if player2 == nil {
			return nil
		}

		return &UpdateMessage{
			TradeStarted:  started,
			TradeFinished: finished,
			Players:       [2]*trades.PlayerInfo{player1, player2},
		}
	case trades.SETTOKEN:
		return &SetTokenMessage{TokenString: msg.MsgArgs[0]}
	case trades.FINISH:
		value, err := strconv.ParseBool(msg.MsgArgs[0])
		if err != nil {
			handleDeserializingError(err)
			return nil
		}

		return &FinishMessage{Success: value}
	case trades.ERROR:
		info := msg.MsgArgs[0]

		fatal, err := strconv.ParseBool(msg.MsgArgs[1])
		if err != nil {
			handleDeserializingError(err)
			return nil
		}

		return &ErrorMessage{
			Info:  info,
			Fatal: fatal,
		}
	default:
		return nil
	}
}

func parsePlayer(strings []string, startAt int) (finishedAt int, player *trades.PlayerInfo) {
	index := startAt
	accepted, err := strconv.ParseBool(strings[index])
	if err != nil {
		handleDeserializingError(err)
		return 0, nil
	}
	index++

	numItems, err := strconv.Atoi(strings[index])
	if err != nil {
		handleDeserializingError(err)
		return 0, nil
	}
	index++

	items := make([]string, numItems)

	for i := 0; i < numItems; i++ {
		items[i] = strings[i+index]
	}

	return index + len(items), &trades.PlayerInfo{
		Items:    items,
		Accepted: accepted,
	}
}

func handleDeserializingError(err error) {
	log.Error("While deserializing got error: ", err)
}
