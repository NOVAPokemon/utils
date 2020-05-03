package battles

import "github.com/pkg/errors"

var (
	ErrorPokemonSelectionPhase  = errors.New("ErrPokemonSelectionPhase")
	ErrorInvalidPokemonSelected = errors.New("ErrInvalidPokemonSelected")
	ErrorPokemonNoHP            = errors.New("ErrPokemonNoHP")
	ErrorCooldown               = errors.New("ErrCooldown")
	ErrorInvalidItemSelected    = errors.New("ErrInvalidItemSelected")
	ErrorItemNotAppliable       = errors.New("ErrItemNotAppliable")
)
