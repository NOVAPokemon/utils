package clients

import (
	"encoding/json"
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

type TradeLobbyClient struct {
	TradesAddr string
	httpClient *http.Client
	Jar        *cookiejar.Jar
	conn       *websocket.Conn
}

func NewTradesClient(addr string, jar *cookiejar.Jar) *TradeLobbyClient {
	return &TradeLobbyClient{
		TradesAddr: addr,
		Jar:        jar,
		httpClient: &http.Client{
			Jar: jar,
		},
	}
}

func (client *TradeLobbyClient) GetAvailableLobbies() []utils.Lobby {
	u := url.URL{Scheme: "http", Host: client.TradesAddr, Path: api.GetTradesPath}

	httpClient := &http.Client{
		Jar: client.Jar,
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

func (client *TradeLobbyClient) CreateTradeLobby(username string) *primitive.ObjectID {
	body := api.CreateLobbyRequest{Username: username}
	req, err := BuildRequest("POST", client.TradesAddr, api.StartTradePath, &body)
	if err != nil {
		log.Error(err)
		return nil
	}

	var lobbyIdHex string
	err = DoRequest(client.httpClient, req, &lobbyIdHex)
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

func (client *TradeLobbyClient) JoinTradeLobby(tradeId *primitive.ObjectID) {
	u := url.URL{Scheme: "ws", Host: client.TradesAddr, Path: fmt.Sprintf(api.JoinTradePath, tradeId.Hex())}
	log.Infof("Connecting to: %s", u.String())

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              client.Jar,
	}

	c, _, err := dialer.Dial(u.String(), nil)
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
