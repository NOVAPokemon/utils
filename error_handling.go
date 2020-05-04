package utils

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
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

func HandleJSONEncodeError(w *http.ResponseWriter, caller string, err error) {
	log.Warnf("Error encoding in %s response:\n", caller)
	log.Warn(err)
	(*w).WriteHeader(http.StatusInternalServerError)
}

func HandleJSONDecodeError(w *http.ResponseWriter, caller string, err error) {
	err = errors.Wrap(err, "error decoding json")
	log.Warnf("Error decoding body from %s request:\n", caller)
	log.Warn(err)
	(*w).WriteHeader(http.StatusBadRequest)
}

func HandleWrongPasswordError(w *http.ResponseWriter, caller, username, password string) {
	log.Warnf("Error wrong password in %s request:\n", caller)
	log.Warn(fmt.Sprintf("%s : %s", username, password))
	(*w).WriteHeader(http.StatusUnauthorized)
}

func HandleJWTSigningError(w *http.ResponseWriter, caller string, err error) {
	log.Warnf("Error signing jwt in %s request:\n", caller)
	log.Warn(err)
	(*w).WriteHeader(http.StatusInternalServerError)
}

func HandleCookieError(w *http.ResponseWriter, caller string, err error) {
	if err == http.ErrNoCookie {
		log.Warnf("Error no cookie in %s request:\n", caller)
		log.Warn(err)
		(*w).WriteHeader(http.StatusUnauthorized)
		return
	}

	log.Warnf("Error other in %s request:\n", caller)
	log.Warn(err)
	(*w).WriteHeader(http.StatusBadRequest)
	return
}

func HandleToSoonToRefreshError(w *http.ResponseWriter, caller string) {
	log.Warnf("Error too soon to refresh in %s request:\n", caller)
	(*w).WriteHeader(http.StatusBadRequest)
}

func HandleJWTVerifyingError(w *http.ResponseWriter, caller string, err error) {
	if err == jwt.ErrSignatureInvalid {
		log.Warnf("Error invalid signature in %s request:\n", caller)
		log.Warn(err)
		(*w).WriteHeader(http.StatusUnauthorized)
		return
	}

	log.Warnf("Error other in %s request:\n", caller)
	log.Warn(err)
	(*w).WriteHeader(http.StatusBadRequest)
	return
}
