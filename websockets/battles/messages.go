package battles

import (
	"encoding/json"
	"github.com/NOVAPokemon/utils/pokemons"
	"github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Message types
const (
	Attack        = "ATTACK"
	Defend        = "DEFEND"
	UpdatePokemon = "UPDATE_POKEMON"
	RemoveItem    = "REMOVE_ITEM"
	UseItem       = "USE_ITEM"
	SelectPokemon = "SELECT_POKEMON"
	Status        = "STATUS"
)

type DefendMessage struct {
	websockets.MessageWithId
}

func (aMsg DefendMessage) SerializeToWSMessage() *websockets.Message {
	return &websockets.Message{
		MsgType: Defend,
		MsgArgs: nil,
	}
}

type AttackMessage struct {
	websockets.TrackedMessage
}

func NewAttackMessage() AttackMessage {
	return AttackMessage{
		TrackedMessage: websockets.NewTrackedMessage(primitive.NewObjectID()),
	}
}

func (aMsg AttackMessage) SerializeToWSMessage() *websockets.Message {
	msgJson, err := json.Marshal(aMsg)
	if err != nil {
		log.Error(websockets.WrapSerializeToWSMessageError(err, Attack))
		return nil
	}

	return &websockets.Message{
		MsgType: Attack,
		MsgArgs: []string{string(msgJson)},
	}
}

type UpdatePokemonMessage struct {
	Owner   bool
	Pokemon pokemons.Pokemon
	websockets.TrackedMessage
}

func (upMsg UpdatePokemonMessage) SerializeToWSMessage() *websockets.Message {
	msgJson, err := json.Marshal(upMsg)
	if err != nil {
		log.Error(websockets.WrapSerializeToWSMessageError(err, UpdatePokemon))
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
		log.Error(websockets.WrapSerializeToWSMessageError(err, RemoveItem))
		return nil
	}

	return &websockets.Message{
		MsgType: RemoveItem,
		MsgArgs: []string{string(msgJson)},
	}
}

type UseItemMessage struct {
	ItemId string
	websockets.TrackedMessage
}

func NewUseItemMessage(itemId string) UseItemMessage {
	return UseItemMessage{
		ItemId:         itemId,
		TrackedMessage: websockets.NewTrackedMessage(primitive.NewObjectID()),
	}
}

func (uiMsg UseItemMessage) SerializeToWSMessage() *websockets.Message {
	msgJson, err := json.Marshal(uiMsg)
	if err != nil {
		log.Error(websockets.WrapSerializeToWSMessageError(err, UseItem))
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
		log.Error(websockets.WrapSerializeToWSMessageError(err, Status))
		return nil
	}

	return &websockets.Message{
		MsgType: Status,
		MsgArgs: []string{string(msgJson)},
	}
}

type SelectPokemonMessage struct {
	PokemonId string
	websockets.TrackedMessage
}

func NewSelectPokemonMessage(pokemonId string) SelectPokemonMessage {
	return SelectPokemonMessage{
		PokemonId:      pokemonId,
		TrackedMessage: websockets.NewTrackedMessage(primitive.NewObjectID()),
	}
}

func (spMsg SelectPokemonMessage) SerializeToWSMessage() *websockets.Message {
	msgJson, err := json.Marshal(spMsg)
	if err != nil {
		log.Error(websockets.WrapSerializeToWSMessageError(err, SelectPokemon))
		return nil
	}

	return &websockets.Message{
		MsgType: SelectPokemon,
		MsgArgs: []string{string(msgJson)},
	}
}

func DeserializeBattleMsg(msg *websockets.Message) (websockets.Serializable, error) {
	switch msg.MsgType {
	case Status:
		var statusMsg StatusMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &statusMsg)
		if err != nil {
			return nil, wrapDeserializeBattleMsgError(err, Status)
		}
		return &statusMsg, nil
	case Attack:
		var attackMessage AttackMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &attackMessage)
		if err != nil {
			return nil, wrapDeserializeBattleMsgError(err, Attack)
		}
		return &attackMessage, nil
	case Defend:
		var defendMessage DefendMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &defendMessage)
		if err != nil {
			return nil, wrapDeserializeBattleMsgError(err, Defend)
		}
		return &defendMessage, nil
	case UpdatePokemon:
		var updatePokemonMessage UpdatePokemonMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &updatePokemonMessage)
		if err != nil {
			return nil, wrapDeserializeBattleMsgError(err, UpdatePokemon)
		}
		return &updatePokemonMessage, nil
	case RemoveItem:
		var removeItemMessage RemoveItemMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &removeItemMessage)
		if err != nil {
			return nil, wrapDeserializeBattleMsgError(err, RemoveItem)
		}
		return &removeItemMessage, nil
	case UseItem:
		var useItemMessage UseItemMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &useItemMessage)
		if err != nil {
			return nil, wrapDeserializeBattleMsgError(err, UseItem)
		}
		return &useItemMessage, nil
	case SelectPokemon:
		var selectPokemonMessage SelectPokemonMessage
		err := json.Unmarshal([]byte(msg.MsgArgs[0]), &selectPokemonMessage)
		if err != nil {
			return nil, wrapDeserializeBattleMsgError(err, SelectPokemon)
		}
		return &selectPokemonMessage, nil
	default:
		deserializedMsg, err := websockets.DeserializeMsg(msg)
		if err != nil {
			return nil, wrapDeserializeBattleMsgError(err, msg.MsgType)
		}
		return deserializedMsg, nil
	}
}
