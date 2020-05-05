package utils

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	// Generic Handler errors
	ErrorInHandlerFormat = "error in %s"

	// Other errors
	errorLoadingConfigs = "error loading configs"
	errorReadingStdin   = "error reading from stdin"
)

var (
	ErrorConnectionUpgrade = errors.New("error upgrading connection")
)

func LogAndSendHTTPError(w *http.ResponseWriter, err error, httpCode int) {
	log.Error(err)
	http.Error(*w, err.Error(), httpCode)
}

func WrapErrorLoadConfigs(err error) error {
	return errors.Wrap(err, errorLoadingConfigs)
}

func WrapErrorReadStdin(err error) error {
	return errors.Wrap(err, errorReadingStdin)
}
