package clients

import (
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/tokens"
	"net/http"
)

type CaughtPokemonMessage struct {
	Caught bool
}

type GeneratorClient struct {
	GeneratorAddr string
	httpClient    *http.Client
}

func NewGeneratorClient(addr string) *GeneratorClient {
	return &GeneratorClient{
		GeneratorAddr: addr,
		httpClient:    &http.Client{},
	}
}

func (client *GeneratorClient) CatchWildPokemon(authToken string) (caught bool, header http.Header, err error) {
	req, err := BuildRequest("GET", client.GeneratorAddr, api.CatchWildPokemonPath, nil)
	if err != nil {
		return false, nil, err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var msg CaughtPokemonMessage
	resp, err := DoRequest(client.httpClient, req, &msg)
	if err != nil {
		return false, nil, err
	}

	if !msg.Caught {
		return false, nil, nil
	}

	return true, resp.Header, nil
}
