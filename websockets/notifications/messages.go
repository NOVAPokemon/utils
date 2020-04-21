package notifications

import (
	"encoding/json"
	"github.com/NOVAPokemon/utils"
	ws "github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	Notification = "NOTIFICATION"
)

// Notification
type NotificationMessage struct {
	Notification utils.Notification
	ws.TrackedMessage
}

func NewNotificationMessage(notification utils.Notification) NotificationMessage {
	return NotificationMessage{
		Notification:   notification,
		TrackedMessage: ws.NewTrackedMessage(primitive.NewObjectID()),
	}
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

func DeserializeNotificationMessage(msg *ws.Message) ws.Serializable {
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
