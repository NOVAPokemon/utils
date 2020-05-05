package tokens

import "github.com/pkg/errors"

var (
	ErrorInvalidPokemonTokens = errors.New("invalid pokemon hashes")
	ErrorInvalidStatsToken    = errors.New("invalid stats token")
	ErrorInvalidItemsToken    = errors.New("invalid items token")
)
