package api

import "fmt"

const TradeIdVar = "tradeId"

const GetTradesPath = "/trades"
const StartTradePath = "/trades/join"
const JoinTradePath = "/trades/join/%s"

var JoinTradeRoute = fmt.Sprintf(JoinTradePath, fmt.Sprintf("{%s}", TradeIdVar))

type CreateLobbyRequest struct {
	Username string `json:"username"`
}
