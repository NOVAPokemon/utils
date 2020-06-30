package location

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	errorAddGym         = "error adding gym"
	errorGetGyms        = "error getting gyms"

	errorGetServerConfig       = "error getting server %s configs"
	errorUpdateServerConfig    = "error updating server %s configs"
	errorGetGlobalServerConfig = "error getting global server configs"
)

func wrapAddGymError(err error) error {
	return errors.Wrap(err, errorAddGym)
}

func wrapGetGymsError(err error) error {
	return errors.Wrap(err, errorGetGyms)
}

func wrapUpdateServerConfig(err error, serverName string) error {
	return errors.Wrap(err, fmt.Sprintf(errorUpdateServerConfig, serverName))
}

func wrapGetServerConfig(err error, serverName string) error {
	return errors.Wrap(err, fmt.Sprintf(errorGetServerConfig, serverName))
}

func wrapGetGlobalServerConfigs(err error) error {
	return errors.Wrap(err, errorGetGlobalServerConfig)
}
