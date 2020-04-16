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

	StatusDefended         = "StatusDefended"
	StatusOpponentDefended = "StatusOpponentDefended"
)

const (
	Attack  = "ATTACK"
	UseItem = "USE_ITEM"
	Defend  = "DEFEND"

	// Update Types
	UpdatePlayerPokemon    = "UPDATE_PLAYER_POKEMON"
	UpdateAdversaryPokemon = "UPDATE_ADVERSARY_POKEMON"
	RemoveItem             = "REMOVE_ITEM"

	DefendSuccess      = "DEFEND_SUCCESS"
	AdversaryDefending = "ADVERSARY_DEFENDING"

	// Setup Message Types
	SelectPokemon = "SELECT_POKEMON"

	Status = "STATUS"
	Error  = "ERROR"
	Finish = "FINISH_BATTLE"
	Start  = "START"

	SetToken = "SET_TOKEN"
)

type (
	TrainerBattleStatus struct {
		Username        string
		TrainerStats    *utils.TrainerStats
		TrainerPokemons map[string]*pokemons.Pokemon
		TrainerItems    map[string]items.Item
		SelectedPokemon *pokemons.Pokemon
		Defending       bool
		Cooldown        bool
		CdTimer         *time.Timer
		UsedItems       map[string]items.Item
	}
)
