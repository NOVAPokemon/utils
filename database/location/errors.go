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
	errorDeleteLocation = "error deleting user %s location"

	errorGetServerConfig       = "error getting server %s configs"
	errorInsertServerConfig    = "error inserting server %s configs"
	errorUpdateServerConfig    = "error updating server %s configs"
	errorGetGlobalServerConfig = "error getting global server configs"
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

func wrapUpdateServerConfig(err error, serverName string) error {
	return errors.Wrap(err, fmt.Sprintf(errorUpdateServerConfig, serverName))
}

func wrapGetServerConfig(err error, serverName string) error {
	return errors.Wrap(err, fmt.Sprintf(errorGetServerConfig, serverName))
}

func wrapInsertServerConfig(err error, serverName string) error {
	return errors.Wrap(err, fmt.Sprintf(errorInsertServerConfig, serverName))
}

func wrapGetGlobalServerConfigs(err error) error {
	return errors.Wrap(err, errorGetGlobalServerConfig)
}
