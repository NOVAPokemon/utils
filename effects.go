package utils

import "errors"

const (
	HealId = iota
	ReviveId
	PokeBallId
	MasterBallId

	NoValue = 0

	PokeBallValue = 60
	MasterBallValue = 100

	HealFactor = 100
)

var (
	ErrNotAppliable = errors.New("item not appliable")
	ErrInvalidId =  errors.New("invalid item id")
)

var (
	Heal       = Effect{Appliable: true, Id: HealId, Value: NoValue}
	Revive     = Effect{Appliable: true, Id: ReviveId, Value: NoValue}
	PokeBall   = Effect{Appliable: false, Id: PokeBallId, Value: PokeBallValue}
	MasterBall = Effect{Appliable: false, Id: MasterBallId, Value: MasterBallValue}
)

func (item *Item) Apply(pokemon *Pokemon) error {
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
