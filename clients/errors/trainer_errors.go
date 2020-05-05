package errors

import "github.com/pkg/errors"

const (
	errorAddTrainer           = "error adding trainer"
	errorListTrainers         = "error listing trainers"
	errorGetTrainerByUsername = "error getting trainer by username"
	errorUpdateTrainerStats   = "error updating trainer stats"
	errorAddItem              = "error adding item to trainer"
	errorRemoveItem           = "error removing item from trainer"
	errorAddPokemon           = "error adding pokemon to trainer"
	errorUpdatePokemon        = "error updating pokemon"
	errorRemovePokemon        = "error removing pokemon from trainer"
	errorGetAllTokens         = "error getting all tokens"
	errorGetStatsToken        = "error getting stats token"
	errorGetPokemonToken      = "error getting pokemon tokens"
	errorGetItemsToken        = "error getting items token"
	errorVerifyItems          = "error verifying items"
	errorVerifyPokemons       = "error verifying pokemons"
	errorVerifyStats          = "error verifying stats pokemons"
	errorSetPokemonTokens     = "error setting pokemon tokens"
	errorAppendPokemonTokens  = "error appending pokemon tokens"
	errorSetStatsToken        = "error setting stats token"
	errorSetItemsToken        = "error setting items token"
)

func WrapAddTrainerError(err error) error {
	return errors.Wrap(err, errorAddTrainer)
}

func WrapListTrainersError(err error) error {
	return errors.Wrap(err, errorListTrainers)
}

func WrapGetTrainerByUsernameError(err error) error {
	return errors.Wrap(err, errorGetTrainerByUsername)
}

func WrapUpdateTrainerStatsError(err error) error {
	return errors.Wrap(err, errorUpdateTrainerStats)
}

func WrapAddItemError(err error) error {
	return errors.Wrap(err, errorAddItem)
}

func WrapRemoveItemError(err error) error {
	return errors.Wrap(err, errorRemoveItem)
}

func WrapAddPokemonError(err error) error {
	return errors.Wrap(err, errorAddPokemon)
}

func WrapUpdatePokemonError(err error) error {
	return errors.Wrap(err, errorUpdatePokemon)
}

func WrapRemovePokemonError(err error) error {
	return errors.Wrap(err, errorRemovePokemon)
}

func WrapGetAllTokensError(err error) error {
	return errors.Wrap(err, errorGetAllTokens)
}

func WrapGetStatsTokenError(err error) error {
	return errors.Wrap(err, errorGetStatsToken)
}

func WrapGetPokemonTokenError(err error) error {
	return errors.Wrap(err, errorGetPokemonToken)
}

func WrapGetItemsTokenError(err error) error {
	return errors.Wrap(err, errorGetItemsToken)
}

func WrapVerifyItemsError(err error) error {
	return errors.Wrap(err, errorVerifyItems)
}

func WrapVerifyPokemonsError(err error) error {
	return errors.Wrap(err, errorVerifyPokemons)
}

func WrapVerifyStatsError(err error) error {
	return errors.Wrap(err, errorVerifyStats)
}

func WrapSetPokemonTokensError(err error) error {
	return errors.Wrap(err, errorSetPokemonTokens)
}

func WrapAppendPokemonTokensError(err error) error {
	return errors.Wrap(err, errorAppendPokemonTokens)
}

func WrapSetStatsTokenError(err error) error {
	return errors.Wrap(err, errorSetStatsToken)
}

func WrapSetItemsTokenError(err error) error {
	return errors.Wrap(err, errorSetItemsToken)
}
