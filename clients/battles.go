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

type BattleLobbyClient struct {
	BattlesAddr string
	Jar         *cookiejar.Jar
	conn        *websocket.Conn
}

type BattleChannels struct {
	Channel       chan *string
	FinishChannel chan struct{}
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

func (client *BattleLobbyClient) QueueForBattle() BattleChannels {

	u := url.URL{Scheme: "ws", Host: client.BattlesAddr, Path: api.QueueForBattlePath}
	log.Infof("Queuing for battle: %s", u.String())

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              client.Jar,
	}

	c, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	channel := make(chan *string)
	finished := make(chan struct{})

	go ReadMessages(c, finished)
	go MainLoop(c, channel, finished)

	return BattleChannels{channel, finished}
}

func (client *BattleLobbyClient) ChallengePlayerToBattle(targetPlayer string) BattleChannels {

	u := url.URL{Scheme: "ws", Host: client.BattlesAddr, Path: fmt.Sprintf(api.ChallengeToBattlePath, targetPlayer)}
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

	channel := make(chan *string)
	finished := make(chan struct{})

	go ReadMessages(c, finished)
	go MainLoop(c, channel, finished)

	return BattleChannels{channel, finished}

}

func (client *BattleLobbyClient) AcceptChallenge(battleId primitive.ObjectID) BattleChannels {

	u := url.URL{Scheme: "ws", Host: client.BattlesAddr, Path: fmt.Sprintf(api.AcceptChallengePath, battleId.Hex())}
	log.Infof("Accepting challenge: %s", u.String())

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              client.Jar,
	}

	c, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	channel := make(chan *string)
	finished := make(chan struct{})

	go ReadMessages(c, finished)
	go MainLoop(c, channel, finished)

	return BattleChannels{channel, finished}

}
