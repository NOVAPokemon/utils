package location

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	errorGetServerConfig    = "error updating server %s configs"
	errorInsertServerConfig = "error updating server %s configs"
	errorUpdateServerConfig = "error updating server %s configs"
)

func wrapUpdateServerConfig(err error, serverName string) error {
	return errors.Wrap(err, fmt.Sprintf(errorUpdateServerConfig, serverName))
}

func wrapGetServerConfig(err error, serverName string) error {
	return errors.Wrap(err, fmt.Sprintf(errorGetServerConfig, serverName))
}

func wrapInsertServerConfig(err error, serverName string) error {
	return errors.Wrap(err, fmt.Sprintf(errorInsertServerConfig, serverName))
}
