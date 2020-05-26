package api

import "fmt"

const TradeIdVar = "tradeId"

const GetTradesPath = "/trades"
const StartTradePath = "/trades/join"
const JoinTradePath = "/trades/join/%s"
const RejectTradePath = "/trades/reject/%s"

var (
	JoinTradeRoute   = fmt.Sprintf(JoinTradePath, fmt.Sprintf("{%s}", TradeIdVar))
	RejectTradeRoute = fmt.Sprintf(RejectTradePath, fmt.Sprintf("{%s}", TradeIdVar))
)

type CreateLobbyRequest struct {
	Username string `json:"username"`
}

type CreateLobbyResponse struct {
	LobbyId    string `json:"lobbyId"`
	ServerName string `json:"serverName"`
}
