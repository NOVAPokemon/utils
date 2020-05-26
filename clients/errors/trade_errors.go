package errors

import "github.com/pkg/errors"

const (
	errorGetTradeLobbies          = "error getting trade lobbies"
	errorCreateTradeLobby         = "error creating trade lobby"
	errorJoinTradeLobby           = "error joining trade lobby"
	errorRejectTradeLobby         = "error rejecting trade lobby"
	errorHandleMessagesTradeLobby = "error handling received messages"
)

func WrapGetTradeLobbiesError(err error) error {
	return errors.Wrap(err, errorGetTradeLobbies)
}

func WrapCreateTradeLobbyError(err error) error {
	return errors.Wrap(err, errorCreateTradeLobby)
}

func WrapJoinTradeLobbyError(err error) error {
	return errors.Wrap(err, errorJoinTradeLobby)
}

func WrapRejectTradeLobbyError(err error) error {
	return errors.Wrap(err, errorRejectTradeLobby)
}

func WrapHandleMessagesTradeError(err error) error {
	return errors.Wrap(err, errorHandleMessagesTradeLobby)
}
