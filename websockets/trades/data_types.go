package trades

import (
	"github.com/NOVAPokemon/utils"
)

// Message Types
const (
	TRADE  = "TRADE"
	ACCEPT = "ACCEPT"

	UPDATE = "UPDATE"

	FINISH = "FINISH"

	// Error
	ERROR = "ERROR"
)

type TradeStatus struct {
	Players       [2]Players
	TradeStarted  bool
	TradeFinished bool
}

type Players struct {
	Items    []*utils.Item
	Accepted bool
}

type TradeMessage struct {
	MsgType string
	MsgArgs []string
}

type ItemsMap = map[string]utils.Item

type UpdateMessage struct {
	message *TradeMessage
	sendTo  int
}

const Everyone = -1