package clients

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	errorMainLoop           = "error in main loop"
	errorBuildRequest       = "error building request"
	errorDoRequest          = "error doing request"
	errorHttpClietNilFormat = "httpclient is nil for: %s"
)

// Wrappers

func wrapBuildRequestError(err error) error {
	return errors.Wrap(err, errorBuildRequest)
}

func wrapDoRequestError(err error) error {
	return errors.Wrap(err, errorDoRequest)
}

// Error builders
func newHttpClientNilError(url string) error {
	return errors.New(fmt.Sprintf(errorHttpClietNilFormat, url))
}

func newUnexpectedResponseStatusError(statusCode int) error {
	return errors.New(fmt.Sprintf("got status code %d", statusCode))
}
