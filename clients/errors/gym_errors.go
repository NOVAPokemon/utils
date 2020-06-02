package errors

import "github.com/pkg/errors"

const (
	errorGetGymInfo = "error getting gym info"
	errorCreateGym  = "error creating gym"
	errorCreateRaid = "error creating raid"
	errorEnterRaid  = "error entering raid"
)

func WrapGetGymInfoError(err error) error {
	return errors.Wrap(err, errorGetGymInfo)
}

func WrapCreateGymError(err error) error {
	return errors.Wrap(err, errorCreateGym)
}

func WrapCreateRaidError(err error) error {
	return errors.Wrap(err, errorCreateRaid)
}

func WrapEnterRaidError(err error) error {
	return errors.Wrap(err, errorEnterRaid)
}
