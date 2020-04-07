package clients

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/tokens"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"strings"
)

type TrainersClient struct {
	TrainersAddr string
	UserAgent    string
	httpClient   *http.Client

	TrainerStatsToken string
	ItemsToken        string
	PokemonTokens     map[string]string

	TrainerStatsClaims *tokens.TrainerStatsToken
	ItemsClaims        *tokens.ItemsToken
	PokemonClaims      map[string]*tokens.PokemonToken
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

	var users []*utils.Trainer
	_, err = DoRequest(c.httpClient, req, &users)
	return users, err
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

func (c *TrainersClient) UpdateTrainerStats(username string, newStats utils.TrainerStats, authToken string) (*utils.TrainerStats, error) {
	req, err := BuildRequest("PUT", c.TrainersAddr, fmt.Sprintf(api.UpdateTrainerStatsPath, username), newStats)
	if err != nil {
		return nil, err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var resultStats utils.TrainerStats
	_, err = DoRequest(c.httpClient, req, &resultStats)
	return &resultStats, err
}

// BAG

func (c *TrainersClient) RemoveItemsFromBag(username string, itemIds []string, authToken string) error {
	var itemIdsPath strings.Builder

	itemIdsPath.WriteString(itemIds[0])
	for i := 1; i < len(itemIds); i++ {
		itemIdsPath.WriteString(",")
		itemIdsPath.WriteString(itemIds[i])
	}

	req, err := BuildRequest("DELETE", c.TrainersAddr,
		fmt.Sprintf(api.RemoveItemFromBagPath, username, itemIdsPath.String()), nil)
	if err != nil {
		return err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	_, err = DoRequest(c.httpClient, req, nil)
	return err
}

func (c *TrainersClient) AddItemsToBag(username string, items []*utils.Item, authToken string) ([]*utils.Item, error) {
	req, err := BuildRequest("POST", c.TrainersAddr, fmt.Sprintf(api.AddItemToBagPath, username), items)
	if err != nil {
		return nil, err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var res []*utils.Item
	_, err = DoRequest(c.httpClient, req, &res)
	return res, err
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

func (c *TrainersClient) UpdateTrainerPokemon(username string, pokemonId string, pokemon utils.Pokemon) (*utils.Pokemon, error) {
	req, err := BuildRequest("PUT", c.TrainersAddr, fmt.Sprintf(api.UpdatePokemonPath, username, pokemonId), pokemon)
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

func (c *TrainersClient) GetAllTrainerTokens(username string, authToken string) (err error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GenerateAllTokensPath, username), nil)
	if err != nil {
		return err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	resp, err := DoRequest(c.httpClient, req, nil)
	if err != nil {
		return err
	}

	// Stats
	if err := c.SetTrainerStatsToken(resp.Header.Get(tokens.StatsTokenHeaderName)); err != nil {
		return err
	}

	// Items
	if err := c.SetItemsToken(resp.Header.Get(tokens.ItemsTokenHeaderName)); err != nil {
		return err
	}

	return c.SetPokemonTokens(resp.Header)
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

	return c.SetTrainerStatsToken(resp.Header.Get(tokens.StatsTokenHeaderName))
}

func (c *TrainersClient) GetPokemonsToken(username string, authToken string) error {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GeneratePokemonsTokenPath, username), nil)
	if err != nil {
		return err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	resp, err := DoRequest(c.httpClient, req, nil)
	if err != nil {
		return err
	}

	return c.SetPokemonTokens(resp.Header)
}

func (c *TrainersClient) GetItemsToken(username, authToken string) error {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GenerateItemsTokenPath, username), nil)
	if err != nil {
		return err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	resp, err := DoRequest(c.httpClient, req, nil)
	if err != nil {
		return err
	}

	return c.SetItemsToken(resp.Header.Get(tokens.ItemsTokenHeaderName))
}

// verifications of tokens

func (c *TrainersClient) VerifyItems(username string, hash []byte, authToken string) (*bool, error) {
	req, err := BuildRequest("POST", c.TrainersAddr, fmt.Sprintf(api.VerifyItemsPath, username), hash)
	if err != nil {
		return nil, err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var res bool
	_, err = DoRequest(c.httpClient, req, &res)
	return &res, err
}

func (c *TrainersClient) VerifyPokemons(username string, hashes map[string][]byte, authToken string) (*bool, error) {
	req, err := BuildRequest("POST", c.TrainersAddr, fmt.Sprintf(api.VerifyPokemonsPath, username), hashes)
	if err != nil {
		return nil, err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var res bool
	_, err = DoRequest(c.httpClient, req, &res)
	return &res, err
}

func (c *TrainersClient) VerifyTrainerStats(username string, hash []byte, authToken string) (*bool, error) {
	req, err := BuildRequest("POST", c.TrainersAddr, fmt.Sprintf(api.VerifyTrainerStatsPath, username), hash)
	if err != nil {
		return nil, err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var res bool
	_, err = DoRequest(c.httpClient, req, &res)
	return &res, err
}

func (c *TrainersClient) SetPokemonTokens(header http.Header) error {
	tkns, ok := header[tokens.PokemonsTokenHeaderName]

	if !ok {
		return errors.New("No pokemon tokens in header")
	}

	auxClaims := make(map[string]*tokens.PokemonToken, len(tkns))
	auxTkns := make(map[string]string, len(tkns))

	i := 0
	added := 0
	for ; i < len(tkns); i++ {

		if len(tkns[i]) == 0 {
			continue
		}

		pokemonClaims, err := tokens.ExtractPokemonToken(tkns[i])
		if err != nil {
			return err
		}

		auxClaims[pokemonClaims.Pokemon.Id.Hex()] = pokemonClaims
		auxTkns[pokemonClaims.Pokemon.Id.Hex()] = tkns[i]

		added++
	}
	logrus.Infof("Trainer has %d pokemmons", added)
	c.PokemonTokens = auxTkns
	c.PokemonClaims = auxClaims

	return nil
}

func (c *TrainersClient) SetTrainerStatsToken(statsToken string) error {
	c.TrainerStatsToken = statsToken

	var err error
	c.TrainerStatsClaims, err = tokens.ExtractStatsToken(c.TrainerStatsToken)
	if err != nil {
		return err
	}

	return nil
}

func (c *TrainersClient) SetItemsToken(itemsToken string) error {
	c.ItemsToken = itemsToken

	var err error
	c.ItemsClaims, err = tokens.ExtractItemsToken(c.ItemsToken)
	if err != nil {
		return err
	}

	return nil
}

// helper methods

func CheckItemsAdded(toAdd, added []*utils.Item) error {
	for i, item := range toAdd {
		item.Id = added[i].Id
	}
	if reflect.DeepEqual(toAdd, added) {
		return nil
	} else {
		return errors.New(fmt.Sprintf("items to add were not successfully added: %+v %+v", toAdd, added))
	}
}

func CheckUpdatedStats(original, updated *utils.TrainerStats) error {
	if reflect.DeepEqual(original, updated) {
		return nil
	} else {
		return errors.New(fmt.Sprintf("stats were not successfully updated: %+v, %+v", original, updated))
	}
}
