package location

import (
	"encoding/json"
	"github.com/NOVAPokemon/utils"
	ws "github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/location"
	"github.com/NOVAPokemon/utils/websockets/messages"
	log "github.com/sirupsen/logrus"
)

// Location
type UpdateLocationMessage struct {
	Location utils.Location
	messages.MessageWithId
}

func (ulMsg UpdateLocationMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(ulMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: location.UpdateLocation,
		MsgArgs: []string{string(msgJson)},
	}
}

type GymsMessage struct {
	Gyms []utils.Gym
	messages.MessageWithId
}

func (gymMsg GymsMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(gymMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: location.Gyms,
		MsgArgs: []string{string(msgJson)},
	}
}

func Deserialize(msg *ws.Message) messages.Serializable {
	switch msg.MsgType {
	case location.UpdateLocation:
		var locationMsg UpdateLocationMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &locationMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &locationMsg
	case location.Gyms:
		var gymsMsg GymsMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &gymsMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &gymsMsg
	default:
		return nil
	}
}
