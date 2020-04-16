package notifications

import (
	"encoding/json"
	"github.com/NOVAPokemon/utils"
	ws "github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/messages"
	tradeMessages "github.com/NOVAPokemon/utils/websockets/messages/trades"
	log "github.com/sirupsen/logrus"
)

const (
	NOTIFICATION = "NOTIFICATION"
)

// Notification
type NotificationMessage struct {
	Notification utils.Notification
	messages.MessageWithId
}

func (nMsg NotificationMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(nMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: NOTIFICATION,
		MsgArgs: []string{string(msgJson)},
	}
}

func Deserialize(msg *ws.Message) messages.Serializable {
	switch msg.MsgType {
	case tradeMessages.START:
		var notificationMsg NotificationMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &notificationMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &notificationMsg
	default:
		return nil
	}
}
