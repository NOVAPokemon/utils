package clients

import (
	"fmt"
	"net/http"
	"os"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/clients/errors"
	"github.com/NOVAPokemon/utils/tokens"
	"github.com/NOVAPokemon/utils/websockets"
	log "github.com/sirupsen/logrus"
)

type MicrotransactionsClient struct {
	MicrotransactionsAddr string
	httpClient            *http.Client
	commsManager          websockets.CommunicationManager
	*BasicClient
}

var defaultMicrotransactionsURL = fmt.Sprintf("%s:%d", utils.Host, utils.MicrotransactionsPort)

func NewMicrotransactionsClient(manager websockets.CommunicationManager, httpCLient *http.Client,
	client *BasicClient) *MicrotransactionsClient {
	microtransactionsURL, exists := os.LookupEnv(utils.MicrotransactionsEnvVar)

	if !exists {
		log.Warn("missing ", utils.MicrotransactionsEnvVar)
		microtransactionsURL = defaultMicrotransactionsURL
	}

	return &MicrotransactionsClient{
		MicrotransactionsAddr: microtransactionsURL,
		httpClient:            httpCLient,
		commsManager:          manager,
		BasicClient:           client,
	}
}

func (c *MicrotransactionsClient) GetOffers() ([]utils.TransactionTemplate, error) {
	req, err := c.BuildRequest("GET", c.MicrotransactionsAddr, api.GetTransactionOffersPath, nil)
	if err != nil {
		return nil, errors.WrapGetOffersError(err)
	}

	var transactionOffers []utils.TransactionTemplate

	_, err = DoRequest(c.httpClient, req, &transactionOffers, c.commsManager)
	return transactionOffers, errors.WrapGetOffersError(err)
}

func (c *MicrotransactionsClient) GetTransactionRecords(authToken string) ([]utils.TransactionRecord, error) {
	req, err := c.BuildRequest("GET", c.MicrotransactionsAddr, api.GetPerformedTransactionsPath, nil)
	if err != nil {
		return nil, errors.WrapGetTransactionsRecordsError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var transactions []utils.TransactionRecord
	_, err = DoRequest(c.httpClient, req, &transactions, c.commsManager)

	return transactions, errors.WrapGetTransactionsRecordsError(err)
}

func (c *MicrotransactionsClient) PerformTransaction(offerName, authToken, statsToken string) (*string,
	string, error) {
	req, err := c.BuildRequest("POST", c.MicrotransactionsAddr,
		fmt.Sprintf(api.MakeTransactionPath, offerName), nil)
	if err != nil {
		return nil, "", errors.WrapPerformTransactionError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)
	req.Header.Set(tokens.StatsTokenHeaderName, statsToken)

	var transactionId string
	resp, err := DoRequest(c.httpClient, req, &transactionId, c.commsManager)
	if err != nil {
		return nil, "", errors.WrapPerformTransactionError(err)
	}

	return &transactionId, resp.Header.Get(tokens.StatsTokenHeaderName), nil
}
