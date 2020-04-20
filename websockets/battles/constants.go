package battles

import (
	"errors"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/pokemons"
	"time"
)

const PokemonsPerBattle = 3
const PokemonsPerRaid = 3
const DefaultCooldown = time.Millisecond * 500
const TimeToStartRaid = time.Second * 20
const DefaultRaidCapacity = 10

var (
	ErrInvalidMessageFormat   = errors.New("ErrInvalidMessageFormat")
	ErrInvalidMessageType     = errors.New("ErrInvalidMessageType")
	ErrPokemonSelectionPhase  = errors.New("ErrPokemonSelectionPhase")
	ErrNoPokemonSelected      = errors.New("ErrNoPokemonSelected")
	ErrInvalidPokemonSelected = errors.New("ErrInvalidPokemonSelected")
	ErrPokemonNoHP            = errors.New("ErrPokemonNoHP")
	ErrCooldown               = errors.New("ErrCooldown")
	ErrNoItemSelected         = errors.New("ErrNoItemSelected")
	ErrInvalidItemSelected    = errors.New("ErrInvalidItemSelected")
	ErrItemNotAppliable       = errors.New("ErrItemNotAppliable")

	StatusDefended         = "You defended an attack"
	StatusOpponentDefended = "Opponent defended the attack"
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

// status messages
const (
	DefendSuccess      = "DEFEND_SUCCESS"
	AdversaryDefending = "ADVERSARY_DEFENDING"
)
