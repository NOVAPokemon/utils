package items

import (
	"errors"
	"github.com/NOVAPokemon/utils/pokemons"
)

const (
	NoId = iota
	HealId
	ReviveId
	PokeBallId
	MasterBallId

	NoValue = 0

	PokeBallValue   = 60
	MasterBallValue = 100

	HealFactor = 100
)

type Effect struct {
	Appliable bool
	Id        int
	Value     int
}

var (
	ErrNotAppliable = errors.New("item not appliable")
	ErrInvalidId    = errors.New("invalid item id")
)

var (
	Heal       = Effect{Appliable: true, Id: HealId, Value: NoValue}
	Revive     = Effect{Appliable: true, Id: ReviveId, Value: NoValue}
	PokeBall   = Effect{Appliable: false, Id: PokeBallId, Value: PokeBallValue}
	MasterBall = Effect{Appliable: false, Id: MasterBallId, Value: MasterBallValue}

	None = Effect{Appliable: false, Id: NoId, Value: NoValue}
)

func (item *Item) Apply(pokemon *pokemons.Pokemon) error {
	if !item.Effect.Appliable {
		return ErrNotAppliable
	}

	switch item.Effect.Id {
	case HealId:
		pokemon.HP += HealFactor
	case ReviveId:
		pokemon.HP = pokemon.MaxHP
	default:
		return ErrInvalidId
	}

	return nil
}

func GetEffectForItem(itemName string) Effect {
	switch itemName {
	case HealName:
		return Heal
	case ReviveName:
		return Revive
	case PokeBallName:
		return PokeBall
	case MasterBallName:
		return MasterBall
	default:
		return None
	}
}
