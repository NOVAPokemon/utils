package battles

import (
	"github.com/pkg/errors"
)

var (
	ErrorPokemonSelectionPhase  = errors.New("error in pokemon selection phase")
	ErrorInvalidPokemonSelected = errors.New("invalid pokemon selected")
	ErrorPokemonNoHP            = errors.New("pokemon has no hp")
	ErrorCooldown               = errors.New("player still in cooldown")
	ErrorInvalidItemSelected    = errors.New("invalid item selected")
	ErrorItemNotAppliable       = errors.New("error item not appliable")
)

