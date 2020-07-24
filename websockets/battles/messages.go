package battles

import (
	"github.com/NOVAPokemon/utils/pokemons"
	"github.com/NOVAPokemon/utils/websockets"
)

// Message types
const (
	// HTTP Requests
	Queue     = "QUEUE"
	Challenge = "CHALLENGE"

	Attack        = "ATTACK"
	Defend        = "DEFEND"
	UpdatePokemon = "UPDATE_POKEMON"
	RemoveItem    = "REMOVE_ITEM"
	UseItem       = "USE_ITEM"
	SelectPokemon = "SELECT_POKEMON"
	Status        = "STATUS"
)

type DefendMessage struct{}

func (dMsg DefendMessage) ConvertToWSMessage() *websockets.WebsocketMsg {
	return websockets.NewRequestMsg(Defend, nil)
}

type AttackMessage struct{}

func (aMsg AttackMessage) ConvertToWSMessage() *websockets.WebsocketMsg {
	return websockets.NewRequestMsg(Attack, nil)
}

type UpdatePokemonMessage struct {
	Owner   bool
	Pokemon pokemons.Pokemon
}

func (upMsg UpdatePokemonMessage) ConvertToWSMessage() *websockets.WebsocketMsg {
	return websockets.NewStandardMsg(UpdatePokemon, upMsg)
}

func (upMsg UpdatePokemonMessage) ConvertToWSMessageWithInfo(info websockets.TrackedInfo) *websockets.WebsocketMsg {
	return websockets.NewReplyMsg(UpdatePokemon, upMsg, info)
}

type RemoveItemMessage struct {
	ItemId string
}

func (riMsg RemoveItemMessage) ConvertToWSMessage(info websockets.TrackedInfo) *websockets.WebsocketMsg {
	return websockets.NewReplyMsg(RemoveItem, riMsg, info)
}

type UseItemMessage struct {
	ItemId string
}

func (uiMsg UseItemMessage) ConvertToWSMessage() *websockets.WebsocketMsg {
	return websockets.NewRequestMsg(UseItem, uiMsg)
}

type StatusMessage struct {
	Message string
}

func (sMsg StatusMessage) ConvertToWSMessage(info websockets.TrackedInfo) *websockets.WebsocketMsg {
	return websockets.NewReplyMsg(Status, sMsg, info)
}

type SelectPokemonMessage struct {
	PokemonId string
}

func (spMsg SelectPokemonMessage) ConvertToWSMessage() *websockets.WebsocketMsg {
	return websockets.NewRequestMsg(SelectPokemon, spMsg)
}
