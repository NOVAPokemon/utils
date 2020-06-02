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
	errorParsingMessage    = "error parsing message from websocket"
	errorHandleRecv        = "error in receive routine"
	errorHandleSend        = "error in sending routine"

	errorInvalidMsgTypeFormat = "error unsupported msg type %s"
	errorDialingMessageFormat = "error dialing %s"

	errorDeserializeMessageFormat = "error deserializing message type %s"
	errorSerializeMessageFormat   = "error serializing message type %s"
)

var (
	ErrorMessageNil           = errors.New("message is nil")
	ErrorInvalidMessageFormat = errors.New("error invalid message format")
	ErrorInvalidMessageType   = errors.New("error invalid message type")

	ErrorTooEarlyToLogEmit    = errors.New("tried logging before setting emitted timestamp")
	ErrorTooEarlyToLogReceive = errors.New("tried logging before setting received timestamp")

	ErrorMsgWasNotEmmitted = errors.New("msg was not emmitted")
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

func WrapDialingError(err error, url string) error {
	return errors.Wrap(err, fmt.Sprintf(errorDialingMessageFormat, url))
}

func WrapSerializeToWSMessageError(err error, msgType string) error {
	return errors.Wrap(err, fmt.Sprintf(errorSerializeMessageFormat, msgType))
}

func wrapDeserializeMsgError(err error, msgType string) error {
	return errors.Wrap(err, fmt.Sprintf(errorDeserializeMessageFormat, msgType))
}

func wrapMsgParsingError(err error) error {
	return errors.Wrap(err, errorParsingMessage)
}

func wrapHandleReceiveError(err error) error {
	return errors.Wrap(err, errorHandleRecv)
}

func wrapHandleSendError(err error) error {
	return errors.Wrap(err, errorHandleSend)
}

// Error builders
func NewInvalidMsgTypeError(msgType string) error {
	return errors.New(fmt.Sprintf(errorInvalidMsgTypeFormat, msgType))
}
