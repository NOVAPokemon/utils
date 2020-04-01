package clients

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"net/http"
	"net/http/cookiejar"
)

type TrainersClient struct {
	TrainersAddr string
	UserAgent    string
	httpClient   *http.Client
	Jar          *cookiejar.Jar
}

// TRAINER

func NewTrainersClient(addr string, jar *cookiejar.Jar) *TrainersClient {
	return &TrainersClient{
		TrainersAddr: addr,
		Jar:          jar,
		httpClient: &http.Client{
			Jar: jar,
		},
	}
}

func (c *TrainersClient) AddTrainer(trainer utils.Trainer) (*utils.Trainer, error) {
	req, err := BuildRequest("POST", c.TrainersAddr, api.AddTrainerPath, trainer)
	if err != nil {
		return nil, err
	}

	var user utils.Trainer

	err = DoRequest(c.httpClient, req, &user)

	return &user, err
}

func (c *TrainersClient) ListTrainers() (*[]utils.Trainer, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, api.GetTrainersPath, nil)
	if err != nil {
		return nil, err
	}

	var users []utils.Trainer
	err = DoRequest(c.httpClient, req, &users)
	return &users, err
}

func (c *TrainersClient) GetTrainerByUsername(username string) (*utils.Trainer, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GetTrainerByUsernamePath, username), nil)
	if err != nil {
		return nil, err
	}

	var user utils.Trainer
	err = DoRequest(c.httpClient, req, &user)
	return &user, err
}

func (c *TrainersClient) UpdateTrainerStats(username string, newStats utils.TrainerStats) (*utils.TrainerStats, error) {
	req, err := BuildRequest("PUT", c.TrainersAddr, fmt.Sprintf(api.UpdateTrainerStatsPath, username), newStats)
	if err != nil {
		return nil, err
	}

	var resultStats utils.TrainerStats
	err = DoRequest(c.httpClient, req, &resultStats)
	return &resultStats, err
}

// BAG

func (c *TrainersClient) RemoveItemFromBag(username string, itemId string) error {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.RemoveItemFromBagPath, username, itemId), nil)
	if err != nil {
		return err
	}

	err = DoRequest(c.httpClient, req, nil)
	return err
}

func (c *TrainersClient) AddItemToBag(username string, item utils.Item) (*utils.Item, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.AddItemToBagPath, username, ), item)
	if err != nil {
		return nil, err
	}

	var res utils.Item
	err = DoRequest(c.httpClient, req, &res)
	return &res, err
}

// POKEMON

func (c *TrainersClient) AddPokemonToTrainer(username string, pokemon utils.Pokemon) (*utils.Pokemon, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.AddPokemonPath, username), pokemon)
	if err != nil {
		return nil, err
	}

	var res utils.Pokemon
	err = DoRequest(c.httpClient, req, &res)
	return &res, err
}

func (c *TrainersClient) RemovePokemonFromTrainer(username string, pokemonId string) error {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.RemovePokemonPath, username, pokemonId), nil)
	if err != nil {
		return err
	}

	err = DoRequest(c.httpClient, req, nil)
	return err
}

// TOKENS

func (c *TrainersClient) GetAllTrainerTokens(username string) error {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GenerateAllTokensPath, username), nil)
	if err != nil {
		return err
	}

	err = DoRequest(c.httpClient, req, nil)
	return err
}

func (c *TrainersClient) GetTrainerStatsToken(username string) error {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GenerateTrainerStatsTokenPath, username), nil)
	if err != nil {
		return err
	}
	err = DoRequest(c.httpClient, req, nil)
	return err
}

func (c *TrainersClient) GetPokemonsToken(username string) error {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GeneratePokemonsTokenPath, username), nil)
	if err != nil {
		return err
	}
	err = DoRequest(c.httpClient, req, nil)
	return err
}

func (c *TrainersClient) GetItemsToken(username string) error {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GenerateItemsTokenPath, username), nil)
	if err != nil {
		return err
	}

	err = DoRequest(c.httpClient, req, nil)
	return err
}

// verifications of tokens

func (c *TrainersClient) VerifyItems(username string, hash []byte) (*bool, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.VerifyItemsPath, username), hash)
	if err != nil {
		return nil, err
	}

	var res bool
	err = DoRequest(c.httpClient, req, &res)
	return &res, err
}

func (c *TrainersClient) VerifyPokemon(username string, pokemonId string, hash []byte) (*bool, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.VerifyPokemonPath, username, pokemonId), hash)
	if err != nil {
		return nil, err
	}

	var res bool
	err = DoRequest(c.httpClient, req, &res)
	return &res, err
}

func (c *TrainersClient) VerifyTrainerStats(username string, hash []byte) (*bool, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.VerifyTrainerStatsPath, username), hash)
	if err != nil {
		return nil, err
	}
	var res bool
	err = DoRequest(c.httpClient, req, &res)
	return &res, err
}
