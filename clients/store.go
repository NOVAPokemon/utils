package clients

import (
	"fmt"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/tokens"
	"net/http"
)

type StoreClient struct {
	StoreAddr  string
	httpClient *http.Client
}

func NewStoreClient(addr string) *StoreClient {
	return &StoreClient{
		StoreAddr:  addr,
		httpClient: &http.Client{},
	}
}

func (c *StoreClient) GetItems(authToken string) ([]*items.StoreItem, error) {
	req, err := BuildRequest("GET", c.StoreAddr, api.GetShopItemsPath, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var respItems []*items.StoreItem
	_, err = DoRequest(c.httpClient, req, &respItems)
	return respItems, err
}

func (c *StoreClient) BuyItem(itemName, authToken, statsToken string) (string, error) {
	req, err := BuildRequest("POST", c.StoreAddr, fmt.Sprintf(api.BuyItemPath, itemName), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)
	req.Header.Set(tokens.StatsTokenHeaderName, statsToken)

	resp, err := DoRequest(c.httpClient, req, nil)
	if err != nil {
		return "", nil
	}

	return resp.Header.Get(tokens.ItemsTokenHeaderName), err
}
