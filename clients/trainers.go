package clients

import (
	"fmt"
	"os"
	"strings"
	"sync"

	http "github.com/bruno-anjos/archimedesHTTPClient"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/clients/errors"
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/pokemons"
	"github.com/NOVAPokemon/utils/tokens"
	"github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
)

type TrainersClient struct {
	TrainersAddr string
	UserAgent    string
	HttpClient   *http.Client

	TrainerStatsToken string
	ItemsToken        string
	PokemonTokens     map[string]string

	TrainerStatsClaims *tokens.TrainerStatsToken
	ItemsClaims        *tokens.ItemsToken
	PokemonClaims      map[string]tokens.PokemonToken
	ClaimsLock         sync.RWMutex

	commsManager websockets.CommunicationManager
}

var defaultTrainersURL = fmt.Sprintf("%s:%d", utils.Host, utils.TrainersPort)

// TRAINER

func NewTrainersClient(client *http.Client, manager websockets.CommunicationManager) *TrainersClient {
	trainersURL, exists := os.LookupEnv(utils.TrainersEnvVar)

	if !exists {
		log.Warn("missing ", utils.TrainersEnvVar)
		trainersURL = defaultTrainersURL
	}

	log.Debug("trainers url set to ", trainersURL)

	return &TrainersClient{
		TrainersAddr: trainersURL,
		HttpClient:   client,
		ClaimsLock:   sync.RWMutex{},
		commsManager: manager,
	}
}

func (c *TrainersClient) AddTrainer(trainer utils.Trainer) (*utils.Trainer, error) {
	req, err := BuildRequest("POST", c.TrainersAddr, api.AddTrainerPath, trainer)
	if err != nil {
		return nil, errors.WrapAddTrainerError(err)
	}

	var user utils.Trainer
	_, err = DoRequest(c.HttpClient, req, &user, c.commsManager)

	return &user, errors.WrapAddTrainerError(err)
}

func (c *TrainersClient) ListTrainers() ([]*utils.Trainer, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, api.GetTrainersPath, nil)
	if err != nil {
		return nil, errors.WrapListTrainersError(err)
	}

	var users []*utils.Trainer
	_, err = DoRequest(c.HttpClient, req, &users, c.commsManager)

	return users, errors.WrapListTrainersError(err)
}

func (c *TrainersClient) GetTrainerByUsername(username string) (*utils.Trainer, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GetTrainerByUsernamePath, username), nil)
	if err != nil {
		return nil, errors.WrapGetTrainerByUsernameError(err)
	}

	var user utils.Trainer
	_, err = DoRequest(c.HttpClient, req, &user, c.commsManager)

	return &user, errors.WrapGetTrainerByUsernameError(err)
}

func (c *TrainersClient) UpdateTrainerStats(username string, newStats utils.TrainerStats,
	authToken string) (*utils.TrainerStats, error) {
	req, err := BuildRequest("PUT", c.TrainersAddr, fmt.Sprintf(api.UpdateTrainerStatsPath, username), newStats)
	if err != nil {
		return nil, errors.WrapUpdateTrainerStatsError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var resultStats utils.TrainerStats
	resp, err := DoRequest(c.HttpClient, req, &resultStats, c.commsManager)
	if err != nil {
		return nil, errors.WrapUpdateTrainerStatsError(err)
	}

	err = c.SetTrainerStatsToken(resp.Header.Get(tokens.StatsTokenHeaderName))

	return &resultStats, errors.WrapUpdateTrainerStatsError(err)
}

// ITEMS

func (c *TrainersClient) AddItems(username string, itemsToAdd []items.Item,
	authToken string) (map[string]items.Item, error) {
	req, err := BuildRequest("POST", c.TrainersAddr, fmt.Sprintf(api.AddItemToBagPath, username), itemsToAdd)
	if err != nil {
		return nil, errors.WrapAddItemError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var res map[string]items.Item
	resp, err := DoRequest(c.HttpClient, req, &res, c.commsManager)
	if err != nil {
		return nil, errors.WrapAddItemError(err)
	}

	err = c.SetItemsToken(resp.Header.Get(tokens.ItemsTokenHeaderName))

	return res, errors.WrapAddItemError(err)
}

func (c *TrainersClient) RemoveItems(username string, itemIds []string,
	authToken string) (map[string]items.Item, error) {
	var itemIdsPath strings.Builder

	itemIdsPath.WriteString(itemIds[0])
	for i := 1; i < len(itemIds); i++ {
		itemIdsPath.WriteString(",")
		itemIdsPath.WriteString(itemIds[i])
	}

	req, err := BuildRequest("DELETE", c.TrainersAddr,
		fmt.Sprintf(api.RemoveItemFromBagPath, username, itemIdsPath.String()), nil)
	if err != nil {
		return nil, errors.WrapRemoveItemError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var res map[string]items.Item
	resp, err := DoRequest(c.HttpClient, req, &res, c.commsManager)
	if err != nil {
		return nil, errors.WrapRemoveItemError(err)
	}

	err = c.SetItemsToken(resp.Header.Get(tokens.ItemsTokenHeaderName))

	return res, errors.WrapRemoveItemError(err)
}

// POKEMON

func (c *TrainersClient) AddPokemonToTrainer(username string, pokemon pokemons.Pokemon) (*pokemons.Pokemon, error) {
	req, err := BuildRequest("POST", c.TrainersAddr, fmt.Sprintf(api.AddPokemonPath, username), pokemon)
	if err != nil {
		return nil, errors.WrapAddPokemonError(err)
	}

	var res pokemons.Pokemon
	resp, err := DoRequest(c.HttpClient, req, &res, c.commsManager)
	if err != nil {
		return nil, errors.WrapAddPokemonError(err)
	}

	err = c.SetPokemonTokens(resp.Header)

	return &res, errors.WrapAddPokemonError(err)
}

func (c *TrainersClient) UpdateTrainerPokemon(username string, pokemonId string, pokemon pokemons.Pokemon,
	authToken string) (*pokemons.Pokemon, error) {
	req, err := BuildRequest("PUT", c.TrainersAddr, fmt.Sprintf(api.UpdatePokemonPath, username, pokemonId), pokemon)
	if err != nil {
		return nil, errors.WrapUpdatePokemonError(err)
	}

	var res pokemons.Pokemon
	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	resp, err := DoRequest(c.HttpClient, req, &res, c.commsManager)
	if err != nil {
		return nil, errors.WrapUpdatePokemonError(err)
	}

	err = c.SetPokemonTokens(resp.Header)
	return &res, errors.WrapUpdatePokemonError(err)
}

func (c *TrainersClient) RemovePokemonFromTrainer(username string, pokemonId string) (*pokemons.Pokemon,
	error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.RemovePokemonPath, username, pokemonId), nil)
	if err != nil {
		return nil, errors.WrapRemovePokemonError(err)
	}

	// TODO duplicate request no?
	_, err = DoRequest(c.HttpClient, req, nil, c.commsManager)

	var res pokemons.Pokemon
	resp, err := DoRequest(c.HttpClient, req, &res, c.commsManager)
	if err != nil {
		return nil, errors.WrapRemovePokemonError(err)
	}

	err = c.SetPokemonTokens(resp.Header)
	return &res, errors.WrapRemovePokemonError(err)
}

// TOKENS

func (c *TrainersClient) GetAllTrainerTokens(username string, authToken string) (err error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GenerateAllTokensPath, username), nil)
	if err != nil {
		return errors.WrapGetAllTokensError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	resp, err := DoRequest(c.HttpClient, req, nil, c.commsManager)
	if err != nil {
		return errors.WrapGetAllTokensError(err)
	}

	// Stats
	if err = c.SetTrainerStatsToken(resp.Header.Get(tokens.StatsTokenHeaderName)); err != nil {
		return errors.WrapGetAllTokensError(err)
	}

	// ItemId
	if err = c.SetItemsToken(resp.Header.Get(tokens.ItemsTokenHeaderName)); err != nil {
		return errors.WrapGetAllTokensError(err)
	}

	err = c.SetPokemonTokens(resp.Header)
	return errors.WrapGetAllTokensError(err)
}

func (c *TrainersClient) GetTrainerStatsToken(username string, authToken string) error {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GenerateTrainerStatsTokenPath, username), nil)
	if err != nil {
		return errors.WrapGetStatsTokenError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	resp, err := DoRequest(c.HttpClient, req, nil, c.commsManager)
	if err != nil {
		return errors.WrapGetStatsTokenError(err)
	}

	err = c.SetTrainerStatsToken(resp.Header.Get(tokens.StatsTokenHeaderName))
	return errors.WrapGetStatsTokenError(err)
}

func (c *TrainersClient) GetPokemonsToken(username string, authToken string) error {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GeneratePokemonsTokenPath, username), nil)
	if err != nil {
		return errors.WrapGetPokemonTokenError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	resp, err := DoRequest(c.HttpClient, req, nil, c.commsManager)
	if err != nil {
		return errors.WrapGetPokemonTokenError(err)
	}

	err = c.SetPokemonTokens(resp.Header)
	return errors.WrapGetPokemonTokenError(err)
}

func (c *TrainersClient) GetItemsToken(username, authToken string) error {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GenerateItemsTokenPath, username), nil)
	if err != nil {
		return errors.WrapGetItemsTokenError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	resp, err := DoRequest(c.HttpClient, req, nil, c.commsManager)
	if err != nil {
		return errors.WrapGetItemsTokenError(err)
	}

	err = c.SetItemsToken(resp.Header.Get(tokens.ItemsTokenHeaderName))
	return errors.WrapGetItemsTokenError(err)
}

// verifications of tokens

func (c *TrainersClient) VerifyItems(username string, hash string, authToken string) (*bool, error) {
	req, err := BuildRequest("POST", c.TrainersAddr, fmt.Sprintf(api.VerifyItemsPath, username), hash)
	if err != nil {
		return nil, errors.WrapVerifyItemsError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var res bool
	_, err = DoRequest(c.HttpClient, req, &res, c.commsManager)
	return &res, errors.WrapVerifyItemsError(err)
}

func (c *TrainersClient) VerifyPokemons(username string, hashes map[string]string, authToken string) (*bool,
	error) {
	req, err := BuildRequest("POST", c.TrainersAddr, fmt.Sprintf(api.VerifyPokemonsPath, username), hashes)
	if err != nil {
		return nil, errors.WrapVerifyPokemonsError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var res bool
	_, err = DoRequest(c.HttpClient, req, &res, c.commsManager)
	return &res, errors.WrapVerifyPokemonsError(err)
}

func (c *TrainersClient) VerifyTrainerStats(username string, hash string, authToken string) (*bool, error) {
	req, err := BuildRequest("POST", c.TrainersAddr, fmt.Sprintf(api.VerifyTrainerStatsPath, username), hash)
	if err != nil {
		return nil, errors.WrapVerifyStatsError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var res bool
	_, err = DoRequest(c.HttpClient, req, &res, c.commsManager)
	return &res, errors.WrapVerifyStatsError(err)
}

func (c *TrainersClient) SetPokemonTokens(header http.Header) error {
	tkns, ok := header[tokens.PokemonsTokenHeaderName]

	if !ok {
		err := errors.WrapSetPokemonTokensError(tokens.ErrorNoPokemonTokens)
		return err
	}

	auxClaims := make(map[string]tokens.PokemonToken, len(tkns))
	auxTkns := make(map[string]string, len(tkns))

	i := 0
	added := 0
	for ; i < len(tkns); i++ {
		if len(tkns[i]) == 0 {
			continue
		}

		pokemonClaims, err := tokens.ExtractPokemonToken(tkns[i])
		if err != nil {
			return errors.WrapSetPokemonTokensError(err)
		}

		auxClaims[pokemonClaims.Pokemon.Id] = *pokemonClaims
		auxTkns[pokemonClaims.Pokemon.Id] = tkns[i]

		added++
	}
	log.Infof("Trainer has %d pokemons", added)
	c.PokemonTokens = auxTkns

	c.ClaimsLock.Lock()
	c.PokemonClaims = auxClaims
	c.ClaimsLock.Unlock()

	return nil
}

func (c *TrainersClient) AppendPokemonTokens(tkns []string) error {
	c.ClaimsLock.Lock()

	i := 0
	added := 0
	for ; i < len(tkns); i++ {
		if len(tkns) == 0 {
			continue
		}

		pokemonClaims, err := tokens.ExtractPokemonToken(tkns[i])
		if err != nil {
			return errors.WrapAppendPokemonTokensError(err)
		}

		pokemonIdString := pokemonClaims.Pokemon.Id

		c.PokemonClaims[pokemonIdString] = *pokemonClaims
		c.PokemonTokens[pokemonIdString] = tkns[i]

		added++
	}

	c.ClaimsLock.Unlock()

	log.Infof("Added %d pokemons to trainer", added)
	return nil
}

func (c *TrainersClient) SetTrainerStatsToken(statsToken string) error {
	c.TrainerStatsToken = statsToken

	var err error
	c.TrainerStatsClaims, err = tokens.ExtractStatsToken(c.TrainerStatsToken)
	if err != nil {
		return errors.WrapSetStatsTokenError(err)
	}

	return nil
}

func (c *TrainersClient) SetItemsToken(itemsToken string) error {
	c.ItemsToken = itemsToken

	var err error
	c.ItemsClaims, err = tokens.ExtractItemsToken(c.ItemsToken)
	if err != nil {
		return errors.WrapSetItemsTokenError(err)
	}

	return nil
}
