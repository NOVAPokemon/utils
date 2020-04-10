package trades

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
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
)

type TradeStatus struct {
	Players       [2]Player
	TradeStarted  bool
	TradeFinished bool
}

type Player struct {
	Items    []*utils.Item
	Accepted bool
}

type PlayerInfo struct {
	Items    []string
	Accepted bool
}

func PlayerToPlayerInfo(player *Player) *PlayerInfo{
	items := make([]string, len(player.Items))

	for i, item := range player.Items {
		items[i] = item.Id.Hex()
	}

	return &PlayerInfo{
		Items:    items,
		Accepted: player.Accepted,
	}
}

type ItemsMap = map[string]utils.Item

func CreateTradeMsg(itemId string) string {
	return fmt.Sprintf("%s %s", TRADE, itemId)
}

func CreateAcceptMsg() string {
	return fmt.Sprintf("%s", ACCEPT)
}
