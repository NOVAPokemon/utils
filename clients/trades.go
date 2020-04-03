package clients

import (
	"encoding/json"
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/tokens"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"net/url"
	"time"
)

type TradeLobbyClient struct {
	TradesAddr string

	conn       *websocket.Conn
}

func NewTradesClient(addr string) *TradeLobbyClient {
	return &TradeLobbyClient{
		TradesAddr: addr,
	}
}

func (client *TradeLobbyClient) GetAvailableLobbies(authToken string) []utils.Lobby {
	req, err := BuildRequest("GET", client.TradesAddr, api.GetTradesPath, nil)
	if err != nil {
		log.Error(err)
		return nil
	}

	var battles []utils.Lobby
	_, err = DoRequest(&http.Client{}, req, &battles)
	if err != nil {
		log.Error(err)
		return nil
	}

	return battles
}

func (client *TradeLobbyClient) CreateTradeLobby(username string, authToken string, itemsToken string) *primitive.ObjectID {
	body := api.CreateLobbyRequest{Username: username}
	req, err := BuildRequest("POST", client.TradesAddr, api.StartTradePath, &body)
	if err != nil {
		log.Error(err)
		return nil
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)
	req.Header.Set(tokens.ItemsTokenHeaderName, itemsToken)

	var lobbyIdHex string
	_, err = DoRequest(&http.Client{}, req, &lobbyIdHex)
	if err != nil {
		log.Error(err)
		return nil
	}

	lobbyId, err := primitive.ObjectIDFromHex(lobbyIdHex)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &lobbyId
}

func (client *TradeLobbyClient) JoinTradeLobby(tradeId *primitive.ObjectID, authToken string, itemsToken string) {
	u := url.URL{Scheme: "ws", Host: client.TradesAddr, Path: fmt.Sprintf(api.JoinTradePath, tradeId.Hex())}
	log.Infof("Connecting to: %s", u.String())

	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)
	header.Set(tokens.ItemsTokenHeaderName, itemsToken)

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	c, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()
	client.conn = c

	finished := make(chan struct{})
	writeChannel := make(chan *string)

	go ReadMessages(c, finished)
	go WriteMessage(writeChannel)

	items := tokens.ExtractAndVerifyItemsToken(itemsToken)
	if err != nil {

	}

	go client.autoTrader(items, writeChannel, finished)

	MainLoop(c, writeChannel, finished)

	log.Info("Finishing...")
}

func (client *TradeLobbyClient) autoTrader(items []utils.Item, writeChannel chan *string, finished chan struct{}) {
	items := client.
}
