package websockets

import (
	"time"
)

const (
	TimeoutVal       = 5
	Timeout          = TimeoutVal * time.Second
	WebsocketTimeout = 30 * time.Second
)
