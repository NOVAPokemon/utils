package notification

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	errorAddNotificationFormat    = "error adding notification to user %s"
	errorRemoveNotificationFormat = "error removing notification %s"
)

var (
	errorNotificationNotFound = errors.New("notification doesnt exist")
)

func wrapAddNotificationError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorAddNotificationFormat, username))
}

func wrapRemoveNotificationError(err error, id string) error {
	return errors.Wrap(err, fmt.Sprintf(errorRemoveNotificationFormat, id))
}
