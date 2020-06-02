package notifications

const ChallengeToBattle = "Challenge"
const WantsToTrade = "WantingTrade"

type WantsToTradeContent struct {
	Username       string
	LobbyId        string
	ServerHostname string
}

type WantsToBattleContent struct {
	Username       string
	LobbyId        string
	ServerHostname string
}
