package clients

import (
	"errors"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/tokens"
	"math/rand"
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

func (client *GeneratorClient) CatchWildPokemon(authToken, itemsTokenString string) (caught bool, header http.Header, err error) {
	itemsToken, err := tokens.ExtractItemsToken(itemsTokenString)
	if err != nil {
		return false, nil, err
	}

	pokeball, err := getRandomPokeball(itemsToken.Items)
	if err != nil {
		return false, nil, err
	}

	req, err := BuildRequest("GET", client.GeneratorAddr, api.CatchWildPokemonPath, pokeball)
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

func getRandomPokeball(itemsFromToken map[string]items.Item) (*items.Item, error) {
	var pokeballs []*items.Item
	for _, item := range itemsFromToken {
		if item.IsPokeBall() {
			pokeballs = append(pokeballs, &item)
		}
	}

	if pokeballs == nil {
		return nil, errors.New("no pokeballs")
	}

	return pokeballs[rand.Intn(len(pokeballs))], nil
}
