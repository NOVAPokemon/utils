package trades

import (
	ws "github.com/NOVAPokemon/utils/websockets"
)

// Message Types
const (
	// HTTP requests
	CreateTrade = "CREATE_TRADE"

	Trade  = "TRADE"
	Accept = "ACCEPT"
	Update = "UPDATE"
)

// Trade
type TradeMessage struct {
	ItemId string
}

func (tMsg TradeMessage) ConvertToWSMessage() *ws.WebsocketMsg {
	return ws.NewRequestMsg(Trade, tMsg)
}

// Accept
type AcceptMessage struct{}

func (aMsg AcceptMessage) ConvertToWSMessage() *ws.WebsocketMsg {
	return ws.NewRequestMsg(Accept, aMsg)
}

// Update
type UpdateMessage struct {
	Players [2]*PlayerInfo
}

func (uMsg UpdateMessage) ConvertToWSMessage(info ws.TrackedInfo) *ws.WebsocketMsg {
	return ws.NewReplyMsg(Update, uMsg, info)
}

func UpdateMessageFromTrade(trade *TradeStatus) *UpdateMessage {
	players := [2]*PlayerInfo{}
	players[0] = PlayerToPlayerInfo(&trade.Players[0])
	players[1] = PlayerToPlayerInfo(&trade.Players[1])
	return &UpdateMessage{
		Players: players,
	}
}
