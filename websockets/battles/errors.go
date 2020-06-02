package battles

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	errorDeserializeBattleMessageFormat = "error deserializing battle message type %s"
)

var (
	ErrorPokemonSelectionPhase  = errors.New("error in pokemon selection phase")
	ErrorInvalidPokemonSelected = errors.New("invalid pokemon selected")
	ErrorPokemonNoHP            = errors.New("pokemon has no hp")
	ErrorCooldown               = errors.New("player still in cooldown")
	ErrorInvalidItemSelected    = errors.New("invalid item selected")
	ErrorItemNotAppliable       = errors.New("error item not appliable")
)

func wrapDeserializeBattleMsgError(err error, msgType string) error {
	return errors.Wrap(err, fmt.Sprintf(errorDeserializeBattleMessageFormat, msgType))
}
