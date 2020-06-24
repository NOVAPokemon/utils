package location

import (
	"encoding/json"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/items"
	ws "github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
)

const (
	UpdateLocation          = "UPDATE_LOCATION"
	UpdateLocationWithTiles = "UPDATE_LOCATION_WITH_TILES"
	Gyms                    = "GYMS"
	Pokemon                 = "POKEMON"
	CatchPokemon            = "CATCH_POKEMON"
	CatchPokemonResponse    = "CATCH_POKEMON_RESPONSE"
	ServersResponse         = "SERVERS_RESPONSE"
	TilesResponse           = "TILES_RESPONSE"
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

type UpdateLocationWithTilesMessage struct {
	TilesPerServer map[string][]int
	ws.MessageWithId
}

func (ulwtMsg UpdateLocationWithTilesMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(ulwtMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: UpdateLocationWithTiles,
		MsgArgs: []string{string(msgJson)},
	}
}

type ServersMessage struct {
	Servers []string
	ws.MessageWithId
}

func (getResponseMsg ServersMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(getResponseMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: ServersResponse,
		MsgArgs: []string{string(msgJson)},
	}
}

type TilesPerServerMessage struct {
	TilesPerServer map[string][]int
	OriginServer string
	ws.MessageWithId
}

func (tilesMsg TilesPerServerMessage) SerializeToWSMessage() *ws.Message {
	msgJson, err := json.Marshal(tilesMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &ws.Message{
		MsgType: TilesResponse,
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
	Pokemon []utils.WildPokemonWithServer
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
	Pokeball    items.Item
	WildPokemon utils.WildPokemonWithServer
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
	Error         string
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

	case UpdateLocationWithTiles:
		var locationWithTilesMsg UpdateLocationWithTilesMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &locationWithTilesMsg)
		if err != nil {
			return nil, wrapDeserializeLocationMsgError(err, UpdateLocationWithTiles)
		}
		return &locationWithTilesMsg, nil

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
			return nil, wrapDeserializeLocationMsgError(err, CatchPokemon)
		}
		return &catchPokemonMsg, nil

	case CatchPokemonResponse:
		var catchPokemonMsgResp CatchWildPokemonMessageResponse
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &catchPokemonMsgResp)
		if err != nil {
			return nil, wrapDeserializeLocationMsgError(err, CatchPokemonResponse)
		}
		return &catchPokemonMsgResp, nil

	case ServersResponse:
		var serversMessage ServersMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &serversMessage)
		if err != nil {
			return nil, wrapDeserializeLocationMsgError(err, ServersResponse)
		}
		return &serversMessage, nil

	case TilesResponse:
		var tilesMsg TilesPerServerMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &tilesMsg)
		if err != nil {
			return nil, wrapDeserializeLocationMsgError(err, TilesResponse)
		}
		return &tilesMsg, nil

	default:
		return nil, wrapDeserializeLocationMsgError(ws.ErrorInvalidMessageType, msg.MsgType)
	}
}
