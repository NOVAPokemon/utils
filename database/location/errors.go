package location

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	errorAddGym         = "error adding gym"
	errorGetGyms        = "error getting gyms"
	errorUpdateLocation = "error updating user %s location"
	errorGetLocation    = "error getting user %s location"
	errorDeleteLocation    = "error deleting user %s location"
)

func wrapAddGymError(err error) error {
	return errors.Wrap(err, errorAddGym)
}

func wrapGetGymsError(err error) error {
	return errors.Wrap(err, errorGetGyms)
}

func wrapUpdateLocation(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorUpdateLocation, username))
}

func wrapGetLocation(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorGetLocation, username))
}

func wrapDeleteLocation(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorDeleteLocation, username))
}
