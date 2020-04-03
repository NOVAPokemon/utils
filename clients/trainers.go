package clients

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/tokens"
	"net/http"
	"strings"
)

type TrainersClient struct {
	TrainersAddr string
	UserAgent    string
	httpClient   *http.Client

	TrainerStatsToken string
	ItemsToken        string
	PokemonTokens     map[string]string
}

// TRAINER

func NewTrainersClient(addr string) *TrainersClient {
	return &TrainersClient{
		TrainersAddr: addr,
		httpClient:   &http.Client{},
	}
}

func (c *TrainersClient) AddTrainer(trainer utils.Trainer) (*utils.Trainer, error) {
	req, err := BuildRequest("POST", c.TrainersAddr, api.AddTrainerPath, trainer)
	if err != nil {
		return nil, err
	}

	var user utils.Trainer

	_, err = DoRequest(c.httpClient, req, &user)

	return &user, err
}

func (c *TrainersClient) ListTrainers() ([]*utils.Trainer, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, api.GetTrainersPath, nil)
	if err != nil {
		return nil, err
	}

	var users []utils.Trainer
	_, err = DoRequest(c.httpClient, req, &users)
	return &users, err
}

func (c *TrainersClient) GetTrainerByUsername(username string) (*utils.Trainer, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GetTrainerByUsernamePath, username), nil)
	if err != nil {
		return nil, err
	}

	var user utils.Trainer
	_, err = DoRequest(c.httpClient, req, &user)
	return &user, err
}

func (c *TrainersClient) UpdateTrainerStats(username string, newStats utils.TrainerStats) (*utils.TrainerStats, error) {
	req, err := BuildRequest("PUT", c.TrainersAddr, fmt.Sprintf(api.UpdateTrainerStatsPath, username), newStats)
	if err != nil {
		return nil, err
	}

	var resultStats utils.TrainerStats
	_, err = DoRequest(c.httpClient, req, &resultStats)
	return &resultStats, err
}

// BAG

func (c *TrainersClient) RemoveItemFromBag(username string, itemId string) error {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.RemoveItemFromBagPath, username, itemId), nil)
	if err != nil {
		return err
	}

	_, err = DoRequest(c.httpClient, req, nil)
	return err
}

func (c *TrainersClient) AddItemToBag(username string, item utils.Item) (*utils.Item, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.AddItemToBagPath, username, ), item)
	if err != nil {
		return nil, err
	}

	var res utils.Item
	_, err = DoRequest(c.httpClient, req, &res)
	return &res, err
}

// POKEMON

func (c *TrainersClient) AddPokemonToTrainer(username string, pokemon utils.Pokemon) (*utils.Pokemon, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.AddPokemonPath, username), pokemon)
	if err != nil {
		return nil, err
	}

	var res utils.Pokemon
	_, err = DoRequest(c.httpClient, req, &res)
	return &res, err
}

func (c *TrainersClient) RemovePokemonFromTrainer(username string, pokemonId string) error {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.RemovePokemonPath, username, pokemonId), nil)
	if err != nil {
		return err
	}

	_, err = DoRequest(c.httpClient, req, nil)
	return err
}

// TOKENS

func (c *TrainersClient) GetAllTrainerTokens(username string) (err error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GenerateAllTokensPath, username), nil)
	if err != nil {
		return err
	}

	resp, err := DoRequest(c.httpClient, req, nil)

	if err != nil {
		return err
	}

	c.TrainerStatsToken = resp.Header.Get(tokens.StatsTokenHeaderName)
	c.ItemsToken = resp.Header.Get(tokens.ItemsTokenTokenName)
	for name, v := range resp.Header {
		if strings.Contains(name, tokens.PokemonsTokenTokenName) {
			split := strings.Split(name, "-")
			c.PokemonTokens[split[len(split)]] = v[0]
		}
	}

	return err
}

func (c *TrainersClient) GetTrainerStatsToken(username string) error {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GenerateTrainerStatsTokenPath, username), nil)
	if err != nil {
		return err
	}
	resp, err := DoRequest(c.httpClient, req, nil)
	if err != nil {
		return err
	}
	c.TrainerStatsToken = resp.Header.Get(tokens.StatsTokenHeaderName)
	return err
}

func (c *TrainersClient) GetPokemonsToken(username string) error {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GeneratePokemonsTokenPath, username), nil)
	if err != nil {
		return err
	}
	resp, err := DoRequest(c.httpClient, req, nil)
	if err != nil {
		return err
	}

	for name, v := range resp.Header {
		if strings.Contains(name, tokens.PokemonsTokenTokenName) {
			split := strings.Split(name, "-")
			c.PokemonTokens[split[len(split)]] = v[0]
		}
	}

	return nil
}

func (c *TrainersClient) GetItemsToken(username string) error {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GenerateItemsTokenPath, username), nil)
	if err != nil {
		return err
	}
	resp, err := DoRequest(c.httpClient, req, nil)
	if err != nil {
		return err
	}
	c.ItemsToken = resp.Header.Get(tokens.ItemsTokenTokenName)
	return err
}

// verifications of tokens

func (c *TrainersClient) VerifyItems(username string, hash []byte) (*bool, error) {
	req, err := BuildRequest("POST", c.TrainersAddr, fmt.Sprintf(api.VerifyItemsPath, username), hash)
	if err != nil {
		return nil, err
	}

	var res bool
	_, err = DoRequest(c.httpClient, req, &res)
	return &res, err
}

func (c *TrainersClient) VerifyPokemons(username string, hashes map[string][]byte) (*bool, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, api.VerifyPokemonsPath, hashes)
	if err != nil {
		return nil, err
	}

	var res bool
	_, err = DoRequest(c.httpClient, req, &res)
	return &res, err
}

func (c *TrainersClient) VerifyTrainerStats(username string, hash []byte) (*bool, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.VerifyTrainerStatsPath, username), hash)
	if err != nil {
		return nil, err
	}
	var res bool
	_, err = DoRequest(c.httpClient, req, &res)
	return &res, err
}
