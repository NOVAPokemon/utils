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
	httpClient *http.Client
	conn       *websocket.Conn
}

func NewTradesClient(addr string) *TradeLobbyClient {
	return &TradeLobbyClient{
		TradesAddr: addr,
		httpClient: &http.Client{
		},
	}
}

func (client *TradeLobbyClient) GetAvailableLobbies() []utils.Lobby {
	u := url.URL{Scheme: "http", Host: client.TradesAddr, Path: api.GetTradesPath}

	httpClient := &http.Client{
	}

	resp, err := httpClient.Get(u.String())

	if err != nil {
		log.Error(err)
		return nil
	}

	var battles []utils.Lobby
	err = json.NewDecoder(resp.Body).Decode(&battles)

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
	req.Header.Set(tokens.ItemsTokenTokenName, itemsToken)

	var lobbyIdHex string
	_, err = DoRequest(client.httpClient, req, &lobbyIdHex)
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
	header.Set(tokens.ItemsTokenTokenName, itemsToken)

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

	MainLoop(c, writeChannel, finished)

	log.Info("Finishing...")
}
