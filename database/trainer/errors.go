package trainer

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	errorGetAllTrainers    = "error getting all trainers"
	errorDeleteAllTrainers = "error deleting all trainers"

	errorAddTrainerFormat         = "error adding trainer %s"
	errorGetTrainerFormat         = "error getting trainer %s"
	errorUpdateTrainerStatsFormat = "error updating trainer %s stats"
	errorDeleteTrainerFormat      = "error deleting trainer %s"

	errorAddItemToTrainerFormat  = "error adding item to trainer %s"
	errorAddItemsToTrainerFormat = "error adding items to trainer %s"

	errorRemoveItemFromTrainerFormat  = "error remove item from trainer %s"
	errorRemoveItemsFromTrainerFormat = "error remove items from trainer %s"

	errorAddPokemonToTrainerFormat = "error add pokemon to trainer %s"
	errorUpdateTrainerPokemonsFormat = "error update trainer %s pokemons"
	errorRemovePokemonFromTrainerFormat = "error removing pokemon from trainer %s"
)

var (
	ErrorTrainerNotFound = errors.New("trainer not found")
	ErrorInvalidLevel    = errors.New("invalid level")
	ErrorInvalidCoins    = errors.New("invalid coin ammount")
	ErrorItemNotFound    = errors.New("item not found")
	ErrorPokemonNotFound = errors.New("pokemon not found")
)

func wrapAddTrainerError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorAddTrainerFormat, username))
}

func wrapGetAllTrainersError(err error) error {
	return errors.Wrap(err, errorGetAllTrainers)
}

func wrapGetTrainerError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorGetTrainerFormat, username))
}

func wrapUpdateTrainerStatsError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorUpdateTrainerStatsFormat, username))
}

func wrapDeleteTrainerError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorDeleteTrainerFormat, username))
}

func wrapDeleteAllTrainersError(err error) error {
	return errors.Wrap(err, errorDeleteAllTrainers)
}

func wrapAddItemToTrainerError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorAddItemToTrainerFormat, username))
}

func wrapAddItemsToTrainerError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorAddItemsToTrainerFormat, username))
}

func wrapRemoveItemToTrainerError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorRemoveItemFromTrainerFormat, username))
}

func wrapRemoveItemsToTrainerError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorRemoveItemsFromTrainerFormat, username))
}

func wrapAddPokemonToTrainerError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorAddPokemonToTrainerFormat, username))
}

func wrapUpdateTrainerPokemonError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorUpdateTrainerPokemonsFormat, username))
}

func wrapRemovePokemonFromTrainerError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorRemovePokemonFromTrainerFormat, username))
}
