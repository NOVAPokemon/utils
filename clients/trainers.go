package clients

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/pokemons"
	"github.com/NOVAPokemon/utils/tokens"
	"github.com/NOVAPokemon/utils/websockets/location"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
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
	PokemonClaims      map[string]*tokens.PokemonToken
}

// TRAINER

func NewTrainersClient(addr string, client *http.Client) *TrainersClient {
	return &TrainersClient{
		TrainersAddr: addr,
		HttpClient:   client,
	}
}

func (c *TrainersClient) AddTrainer(trainer utils.Trainer) (*utils.Trainer, error) {
	req, err := BuildRequest("POST", c.TrainersAddr, api.AddTrainerPath, trainer)
	if err != nil {
		return nil, err
	}

	var user utils.Trainer

	_, err = DoRequest(c.HttpClient, req, &user)

	return &user, err
}

func (c *TrainersClient) ListTrainers() ([]*utils.Trainer, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, api.GetTrainersPath, nil)
	if err != nil {
		return nil, err
	}

	var users []*utils.Trainer
	_, err = DoRequest(c.HttpClient, req, &users)
	return users, err
}

func (c *TrainersClient) GetTrainerByUsername(username string) (*utils.Trainer, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GetTrainerByUsernamePath, username), nil)
	if err != nil {
		return nil, err
	}

	var user utils.Trainer
	_, err = DoRequest(c.HttpClient, req, &user)
	return &user, err
}

func (c *TrainersClient) UpdateTrainerStats(username string, newStats utils.TrainerStats, authToken string) (*utils.TrainerStats, error) {
	req, err := BuildRequest("PUT", c.TrainersAddr, fmt.Sprintf(api.UpdateTrainerStatsPath, username), newStats)
	if err != nil {
		return nil, err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var resultStats utils.TrainerStats
	resp, err := DoRequest(c.HttpClient, req, &resultStats)

	if err != nil {
		return nil, err
	}

	if err := c.SetTrainerStatsToken(resp.Header.Get(tokens.StatsTokenHeaderName)); err != nil {
		return nil, err
	}

	return &resultStats, err
}

// BAG

func (c *TrainersClient) RemoveItemsFromBag(username string, itemIds []string, authToken string) (map[string]items.Item, error) {
	var itemIdsPath strings.Builder

	itemIdsPath.WriteString(itemIds[0])
	for i := 1; i < len(itemIds); i++ {
		itemIdsPath.WriteString(",")
		itemIdsPath.WriteString(itemIds[i])
	}

	req, err := BuildRequest("DELETE", c.TrainersAddr,
		fmt.Sprintf(api.RemoveItemFromBagPath, username, itemIdsPath.String()), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var res map[string]items.Item
	resp, err := DoRequest(c.HttpClient, req, &res)

	if err != nil {
		return nil, err
	}

	if err := c.SetItemsToken(resp.Header.Get(tokens.ItemsTokenHeaderName)); err != nil {
		return nil, err
	}

	return res, nil
}

func (c *TrainersClient) AddItemsToBag(username string, itemsToAdd []items.Item, authToken string) (map[string]items.Item, error) {
	req, err := BuildRequest("POST", c.TrainersAddr, fmt.Sprintf(api.AddItemToBagPath, username), itemsToAdd)
	if err != nil {
		return nil, err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var res map[string]items.Item
	resp, err := DoRequest(c.HttpClient, req, &res)

	if err != nil {
		return nil, err
	}

	if err := c.SetItemsToken(resp.Header.Get(tokens.ItemsTokenHeaderName)); err != nil {
		return nil, err
	}

	return res, err
}

// POKEMON

func (c *TrainersClient) AddPokemonToTrainer(username string, pokemon pokemons.Pokemon) (map[string]pokemons.Pokemon, error) {
	req, err := BuildRequest("POST", c.TrainersAddr, fmt.Sprintf(api.AddPokemonPath, username), pokemon)
	if err != nil {
		return nil, err
	}

	var res map[string]pokemons.Pokemon
	resp, err := DoRequest(c.HttpClient, req, &res)

	if err != nil {
		return nil, err
	}

	if err := c.SetPokemonTokens(resp.Header); err != nil {
		return nil, err
	}
	return res, err
}

func (c *TrainersClient) UpdateTrainerPokemon(username string, pokemonId string, pokemon pokemons.Pokemon) (map[string]pokemons.Pokemon, error) {
	req, err := BuildRequest("PUT", c.TrainersAddr, fmt.Sprintf(api.UpdatePokemonPath, username, pokemonId), pokemon)
	if err != nil {
		return nil, err
	}

	var res map[string]pokemons.Pokemon
	resp, err := DoRequest(c.HttpClient, req, &res)

	if err != nil {
		return nil, err
	}

	if err := c.SetPokemonTokens(resp.Header); err != nil {
		return nil, err
	}

	return res, err
}

func (c *TrainersClient) RemovePokemonFromTrainer(username string, pokemonId string) (map[string]pokemons.Pokemon, error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.RemovePokemonPath, username, pokemonId), nil)
	if err != nil {
		return nil, err
	}

	_, err = DoRequest(c.HttpClient, req, nil)

	var res map[string]pokemons.Pokemon
	resp, err := DoRequest(c.HttpClient, req, &res)

	if err != nil {
		return nil, err
	}

	if err := c.SetPokemonTokens(resp.Header); err != nil {
		return nil, err
	}

	return res, err
}

// TOKENS

func (c *TrainersClient) GetAllTrainerTokens(username string, authToken string) (err error) {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GenerateAllTokensPath, username), nil)
	if err != nil {
		return err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	resp, err := DoRequest(c.HttpClient, req, nil)
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

func (c *TrainersClient) GetTrainerStatsToken(username string, authToken string) error {
	req, err := BuildRequest("GET", c.TrainersAddr, fmt.Sprintf(api.GenerateTrainerStatsTokenPath, username), nil)
	if err != nil {
		return err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	resp, err := DoRequest(c.HttpClient, req, nil)
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

	resp, err := DoRequest(c.HttpClient, req, nil)
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

	resp, err := DoRequest(c.HttpClient, req, nil)
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
	_, err = DoRequest(c.HttpClient, req, &res)
	return &res, err
}

func (c *TrainersClient) VerifyPokemons(username string, hashes map[string][]byte, authToken string) (*bool, error) {
	req, err := BuildRequest("POST", c.TrainersAddr, fmt.Sprintf(api.VerifyPokemonsPath, username), hashes)
	if err != nil {
		return nil, err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var res bool
	_, err = DoRequest(c.HttpClient, req, &res)
	return &res, err
}

func (c *TrainersClient) VerifyTrainerStats(username string, hash []byte, authToken string) (*bool, error) {
	req, err := BuildRequest("POST", c.TrainersAddr, fmt.Sprintf(api.VerifyTrainerStatsPath, username), hash)
	if err != nil {
		return nil, err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var res bool
	_, err = DoRequest(c.HttpClient, req, &res)
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
	log.Infof("Trainer has %d pokemons", added)
	c.PokemonTokens = auxTkns
	c.PokemonClaims = auxClaims

	return nil
}

func (c *TrainersClient) AppendPokemonToken(header http.Header) error {
	tkns, ok := header[tokens.PokemonsTokenHeaderName]

	if !ok {
		return errors.New("No pokemon tokens in header")
	}

	i := 0
	added := 0
	for ; i < len(tkns); i++ {
		if len(tkns[i]) == 0 {
			continue
		}

		log.Info(tkns[i])

		pokemonClaims, err := tokens.ExtractPokemonToken(tkns[i])
		if err != nil {
			return err
		}

		pokemonIdString := pokemonClaims.Pokemon.Id.Hex()

		c.PokemonClaims[pokemonIdString] = pokemonClaims
		c.PokemonTokens[pokemonIdString] = tkns[i]

		added++
	}

	log.Infof("Added %d pokemons to trainer", added)
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

func (c *TrainersClient) StartLocationUpdates(authToken string) {
	u := url.URL{Scheme: "ws", Host: c.TrainersAddr, Path: fmt.Sprintf(api.UpdateRegionPath)}
	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)
	writeMut := sync.Mutex{}

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	log.Info("Dialing: ", u.String())
	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(location.Timeout))
	conn.SetPingHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(location.Timeout))
		writeMut.Lock()
		_ = conn.WriteMessage(websocket.PongMessage, nil)
		writeMut.Unlock()
		return nil
	})

	var updateTicker = time.NewTicker(location.UpdateCooldown)
	var inChan = make(chan string)
	finish := make(chan *struct{})
	go func() {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Error(err)
			finish <- nil
			return
		} else {
			inChan <- string(msg)
		}
	}()

	for {
		select {
		case <-updateTicker.C:
			//log.Info("Updating location...")
			loc, err := c.getLocation()
			if err != nil {
				log.Error(err)
				return
			}
			writeMut.Lock()
			err = conn.WriteJSON(*loc)
			writeMut.Unlock()
			if err != nil {
				log.Error(err)
				return
			}

			if err != nil {
				log.Error(err)
				return
			}
			_ = conn.SetReadDeadline(time.Now().Add(location.Timeout))
		case msg := <-inChan:
			log.Info(msg)
		case <-finish:
			log.Warn("Trainer stopped updating location")
		}

	}
}

func (c *TrainersClient) getLocation() (*utils.Location, error) {
	return &utils.Location{
		Latitude: rand.Float64(), Longitude: rand.Float64(),
	}, nil // TODO
}

/*
// helper methods

func CheckItemsAdded(original, new map[string]items.Item, added []items.Item) error {
	for i, item := range toAdd {
		item.Id = added[i].Id
	}
	if reflect.DeepEqual(toAdd, added) {
		return nil
	} else {
		return errors.New(fmt.Sprintf("items to add were not successfully added: %+v %+v", toAdd, added))
	}
}

*/

func CheckUpdatedStats(original, updated *utils.TrainerStats) error {
	if original.XP == updated.XP && original.Coins == updated.Coins {
		return nil
	} else {
		return errors.New(fmt.Sprintf("stats were not successfully updated: %+v, %+v", original, updated))
	}
}
