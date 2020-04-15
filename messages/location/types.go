package location

import (
	"encoding/json"
	"github.com/NOVAPokemon/utils"
	ws "github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/location"
	log "github.com/sirupsen/logrus"
)

// Location
type UpdateLocationMessage struct {
	Location utils.Location
}

func (ulMsg UpdateLocationMessage) Serialize() *ws.Message {
	jsonLocation, err := json.Marshal(ulMsg.Location)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: location.UpdateLocation,
		MsgArgs: []string{string(jsonLocation)},
	}
}

type GymsMessage struct {
	Gyms []utils.Gym
}

func (gymMsg GymsMessage) Serialize() *ws.Message {
	jsonGyms, err := json.Marshal(gymMsg.Gyms)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: location.Gyms,
		MsgArgs: []string{string(jsonGyms)},
	}
}

func Deserialize(msg *ws.Message) interface{} {
	switch msg.MsgType {
	case location.UpdateLocation:
		var location utils.Location
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &location)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &UpdateLocationMessage{Location: location}
	case location.Gyms:
		var gyms []utils.Gym
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &gyms)
		if err != nil {
			log.Error(err)
			return nil
		}

		return &GymsMessage{Gyms: gyms}
	default:
		return nil
	}
}
