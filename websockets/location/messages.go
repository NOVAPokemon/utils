package location

import (
	"encoding/json"
	"github.com/NOVAPokemon/utils"
	ws "github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
)

const (
	UpdateLocation = "UPDATE_LOCATION"
	Gyms           = "GYMS"
)

// Location
type UpdateLocationMessage struct {
	Location utils.Location
	ws.MessageWithId
}

func (ulMsg UpdateLocationMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(ulMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: UpdateLocation,
		MsgArgs: []string{string(msgJson)},
	}
}

type GymsMessage struct {
	Gyms []utils.Gym
	ws.MessageWithId
}

func (gymMsg GymsMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(gymMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: Gyms,
		MsgArgs: []string{string(msgJson)},
	}
}

func Deserialize(msg *ws.Message) ws.Serializable {
	switch msg.MsgType {
	case UpdateLocation:
		var locationMsg UpdateLocationMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &locationMsg)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &locationMsg
	case Gyms:
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
