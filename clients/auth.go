package clients

import (
	"errors"
	"fmt"
	http "github.com/bruno-anjos/archimedesHTTPClient"
	"os"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	errors2 "github.com/NOVAPokemon/utils/clients/errors"
	"github.com/NOVAPokemon/utils/tokens"
	"github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
)

type AuthClient struct {
	AuthToken    string
	AuthAddr     string
	httpClient   *http.Client
	commsManager websockets.CommunicationManager
}

var defaultAuthURL = fmt.Sprintf("%s:%d", utils.Host, utils.AuthenticationPort)

func NewAuthClient(commsManager websockets.CommunicationManager) *AuthClient {
	authURL, exists := os.LookupEnv(utils.AuthenticationEnvVar)

	if !exists {
		log.Warn("missing ", utils.AuthenticationEnvVar)
		authURL = defaultAuthURL
	}

	return &AuthClient{
		AuthAddr:     authURL,
		httpClient:   &http.Client{},
		commsManager: commsManager,
	}
}

func (client *AuthClient) LoginWithUsernameAndPassword(username, password string) error {
	userJSON := utils.UserJSON{Username: username, Password: password}
	req, err := BuildRequest("POST", client.AuthAddr, api.LoginPath, userJSON)

	resp, err := DoRequest(client.httpClient, req, nil, client.commsManager)
	if err != nil {
		return errors2.WrapLoginError(err)
	}

	client.AuthToken = resp.Header.Get(tokens.AuthTokenHeaderName)

	return nil
}

func (client *AuthClient) Register(username string, password string) error {
	userJSON := utils.UserJSON{Username: username, Password: password}
	req, err := BuildRequest("POST", client.AuthAddr, api.RegisterPath, userJSON)

	resp, err := DoRequest(client.httpClient, req, nil, client.commsManager)
	if err != nil {
		return errors2.WrapRegisterError(err)
	}
	client.AuthToken = resp.Header.Get(tokens.AuthTokenHeaderName)
	return nil
}

func (client *AuthClient) RefreshAuthToken() error {
	req, err := BuildRequest("GET", client.AuthAddr, api.RefreshPath, nil)
	if err != nil {
		return errors2.WrapRefreshAuthTokenError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, client.AuthToken)
	log.Info("Refreshing token....")

	resp, err := DoRequest(client.httpClient, req, nil, client.commsManager)
	if err != nil {
		return errors2.WrapRefreshAuthTokenError(err)
	}

	if resp.Header.Get(tokens.AuthTokenHeaderName) == "" {
		return errors2.WrapRefreshAuthTokenError(errors.New("auth token is an empty string"))
	}
	client.AuthToken = resp.Header.Get(tokens.AuthTokenHeaderName)
	return nil
}
