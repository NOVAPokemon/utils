package battles

import (
	"errors"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/pokemons"
	"time"
)

var (
	ErrInvalidMessageFormat   = errors.New("ErrInvalidMessageFormat")
	ErrInvalidMessageType     = errors.New("ErrInvalidMessageType")
	ErrPokemonSelectionPhase  = errors.New("ErrPokemonSelectionPhase")
	ErrInvalidPokemonSelected = errors.New("ErrInvalidPokemonSelected")
	ErrPokemonNoHP            = errors.New("ErrPokemonNoHP")
	ErrCooldown               = errors.New("ErrCooldown")
	ErrInvalidItemSelected    = errors.New("ErrInvalidItemSelected")
	ErrItemNotAppliable       = errors.New("ErrItemNotAppliable")

	StatusDefended = "You defended an attack"
)

type BattleChannels struct {
	OutChannel    chan *string
	InChannel     chan *string
	FinishChannel chan struct{}
}

type (
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
