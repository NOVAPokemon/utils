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

	errorInvalidMsgTypeFormat = "error unsupported msg type %s"
	errorDialingMessageFormat = "error dialing %s"

	errorLobbyFull                = "lobby %s is full"
	errorLobbyStarted             = "lobby %s already started"
	errorLobbyFinished            = "lobby %s already finished"
)

var (
	ErrorInvalidMessageType   = errors.New("error invalid message type")

	ErrorTooEarlyToLogEmit    = errors.New("tried logging before setting emitted timestamp")
	ErrorTooEarlyToLogReceive = errors.New("tried logging before setting received timestamp")

	ErrorMsgWasNotEmmitted    = errors.New("msg was not emmitted")
	ErrorLobbyIsFull          = errors.New("lobby is full")
	ErrorLobbyAlreadyFinished = errors.New("lobby finished")
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

func wrapMsgParsingError(err error) error {
	return errors.Wrap(err, errorParsingMessage)
}

// Error builders
func NewInvalidMsgTypeError(msgType string) error {
	return errors.New(fmt.Sprintf(errorInvalidMsgTypeFormat, msgType))
}

func NewLobbyIsFullError(lobbyId string) error {
	return errors.WithMessage(ErrorLobbyIsFull, fmt.Sprintf(errorLobbyFull, lobbyId))
}

func NewLobbyStartedError(lobbyId string) error {
	return errors.New(fmt.Sprintf(errorLobbyStarted, lobbyId))
}

func NewLobbyFinishedError(lobbyId string) error {
	return errors.WithMessage(ErrorLobbyAlreadyFinished, fmt.Sprintf(errorLobbyFinished, lobbyId))
}
