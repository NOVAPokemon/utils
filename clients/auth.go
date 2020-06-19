package clients

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	errors2 "github.com/NOVAPokemon/utils/clients/errors"
	"github.com/NOVAPokemon/utils/tokens"
	log "github.com/sirupsen/logrus"
)

type AuthClient struct {
	AuthToken  string
	AuthAddr   string
	httpClient *http.Client
}

var defaultAuthURL = fmt.Sprintf("%s:%d", utils.Host, utils.AuthenticationPort)

func NewAuthClient() *AuthClient {
	authURL, exists := os.LookupEnv(utils.AuthenticationEnvVar)

	if !exists {
		log.Warn("missing ", utils.AuthenticationEnvVar)
		authURL = defaultAuthURL
	}

	return &AuthClient{
		AuthAddr:   authURL,
		httpClient: &http.Client{},
	}
}

func (client *AuthClient) LoginWithUsernameAndPassword(username, password string) error {
	userJSON := utils.UserJSON{Username: username, Password: password}
	req, err := BuildRequest("POST", client.AuthAddr, api.LoginPath, userJSON)

	resp, err := DoRequest(client.httpClient, req, nil)
	if err != nil {
		return errors2.WrapLoginError(err)
	}

	client.AuthToken = resp.Header.Get(tokens.AuthTokenHeaderName)

	return nil
}

func (client *AuthClient) Register(username string, password string) error {
	userJSON := utils.UserJSON{Username: username, Password: password}
	req, err := BuildRequest("POST", client.AuthAddr, api.RegisterPath, userJSON)

	resp, err := DoRequest(client.httpClient, req, nil)
	if err != nil {
		return errors2.WrapRegisterError(err)
	}
	client.AuthToken = resp.Header.Get(tokens.AuthTokenHeaderName)
	return nil
}

func (client *AuthClient) RefreshAuthToken() error {
	req, err := BuildRequest("GET", client.AuthAddr, api.RefreshPath, nil)
	log.Info("Refreshing token....")
	resp, err := DoRequest(client.httpClient, req, nil)
	if err != nil {
		return errors2.WrapRefreshAuthTokenError(err)
	}

	if resp.Header.Get(tokens.AuthTokenHeaderName) == "" {
		return errors2.WrapRefreshAuthTokenError(errors.New("auth token is an empty string"))
	}
	client.AuthToken = resp.Header.Get(tokens.AuthTokenHeaderName)
	return nil
}
