package tokens

import "github.com/pkg/errors"

const (
	errorExtractVerifyAuthToken    = "error extracting and verifying auth token"
	errorExtractVerifyStatsToken   = "error extracting and verifying stats token"
	errorExtractVerifyPokemonToken = "error extracting and verifying pokemon tokens"
	errorExtractVerifyItemsToken   = "error extracting and verifying items token"
	errorExtractStatsToken         = "error extracting stats token"
	errorExtractPokemonToken       = "error extracting pokemon token"
	errorExtractItemsToken         = "error extracting items token"

	errorParsingToken              = "error parsing token"
)

var (
	ErrorNoPokemonTokens = errors.New("no pokemon tokens in header")

	ErrorInvalidAuthToken     = errors.New("invalid auth token")
	ErrorInvalidPokemonTokens = errors.New("invalid pokemon hashes")
	ErrorInvalidStatsToken    = errors.New("invalid stats token")
	ErrorInvalidItemsToken    = errors.New("invalid items token")
)

func wrapExtractVerifyAuthTokenError(err error) error {
	return errors.Wrap(err, errorExtractVerifyAuthToken)
}

func wrapExtractVerifyStatsTokenError(err error) error {
	return errors.Wrap(err, errorExtractVerifyStatsToken)
}

func wrapExtractVerifyPokemonTokensError(err error) error {
	return errors.Wrap(err, errorExtractVerifyPokemonToken)
}

func wrapExtractVerifyItemsTokenError(err error) error {
	return errors.Wrap(err, errorExtractVerifyItemsToken)
}

func wrapExtractStatsTokenError(err error) error {
	return errors.Wrap(err, errorExtractStatsToken)
}

func wrapExtractPokemonTokenError(err error) error {
	return errors.Wrap(err, errorExtractPokemonToken)
}

func wrapExtractItemsTokenError(err error) error {
	return errors.Wrap(err, errorExtractItemsToken)
}

func wrapParsingTokenError(err error) error {
	return errors.Wrap(err, errorParsingToken)
}
