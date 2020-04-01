package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type TrainersClient struct {
	TrainersAddr string
	UserAgent    string
	Jar          *cookiejar.Jar
}

// TRAINER

func (c *TrainersClient) ListTrainers() ([]utils.Trainer, error) {
	req, err := c.newRequest("GET", "/users", nil)
	if err != nil {
		return nil, err
	}
	var users []utils.Trainer
	_, err = c.do(req, &users)
	return users, err
}

func (c *TrainersClient) GetTrainerByUsername(username string) (*utils.Trainer, error) {
	req, err := c.newRequest("GET", fmt.Sprintf(api.GetTrainerByUsernamePath, username), nil)
	if err != nil {
		return nil, err
	}

	var user *utils.Trainer
	_, err = c.do(req, user)
	return user, err
}

func (c *TrainersClient) UpdateTrainerStats(username string, newStats utils.TrainerStats) (*utils.TrainerStats, error) {
	req, err := c.newRequest("PUT", fmt.Sprintf(api.UpdateTrainerStatsPath, username), newStats)
	if err != nil {
		return nil, err
	}

	var resultStats *utils.TrainerStats
	_, err = c.do(req, resultStats)
	return resultStats, err
}

// BAG

func (c *TrainersClient) RemoveItemFromBag(username string, itemId string) error {
	req, err := c.newRequest("GET", fmt.Sprintf(api.RemoveItemFromBagPath, username, itemId), nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

func (c *TrainersClient) AddItemToBag(username string, item utils.Item) (*utils.Item, error) {
	req, err := c.newRequest("GET", fmt.Sprintf(api.AddItemToBagPath, username, ), item)
	if err != nil {
		return nil, err
	}
	var res *utils.Item
	_, err = c.do(req, res)
	return res, err
}

// POKEMON

func (c *TrainersClient) AddPokemonToTrainer(username string, pokemon utils.Pokemon) (*utils.Pokemon, error) {
	req, err := c.newRequest("GET", fmt.Sprintf(api.AddPokemonPath, username), pokemon)
	if err != nil {
		return nil, err
	}

	var res *utils.Pokemon
	_, err = c.do(req, res)
	return res, err
}

func (c *TrainersClient) RemovePokemonFromTrainer(username string, pokemonId string) error {
	req, err := c.newRequest("GET", fmt.Sprintf(api.RemovePokemonPath, username, pokemonId), nil)
	if err != nil {
		return err
	}

	var res *utils.Pokemon
	_, err = c.do(req, res)
	return err
}

// TOKENS

func (c *TrainersClient) GetAllTrainerTokens(username string) error {
	req, err := c.newRequest("GET", fmt.Sprintf(api.GenerateAllTokensPath, username), nil)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err
}

func (c *TrainersClient) GetTrainerStatsToken(username string) error {
	req, err := c.newRequest("GET", fmt.Sprintf(api.GenerateTrainerStatsTokenPath, username), nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

func (c *TrainersClient) GetPokemonsToken(username string) error {
	req, err := c.newRequest("GET", fmt.Sprintf(api.GeneratePokemonsTokenPath, username), nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

func (c *TrainersClient) GetItemsToken(username string) error {
	req, err := c.newRequest("GET", fmt.Sprintf(api.GenerateItemsTokenPath, username), nil)
	if err != nil {
		return err
	}
	_, err = c.do(req, nil)
	return err
}

func (c *TrainersClient) VerifyItems(username string, hash []byte) (*bool, error) {
	req, err := c.newRequest("GET", fmt.Sprintf(api.VerifyItemsPath, username), hash)
	if err != nil {
		return nil, err
	}
	var res *bool
	_, err = c.do(req, res)
	return res, err
}

func (c *TrainersClient) VerifyPokemon(username string, pokemonId string, hash []byte) (*bool, error) {
	req, err := c.newRequest("GET", fmt.Sprintf(api.VerifyPokemonPath, username, pokemonId), hash)
	if err != nil {
		return nil, err
	}
	var res *bool
	_, err = c.do(req, res)
	return res, err
}

func (c *TrainersClient) VerifyTrainerStats(username string, hash []byte) (*bool, error) {
	req, err := c.newRequest("GET", fmt.Sprintf(api.VerifyTrainerStatsPath, username), hash)
	if err != nil {
		return nil, err
	}
	var res *bool
	_, err = c.do(req, res)
	return res, err
}

// helper functions

func (c *TrainersClient) newRequest(method, path string, body interface{}) (*http.Request, error) {
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	u := url.URL{Scheme: "http", Host: c.TrainersAddr, Path: "/battles"}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	return req, nil
}

func (c *TrainersClient) do(req *http.Request, v interface{}, ) (*http.Response, error) {

	httpClient := &http.Client{
		Jar: c.Jar,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(v)
	return resp, err
}
