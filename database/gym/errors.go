package location

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	errorGetConfig          = "error getting configs"
	errorUpdateServerConfig = "error updating server %s configs"
)

func wrapUpdateServerConfig(err error, serverName string) error {
	return errors.Wrap(err, fmt.Sprintf(errorUpdateServerConfig, serverName))
}

func wrapGetConfig(err error) error {
	return errors.Wrap(err, errorGetConfig)
}
