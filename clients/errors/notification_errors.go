package errors

import "github.com/pkg/errors"

const (
	errorListeningNotifications = "error listening to notifications"
	errorStopListening          = "error stopping listen on notifications"
	errorAddNotification        = "error adding notification"
	errorGetOthersListening     = "error getting others listening"
)

func WrapListeningNotificationsError(err error) error {
	return errors.Wrap(err, errorListeningNotifications)
}

func WrapStopListeningError(err error) error {
	return errors.Wrap(err, errorStopListening)
}

func WrapAddNotificationError(err error) error {
	return errors.Wrap(err, errorAddNotification)
}

func WrapGetOthersListeningError(err error) error {
	return errors.Wrap(err, errorGetOthersListening)
}
