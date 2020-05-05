package errors

import "github.com/pkg/errors"

const (
	errorStartLocationUpdate = "error starting location updates"
	errorAddGymLocation      = "error adding gym location"
	errorCatchWildPokemon    = "error catching wild pokemon"
	errorConnect             = "error connecting"
	errorUpdateLocation      = "error updating location"
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

func WrapUpdateLocation(err error) error {
	return errors.Wrap(err, errorUpdateLocation)
}
