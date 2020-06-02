package trades

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	errorDeserializeTradeMessageFormat = "error deserializing trade message type %s"
)

func wrapDeserializeTradeMsgError(err error, msgType string) error {
	return errors.Wrap(err, fmt.Sprintf(errorDeserializeTradeMessageFormat, msgType))
}
