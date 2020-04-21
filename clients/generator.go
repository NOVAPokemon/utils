package clients

import (
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/pokemons"
	"net/http"
)

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

func (client *GeneratorClient) GetRaidBoss() (*pokemons.Pokemon, error) {
	req, err := BuildRequest("GET", client.GeneratorAddr, api.GenerateRaidBossPath, nil)
	if err != nil {
		return nil, err
	}

	raidBoss := &pokemons.Pokemon{}
	_, err = DoRequest(client.httpClient, req, raidBoss)
	return raidBoss, err
}
