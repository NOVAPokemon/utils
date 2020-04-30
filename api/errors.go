package api

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	errorDecodingJSON = "error decoding json"
)

func wrapJSONDecodingError(err error) error {
	return errors.Wrap(err, fmt.Sprintf(errorDecodingJSON))
}
