package websockets

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	errorUpgradeConnection = "error upgrading connection to websocket"
	errorClosingConnection = "error closing connection"
	errorWritingMessage    = "error writing message"
	errorReadingMessage    = "error reading message"

	errorParsingMessageFormat = "error parsing message from websocket, %s"
	errorInvalidMsgTypeFormat = "error unsupported msg type %s"
	errorDialingMessageFormat = "error dialing %s"
)

var (
	ErrorInvalidMessageFormat = errors.New("error parsing message from websocket")
	ErrorInvalidMessageType   = errors.New("error invalid message type")
)

// Wrappers
func WrapUpgradeConnectionError(err error) error {
	return errors.Wrap(err, errorUpgradeConnection)
}

func WrapClosingConnectionError(err error) error {
	return errors.Wrap(err, errorClosingConnection)
}

func WrapWritingMessageError(err error) error {
	return errors.Wrap(err, errorWritingMessage)
}

func WrapReadingMessageError(err error) error {
	return errors.Wrap(err, errorReadingMessage)
}

func WrapMsgParsingError(err error, msgString string) error {
	return errors.Wrap(err, fmt.Sprintf(errorParsingMessageFormat, msgString))
}

func WrapDialingError(err error, url string) error {
	return errors.Wrap(err, fmt.Sprintf(errorDialingMessageFormat, url))
}

// Error builders
func NewInvalidMsgTypeError(msgType string) error {
	return errors.New(fmt.Sprintf(errorInvalidMsgTypeFormat, msgType))
}
