package utils

import (
	originalHttp "net/http"

	http "github.com/bruno-anjos/archimedesHTTPClient"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	// Generic Handler errors
	ErrorInHandlerFormat = "error in %s"

	// Other errors
	errorLoadingConfigs = "error loading configs"
	errorReadingStdin   = "error reading from stdin"
)

func LogAndSendHTTPError(w *originalHttp.ResponseWriter, err error, httpCode int) {
	log.Error(err)
	http.Error(*w, err.Error(), httpCode)
}
func LogWarnAndSendHTTPError(w *originalHttp.ResponseWriter, err error, httpCode int) {
	log.Warn(err)
	http.Error(*w, err.Error(), httpCode)
}

func WrapErrorLoadConfigs(err error) error {
	return errors.Wrap(err, errorLoadingConfigs)
}

func WrapErrorReadStdin(err error) error {
	return errors.Wrap(err, errorReadingStdin)
}
