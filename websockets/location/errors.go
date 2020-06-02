package location

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	errorDeserializeLocationMessageFormat = "error deserializing location message type %s"
)

func wrapDeserializeLocationMsgError(err error, msgType string) error {
	return errors.Wrap(err, fmt.Sprintf(errorDeserializeLocationMessageFormat, msgType))
}
