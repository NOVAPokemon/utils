package notifications

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	errorDeserializeNotificationMessageFormat = "error deserializing notification message type %s"
)

func wrapDeserializeNotificationMsgError(err error, msgType string) error {
	return errors.Wrap(err, fmt.Sprintf(errorDeserializeNotificationMessageFormat, msgType))
}