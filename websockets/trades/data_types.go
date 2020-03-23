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
	Players       [2]Players
	TradeStarted  bool
	TradeFinished bool
}

type Players struct {
	Items    []string
	Accepted bool
}

type TradeMessage struct {
	MsgType string
	MsgArgs []string
}
