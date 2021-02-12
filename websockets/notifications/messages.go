package notifications

import (
	"github.com/NOVAPokemon/utils"
	ws "github.com/NOVAPokemon/utils/websockets"
)

const (
	Notification = "NOTIFICATION"
)

// Notification
type NotificationMessage struct {
	Notification utils.Notification
	Info         ws.TrackedInfo
}

func (nMsg NotificationMessage) ConvertToWSMessage() *ws.WebsocketMsg {
	return ws.NewReplyMsg(Notification, nMsg, nMsg.Info)
}
