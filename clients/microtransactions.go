package clients

import (
	"fmt"
	"net/http"
	"os"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/clients/errors"
	"github.com/NOVAPokemon/utils/tokens"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MicrotransactionsClient struct {
	MicrotransactionsAddr string
	httpClient            *http.Client
	commsManager          utils.CommunicationManager
}

var defaultMicrotransactionsURL = fmt.Sprintf("%s:%d", utils.Host, utils.MicrotransactionsPort)

func NewMicrotransactionsClient(manager utils.CommunicationManager) *MicrotransactionsClient {
	microtransactionsURL, exists := os.LookupEnv(utils.MicrotransactionsEnvVar)

	if !exists {
		log.Warn("missing ", utils.MicrotransactionsEnvVar)
		microtransactionsURL = defaultMicrotransactionsURL
	}

	return &MicrotransactionsClient{
		MicrotransactionsAddr: microtransactionsURL,
		httpClient:            &http.Client{},
		commsManager:          manager,
	}
}

func (c *MicrotransactionsClient) GetOffers() ([]utils.TransactionTemplate, error) {
	req, err := BuildRequest("GET", c.MicrotransactionsAddr, api.GetTransactionOffersPath, nil)
	if err != nil {
		return nil, errors.WrapGetOffersError(err)
	}

	var transactionOffers []utils.TransactionTemplate

	_, err = DoRequest(c.httpClient, req, &transactionOffers, c.commsManager)
	return transactionOffers, errors.WrapGetOffersError(err)
}

func (c *MicrotransactionsClient) GetTransactionRecords(authToken string) ([]utils.TransactionRecord, error) {
	req, err := BuildRequest("GET", c.MicrotransactionsAddr, api.GetPerformedTransactionsPath, nil)
	if err != nil {
		return nil, errors.WrapGetTransactionsRecordsError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var transactions []utils.TransactionRecord
	_, err = DoRequest(c.httpClient, req, &transactions, c.commsManager)

	return transactions, errors.WrapGetTransactionsRecordsError(err)
}

func (c *MicrotransactionsClient) PerformTransaction(offerName, authToken, statsToken string) (*primitive.ObjectID,
	string, error) {
	req, err := BuildRequest("POST", c.MicrotransactionsAddr,
		fmt.Sprintf(api.MakeTransactionPath, offerName), nil)
	if err != nil {
		return nil, "", errors.WrapPerformTransactionError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)
	req.Header.Set(tokens.StatsTokenHeaderName, statsToken)

	transactionId := &primitive.ObjectID{}
	resp, err := DoRequest(c.httpClient, req, transactionId, c.commsManager)
	if err != nil {
		return nil, "", errors.WrapPerformTransactionError(err)
	}

	return transactionId, resp.Header.Get(tokens.StatsTokenHeaderName), nil
}
