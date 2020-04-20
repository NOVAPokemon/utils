package battles

import (
	"encoding/json"
	"github.com/NOVAPokemon/utils/pokemons"
	"github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
)

// Message types
const (
	Attack        = "ATTACK"
	Defend        = "DEFEND"
	Error         = "ERROR"          // done
	UpdatePokemon = "UPDATE_POKEMON" // done
	RemoveItem    = "REMOVE_ITEM"    // done
	SetToken      = "SET_TOKEN"      // done
	UseItem       = "USE_ITEM"       // done
	SelectPokemon = "SELECT_POKEMON" // done
	Start         = "START"          // done
	Finish        = "FINISH"         // done
	Status        = "STATUS"         // done

)

type DefendMessage struct {
	websockets.MessageWithId
}

func (aMsg DefendMessage) SerializeToWSMessage() *websockets.Message {
	return &websockets.Message{
		MsgType: Defend,
		MsgArgs: []string{},
	}
}

type AttackMessage struct {
	websockets.MessageWithId
}

func (aMsg AttackMessage) SerializeToWSMessage() *websockets.Message {
	return &websockets.Message{
		MsgType: Attack,
		MsgArgs: []string{},
	}
}

type UpdatePokemonMessage struct {
	Owner   bool
	Pokemon pokemons.Pokemon
	websockets.MessageWithId
}

func (upMsg UpdatePokemonMessage) SerializeToWSMessage() *websockets.Message {
	msgJson, err := json.Marshal(upMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &websockets.Message{
		MsgType: UpdatePokemon,
		MsgArgs: []string{string(msgJson)},
	}
}

type RemoveItemMessage struct {
	ItemId string
	websockets.MessageWithId
}

func (riMsg RemoveItemMessage) SerializeToWSMessage() *websockets.Message {
	msgJson, err := json.Marshal(riMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &websockets.Message{
		MsgType: RemoveItem,
		MsgArgs: []string{string(msgJson)},
	}
}

type UseItemMessage struct {
	ItemId string
	websockets.MessageWithId
}

func (uiMsg UseItemMessage) SerializeToWSMessage() *websockets.Message {
	msgJson, err := json.Marshal(uiMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &websockets.Message{
		MsgType: UseItem,
		MsgArgs: []string{string(msgJson)},
	}
}

type StatusMessage struct {
	Message string
	websockets.MessageWithId
}

func (statusMsg StatusMessage) SerializeToWSMessage() *websockets.Message {
	msgJson, err := json.Marshal(statusMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &websockets.Message{
		MsgType: Status,
		MsgArgs: []string{string(msgJson)},
	}
}

type SelectPokemonMessage struct {
	PokemonId string
	websockets.MessageWithId
}

func (spMsg SelectPokemonMessage) SerializeToWSMessage() *websockets.Message {
	msgJson, err := json.Marshal(spMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &websockets.Message{
		MsgType: SelectPokemon,
		MsgArgs: []string{string(msgJson)},
	}
}

// Could be utils messages:

// Start
type StartMessage struct {
	websockets.MessageWithId
}

func (sMsg StartMessage) SerializeToWSMessage() *websockets.Message {
	msgJson, err := json.Marshal(sMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &websockets.Message{
		MsgType: Start,
		MsgArgs: []string{string(msgJson)},
	}
}

// Finish
type FinishMessage struct {
	Success bool
	websockets.MessageWithId
}

func (fMsg FinishMessage) SerializeToWSMessage() *websockets.Message {
	msgJson, err := json.Marshal(fMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &websockets.Message{
		MsgType: Finish,
		MsgArgs: []string{string(msgJson)},
	}
}

// SetToken
type SetTokenMessage struct {
	TokenField   string
	TokensString [] string
	websockets.MessageWithId
}

func (sMsg SetTokenMessage) SerializeToWSMessage() *websockets.Message {
	msgJson, err := json.Marshal(sMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &websockets.Message{
		MsgType: SetToken,
		MsgArgs: []string{string(msgJson)},
	}
}

// Error
type ErrorMessage struct {
	Info  string
	Fatal bool
	websockets.MessageWithId
}

func (eMsg ErrorMessage) SerializeToWSMessage() *websockets.Message {
	msgJson, err := json.Marshal(eMsg)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &websockets.Message{
		MsgType: Error,
		MsgArgs: []string{string(msgJson)},
	}
}

/*
const (
	Status        = "STATUS"         // done
	Start         = "START"          // done
	Finish        = "FINISH"         // done
	Error         = "ERROR"          // done
	SetToken      = "SET_TOKEN"      // done

	Move          = "MOVE"           // done
	UpdatePokemon = "UPDATE_POKEMON" // done
	RemoveItem    = "REMOVE_ITEM"    // done
	UseItem       = "USE_ITEM"       // done
	SelectPokemon = "SELECT_POKEMON" // done
)
*/

func DeserializeBattleMsg(msg *websockets.Message) websockets.Serializable {
	switch msg.MsgType {
	case Status:
		var statusMsg StatusMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &statusMsg)
		if err != nil {
			log.Error(err)
			return nil
		}
		return &statusMsg

	case Start:
		var startMsg StartMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &startMsg)
		if err != nil {
			log.Error(err)
			return nil
		}
	case Finish:
		var finishMsg FinishMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &finishMsg)
		if err != nil {
			log.Error(err)
			return nil
		}
		return &finishMsg

	case Error:
		var errMsg ErrorMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &errMsg)
		if err != nil {
			log.Error(err)
			return nil
		}
		return &errMsg

	case SetToken:
		var setTokenMessage SetTokenMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &setTokenMessage)
		if err != nil {
			log.Error(err)
			return nil
		}
		return &setTokenMessage

	case Attack:
		var attackMessage AttackMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &attackMessage)
		if err != nil {
			log.Error(err)
			return nil
		}
		return &attackMessage

	case Defend:
		var defendMessage DefendMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &defendMessage)
		if err != nil {
			log.Error(err)
			return nil
		}
		return &defendMessage

	case UpdatePokemon:
		var updatePokemonMessage UpdatePokemonMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &updatePokemonMessage)
		if err != nil {
			log.Error(err)
			return nil
		}
		return &updatePokemonMessage

	case RemoveItem:
		var removeItemMessage RemoveItemMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &removeItemMessage)
		if err != nil {
			log.Error(err)
			return nil
		}
		return &removeItemMessage

	case UseItem:
		var useItemMessage UseItemMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &useItemMessage)
		if err != nil {
			log.Error(err)
			return nil
		}
		return &useItemMessage
	case SelectPokemon:
		var selectPokemonMessage SelectPokemonMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &selectPokemonMessage)
		if err != nil {
			log.Error(err)
			return nil
		}
		return &selectPokemonMessage

	default:
		log.Info("invalid msg type")
		return nil
	}
	return nil
}
