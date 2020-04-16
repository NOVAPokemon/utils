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
	ATTACK   = "ATTACK"
	USE_ITEM = "USE_ITEM"
	DEFEND   = "DEFEND"

	// Update Types
	UPDATE_PLAYER_POKEMON    = "UPDATE_PLAYER_POKEMON"
	UPDATE_ADVERSARY_POKEMON = "UPDATE_ADVERSARY_POKEMON"
	REMOVE_ITEM              = "REMOVE_ITEM"

	DEFEND_SUCCESS      = "DEFEND_SUCCESS"
	ADVERSARY_DEFENDING = "ADVERSARY_DEFENDING"

	// Setup Message Types
	SELECT_POKEMON = "SELECT_POKEMON"

	STATUS = "STATUS"
	ERROR  = "ERROR"
	FINISH = "FINISH_BATTLE"
	START  = "START"

	SET_TOKEN = "SET_TOKEN"
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
