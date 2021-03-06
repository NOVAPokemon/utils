package battles

import (
	"time"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/pokemons"
	"github.com/NOVAPokemon/utils/websockets"
)

var (
	StatusDefended = "You defended an attack"
	StatusDefending = "You are defending"
	StatusEnemyDefended = "Enemy defended your attack"
)

type (
	BattleChannels struct {
		OutChannel      chan *websockets.WebsocketMsg
		InChannel       chan *websockets.WebsocketMsg
		RejectedChannel chan struct{}
		FinishChannel   chan struct{}
	}

	TrainerBattleStatus struct {
		Username        string
		TrainerStats    *utils.TrainerStats
		TrainerPokemons map[string]*pokemons.Pokemon
		TrainerItems    map[string]items.Item
		SelectedPokemon *pokemons.Pokemon
		AllPokemonsDead bool
		Defending       bool
		Cooldown        bool
		CdTimer         *time.Timer
		UsedItems       map[string]items.Item
	}
)

func (status *TrainerBattleStatus) AreAllPokemonsDead() bool {
	for _, pokemon := range status.TrainerPokemons {
		if pokemon.HP > 0 {
			return false
		}
	}
	return true
}
