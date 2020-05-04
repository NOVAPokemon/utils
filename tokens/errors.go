package tokens

import "github.com/pkg/errors"

var (
	ErrorPokemonTokens = errors.New("invalid pokemon hashes")
	ErrorStatsToken    = errors.New("invalid stats token")
	ErrorItemsToken    = errors.New("invalid items token")
)
