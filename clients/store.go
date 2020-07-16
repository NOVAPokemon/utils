package clients

import (
	"fmt"
	"net/http"
	"os"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/clients/errors"
	"github.com/NOVAPokemon/utils/comms_manager"
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/tokens"
	log "github.com/sirupsen/logrus"
)

type StoreClient struct {
	StoreAddr    string
	httpClient   *http.Client
	commsManager comms_manager.CommunicationManager
}

var defaultStoreURL = fmt.Sprintf("%s:%d", utils.Host, utils.StorePort)

func NewStoreClient(commsManager comms_manager.CommunicationManager) *StoreClient {
	storeURL, exists := os.LookupEnv(utils.StoreEnvVar)

	if !exists {
		log.Warn("missing ", utils.StoreEnvVar)
		storeURL = defaultStoreURL
	}

	return &StoreClient{
		StoreAddr:    storeURL,
		httpClient:   &http.Client{},
		commsManager: commsManager,
	}
}

func (c *StoreClient) GetItems(authToken string) ([]*items.StoreItem, error) {
	req, err := BuildRequest("GET", c.StoreAddr, api.GetShopItemsPath, nil)
	if err != nil {
		return nil, errors.WrapGetItemsError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var respItems []*items.StoreItem
	_, err = DoRequest(c.httpClient, req, &respItems, c.commsManager)
	return respItems, errors.WrapGetItemsError(err)
}

func (c *StoreClient) BuyItem(itemName, authToken, statsToken string) (string, string, error) {
	req, err := BuildRequest("POST", c.StoreAddr, fmt.Sprintf(api.BuyItemPath, itemName), nil)
	if err != nil {
		return "", "", errors.WrapBuyItemError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)
	req.Header.Set(tokens.StatsTokenHeaderName, statsToken)

	resp, err := DoRequest(c.httpClient, req, nil, c.commsManager)
	if err != nil {
		return "", "", errors.WrapBuyItemError(err)
	}

	return resp.Header.Get(tokens.StatsTokenHeaderName), resp.Header.Get(tokens.ItemsTokenHeaderName), nil
}
