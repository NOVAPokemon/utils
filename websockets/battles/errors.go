package battles

import "github.com/pkg/errors"

var (
	ErrorInvalidMessageFormat   = errors.New("ErrInvalidMessageFormat")
	ErrorInvalidMessageType     = errors.New("ErrInvalidMessageType")
	ErrorPokemonSelectionPhase  = errors.New("ErrPokemonSelectionPhase")
	ErrorInvalidPokemonSelected = errors.New("ErrInvalidPokemonSelected")
	ErrorPokemonNoHP            = errors.New("ErrPokemonNoHP")
	ErrorCooldown               = errors.New("ErrCooldown")
	ErrorInvalidItemSelected    = errors.New("ErrInvalidItemSelected")
	ErrorItemNotAppliable       = errors.New("ErrItemNotAppliable")
)
