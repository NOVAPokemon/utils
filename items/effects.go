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

	PokeBallValue   = 75
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
	HealEffect       = Effect{Appliable: true, Id: HealId, Value: NoValue}
	ReviveEffect     = Effect{Appliable: true, Id: ReviveId, Value: NoValue}
	PokeBallEffect   = Effect{Appliable: false, Id: PokeBallId, Value: PokeBallValue}
	MasterBallEffect = Effect{Appliable: false, Id: MasterBallId, Value: MasterBallValue}

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
		return HealEffect
	case ReviveName:
		return ReviveEffect
	case PokeBallName:
		return PokeBallEffect
	case MasterBallName:
		return MasterBallEffect
	default:
		return None
	}
}
