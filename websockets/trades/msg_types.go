package trades

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/websockets"
)

// Message Types
const (
	START = "START"

	TRADE  = "TRADE"
	ACCEPT = "ACCEPT"

	UPDATE = "UPDATE"

	SET_TOKEN = "SET_TOKEN"

	FINISH = "FINISH_TRADE"

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

func CreateTradeMsg(itemId string) string {
	return fmt.Sprintf("%s %s", TRADE, itemId)
}

func CreateAcceptMsg() string {
	return fmt.Sprintf("%s", ACCEPT)
}
