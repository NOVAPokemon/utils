package trades

import (
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/websockets"
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

type UpdateMessage struct {
	message *websockets.Message
	sendTo  int
}

type ItemsMap = map[string]utils.Item

const Everyone = -1
