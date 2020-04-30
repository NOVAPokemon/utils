package notification

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	errorGetNotificationFormat = "error getting user %s notifications"
	errorAddNotificationFormat = "error adding notification to user %s"
	errorRemoveNotificationFormat = "error removing notification %s"
)

var (
	errorNotificationNotFound = errors.New("notification doesnt exist")
)

func wrapGetNotificationError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorGetNotificationFormat, username))
}

func wrapAddNotificationError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorAddNotificationFormat, username))
}

func wrapRemoveNotificationError(err error, id string) error {
	return errors.Wrap(err, fmt.Sprintf(errorRemoveNotificationFormat, id))
}
