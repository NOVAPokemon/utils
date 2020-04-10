package clients

import (
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/tokens"
	"net/http"
)

type AuthClient struct {
	AuthToken  string
	AuthAddr   string
	httpClient *http.Client
}

func NewAuthClient(addr string) *AuthClient {
	return &AuthClient{
		AuthAddr:   addr,
		httpClient: &http.Client{},
	}
}

func (client *AuthClient) LoginWithUsernameAndPassword(username, password string) error {
	req, err := BuildRequest("POST", client.AuthAddr, api.LoginPath, utils.UserJSON{Username: username, Password: password})

	resp, err := DoRequest(client.httpClient, req, nil)
	if err != nil {
		return err
	}

	client.AuthToken = resp.Header.Get(tokens.AuthTokenHeaderName)

	return nil
}

func (client *AuthClient) Register(username string, password string) error {
	req, err := BuildRequest("POST", client.AuthAddr, api.RegisterPath, utils.UserJSON{Username: username, Password: password})

	resp, err := DoRequest(client.httpClient, req, nil)
	if err != nil {
		return err
	}

	client.AuthToken = resp.Header.Get(tokens.AuthTokenHeaderName)

	return nil
}
