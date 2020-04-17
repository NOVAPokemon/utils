package notifications

import (
	"encoding/json"
	"github.com/NOVAPokemon/utils"
	ws "github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
)

const (
	Notification = "NOTIFICATION"
)

// Notification
type NotificationMessage struct {
	Notification utils.Notification
	ws.MessageWithId
}

func (nMsg NotificationMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(nMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: Notification,
		MsgArgs: []string{string(msgJson)},
	}
}

func Deserialize(msg *ws.Message) ws.Serializable {
	switch msg.MsgType {
	case Notification:
		var notificationMsg NotificationMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &notificationMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &notificationMsg
	default:
		log.Info("invalid msg type")
		return nil
	}
}
