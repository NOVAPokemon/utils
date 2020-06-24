package errors

import "github.com/pkg/errors"

const (
	errorStartLocationUpdate  = "error starting location updates"
	errorAddGymLocation       = "error adding gym location"
	errorCatchWildPokemon     = "error catching wild pokemon"
	errorConnect              = "error connecting"
	errorGetServerForLocation = "error fetching server for current location"
	errorUpdateConnections    = "error updating connections"
	errorHandleLocationMsg    = "error handling location message"
)

var (
	ErrorNoPokemonsVinicity = errors.New("no pokemons in vicinity")
	ErrorNoPokeballs        = errors.New("no pokeballs")
)

func WrapStartLocationUpdatesError(err error) error {
	return errors.Wrap(err, errorStartLocationUpdate)
}

func WrapAddGymLocationError(err error) error {
	return errors.Wrap(err, errorAddGymLocation)
}

func WrapCatchWildPokemonError(err error) error {
	return errors.Wrap(err, errorCatchWildPokemon)
}

func WrapConnectError(err error) error {
	return errors.Wrap(err, errorConnect)
}

func WrapGetServerForLocationError(err error) error {
	return errors.Wrap(err, errorGetServerForLocation)
}

func WrapUpdateConnectionsError(err error) error {
	return errors.Wrap(err, errorUpdateConnections)
}

func WrapHandleLocationMsgError(err error) error {
	return errors.Wrap(err, errorHandleLocationMsg)
}
