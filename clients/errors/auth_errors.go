package errors

import "github.com/pkg/errors"

const (
	errorLogin    = "error logging in"
	errorRegister = "error registering"
)

func WrapLoginError(err error) error {
	return errors.Wrap(err, errorLogin)
}

func WrapRegisterError(err error) error {
	return errors.Wrap(err, errorRegister)
}
