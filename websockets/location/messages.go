package location

import (
	"encoding/json"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/items"
	ws "github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
)

const (
	UpdateLocation       = "UPDATE_LOCATION"
	Gyms                 = "GYMS"
	Pokemon              = "POKEMON"
	CatchPokemon         = "CATCH_POKEMON"
	CatchPokemonResponse = "CATCH_POKEMON_RESPONSE"
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
	Gyms []utils.GymWithServer
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

type PokemonMessage struct {
	Pokemon []utils.WildPokemon
	ws.MessageWithId
}

func (pokemonMsg PokemonMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(pokemonMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: Pokemon,
		MsgArgs: []string{string(msgJson)},
	}
}

type CatchWildPokemonMessage struct {
	Pokeball items.Item
	Pokemon  string
	ws.MessageWithId
}

func (catchPokemonMsg CatchWildPokemonMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(catchPokemonMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: CatchPokemon,
		MsgArgs: []string{string(msgJson)},
	}
}

type CatchWildPokemonMessageResponse struct {
	Caught        bool
	PokemonTokens []string
	ws.MessageWithId
}

func (catchPokemonMsgResp CatchWildPokemonMessageResponse) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(catchPokemonMsgResp)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: CatchPokemonResponse,
		MsgArgs: []string{string(msgJson)},
	}
}

func DeserializeLocationMsg(msg *ws.Message) (ws.Serializable, error) {
	switch msg.MsgType {
	case UpdateLocation:
		var locationMsg UpdateLocationMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &locationMsg)
		if err != nil {
			return nil, wrapDeserializeLocationMsgError(err, UpdateLocation)
		}
		return &locationMsg, nil
	case Gyms:
		var gymsMsg GymsMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &gymsMsg)
		if err != nil {
			return nil, wrapDeserializeLocationMsgError(err, Gyms)
		}
		return &gymsMsg, nil

	case Pokemon:
		var pokemonMsg PokemonMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &pokemonMsg)
		if err != nil {
			log.Error(err)
			return nil, wrapDeserializeLocationMsgError(err, Pokemon)
		}
		return &pokemonMsg, nil

	case CatchPokemon:
		var catchPokemonMsg CatchWildPokemonMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &catchPokemonMsg)
		if err != nil {
			return nil, wrapDeserializeLocationMsgError(err, Gyms)
		}
		return &catchPokemonMsg, nil

	case CatchPokemonResponse:
		var catchPokemonMsgResp CatchWildPokemonMessageResponse
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &catchPokemonMsgResp)
		if err != nil {
			return nil, wrapDeserializeLocationMsgError(err, Gyms)
		}
		return &catchPokemonMsgResp, nil

	default:
		return nil, wrapDeserializeLocationMsgError(ws.ErrorInvalidMessageType, msg.MsgType)
	}
}
