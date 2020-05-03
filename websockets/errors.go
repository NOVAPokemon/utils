package websockets

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	errorParsingMessageFormat = "error parsing message from websocket, %s"
	errorInvalidMsgTypeFormat = "error unsupported msg type %s"
)

var (
	ErrorInvalidMessageFormat = errors.New("error parsing message from websocket")
	ErrorInvalidMessageType   = errors.New("error invalid message type")
)

func WrapMsgParsingError(err error, msgString string) error {
	return errors.Wrap(err, fmt.Sprintf(errorParsingMessageFormat, msgString))
}

func NewInvalidMsgTypeError(msgType string) error {
	return errors.New(fmt.Sprintf(errorInvalidMsgTypeFormat, msgType))
}
