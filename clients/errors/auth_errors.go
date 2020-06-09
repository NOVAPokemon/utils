package errors

import "github.com/pkg/errors"

const (
	errorLogin           = "error logging in"
	errorRegister        = "error registering"
	errorRefreshingToken = "error refreshing auth token"
)

func WrapLoginError(err error) error {
	return errors.Wrap(err, errorLogin)
}

func WrapRegisterError(err error) error {
	return errors.Wrap(err, errorRegister)
}

func WrapRefreshAuthTokenError(err error) error {
	return errors.Wrap(err, errorRefreshingToken)
}
