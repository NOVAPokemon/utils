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

type BattleLobbyClient struct {
	BattlesAddr string
	conn        *websocket.Conn
	httpClient  http.Client
}

type BattleChannels struct {
	OutChannel    chan *string
	InChannel     chan *string
	FinishChannel chan struct{}
}

func init() {

}

func (client *BattleLobbyClient) GetAvailableLobbies() []utils.Lobby {

	u := url.URL{Scheme: "http", Host: client.BattlesAddr, Path: "/battles"}

	resp, err := http.Get(u.String())

	if err != nil {
		log.Error(err)
		return nil
	}

	var availableBattles []utils.Lobby
	err = json.NewDecoder(resp.Body).Decode(&availableBattles)

	if err != nil {
		log.Error(err)
		return nil
	}

	return availableBattles
}

func (client *BattleLobbyClient) QueueForBattle(authToken string, pokemonsTokens []string) (*BattleChannels, error) {

	u := url.URL{Scheme: "ws", Host: client.BattlesAddr, Path: api.QueueForBattlePath}
	log.Infof("Queuing for battle: %s", u.String())

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)
	header[tokens.PokemonsTokenHeaderName] = pokemonsTokens

	c, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	outChannel := make(chan *string)
	inChannel := make(chan *string)
	finished := make(chan struct{})

	go ReadMessagesToChan(c, inChannel, finished)
	go MainLoop(c, outChannel, finished)

	return &BattleChannels{outChannel, inChannel, finished}, nil
}

func (client *BattleLobbyClient) ChallengePlayerToBattle(authToken string, pokemonsTokens []string, targetPlayer string) (*BattleChannels, error) {

	u := url.URL{Scheme: "ws", Host: client.BattlesAddr, Path: fmt.Sprintf(api.ChallengeToBattlePath, targetPlayer)}
	log.Infof("Connecting to: %s", u.String())

	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)
	header[tokens.PokemonsTokenHeaderName] = pokemonsTokens

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	c, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		log.Fatal(err)
	}

	outChannel := make(chan *string)
	inChannel := make(chan *string)
	finished := make(chan struct{})

	go ReadMessagesToChan(c, inChannel, finished)
	go MainLoop(c, outChannel, finished)

	return &BattleChannels{outChannel, inChannel, finished}, nil

}

func (client *BattleLobbyClient) AcceptChallenge(authToken string, pokemonsTokens []string, battleId primitive.ObjectID) (*BattleChannels, error) {

	u := url.URL{Scheme: "ws", Host: client.BattlesAddr, Path: fmt.Sprintf(api.AcceptChallengePath, battleId.Hex())}
	log.Infof("Accepting challenge: %s", u.String())

	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)
	header[tokens.PokemonsTokenHeaderName] = pokemonsTokens

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	c, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	outChannel := make(chan *string)
	inChannel := make(chan *string)
	finished := make(chan struct{})

	go ReadMessagesToChan(c, inChannel, finished)
	go MainLoop(c, outChannel, finished)

	return &BattleChannels{outChannel, inChannel, finished}, nil

}
