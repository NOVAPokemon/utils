package user

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	errorGetAllUsers    = "error getting all users from db"
	errorDeleteAllUsers = "error deleting all users from db"

	errorAddUserFormat         = "error adding user %s to db"
	errorCheckUserExistsFormat = "error checking if user %s existed in db"
	errorGetUserFormat         = "error getting user %s from db"
	errorUpdateUserFormat      = "error updating user %s in db"
	errorDeleteUserFormat      = "error deleting user %s from db"
)

func wrapAddUserError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorAddUserFormat, username))
}

func wrapUserExistsError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorCheckUserExistsFormat, username))
}

func wrapGetUserError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorGetUserFormat, username))
}

func wrapGetAllUsersError(err error) error {
	return errors.Wrap(err, errorGetAllUsers)
}

func wrapUpdateUserError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorUpdateUserFormat, username))
}

func wrapDeleteUserError(err error, username string) error {
	return errors.Wrap(err, fmt.Sprintf(errorDeleteUserFormat, username))
}

func wrapDeleteAllUsersError(err error) error {
	return errors.Wrap(err, errorDeleteAllUsers)
}
