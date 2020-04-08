package clients

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/tokens"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type MicrotransactionsClient struct {
	MicrotransactionsAddr string
	httpClient            *http.Client
}

func NewMicrotransactionsClient(addr string) *MicrotransactionsClient {
	return &MicrotransactionsClient{
		MicrotransactionsAddr: addr,
		httpClient:            &http.Client{},
	}
}

func (c *MicrotransactionsClient) GetOffers() ([]utils.TransactionTemplate, error) {
	req, err := BuildRequest("GET", c.MicrotransactionsAddr, api.GetTransactionOffersRoute, nil)
	if err != nil {
		return nil, err
	}

	var transactionOffers []utils.TransactionTemplate
	_, err = DoRequest(c.httpClient, req, &transactionOffers)
	return transactionOffers, err
}

func (c *MicrotransactionsClient) GetTransactionRecords(authToken string) ([]utils.TransactionRecord, error) {
	req, err := BuildRequest("GET", c.MicrotransactionsAddr, api.GetPerformedTransactionsPath, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var transactions []utils.TransactionRecord
	_, err = DoRequest(c.httpClient, req, &transactions)

	return transactions, err
}

func (c *MicrotransactionsClient) PerformTransaction(offerName, authToken, statsToken string) (*primitive.ObjectID, string, error) {
	req, err := BuildRequest("POST", c.MicrotransactionsAddr, fmt.Sprintf(api.MakeTransactionPath, offerName), nil)
	if err != nil {
		return nil, "", err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)
	req.Header.Set(tokens.StatsTokenHeaderName, statsToken)

	transactionId := &primitive.ObjectID{}
	resp, err := DoRequest(c.httpClient, req, transactionId)
	if err != nil {
		return nil, "", nil
	}

	return transactionId, resp.Header.Get(tokens.StatsTokenHeaderName), err
}
