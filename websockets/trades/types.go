package trades

import (
	"github.com/NOVAPokemon/utils/items"
)

type TradeStatus struct {
	Players       [2]Player
	TradeStarted  bool
	TradeFinished bool
}

type Player struct {
	Items    []items.Item
	Accepted bool
}

type PlayerInfo struct {
	Items    []string
	Accepted bool
}

func PlayerToPlayerInfo(player *Player) *PlayerInfo {
	playerItems := make([]string, len(player.Items))

	for i, item := range player.Items {
		playerItems[i] = item.Id.Hex()
	}

	return &PlayerInfo{
		Items:    playerItems,
		Accepted: player.Accepted,
	}
}

type ItemsMap = map[string]items.Item
