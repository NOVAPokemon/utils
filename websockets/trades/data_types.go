package trades

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
	Players       [2]players
	tradeStarted  bool
	tradeFinished bool
}

type players struct {
	Items    []string
	Accepted bool
}

type TradeMessage struct {
	msgType string
	msgArgs []string
}
