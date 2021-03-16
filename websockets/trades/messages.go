package trades

import (
	ws "github.com/NOVAPokemon/utils/websockets"
)

// Message Types
const (
	// HTTP requests
	CreateTrade = "CREATE_TRADE"

	JoinTrade   = "JOIN_TRADE"
	StartTrade  = "START_TRADE"
	RejectTrade = "REJECT_TRADE"
	ErrorTrade  = "ERROR_TRADE"
	Trade       = "TRADE"
	Accept      = "ACCEPT"
	Update      = "UPDATE_TRADE"
)

type StartTradeMessage struct{}

func (s StartTradeMessage) ConvertToWSMessage(info ws.TrackedInfo) *ws.WebsocketMsg {
	return ws.NewReplyMsg(StartTrade, nil, info)
}

type RejectTradeMessage struct{}

func (s RejectTradeMessage) ConvertToWSMessage(info ws.TrackedInfo) *ws.WebsocketMsg {
	return ws.NewReplyMsg(RejectTrade, nil, info)
}

type ErrorTradeMessage struct {
	Info  string
	Fatal bool
}

func (e ErrorTradeMessage) ConvertToWSMessage(info ws.TrackedInfo) *ws.WebsocketMsg {
	return ws.NewReplyMsg(ErrorTrade, e, info)
}

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
