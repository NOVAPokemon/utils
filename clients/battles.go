package clients

import (
	"encoding/json"
	"github.com/NOVAPokemon/utils"
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

func (client *BattleLobbyClient) CreateBattleLobby() {

	u := url.URL{Scheme: "ws", Host: client.BattlesAddr, Path: "/battles/join"}
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

	defer c.Close()
	client.conn = c

	inChannel := make(chan *string)
	finished := make(chan struct{})
	go WriteMessage(inChannel)
	go ReadMessages(c, finished)

	MainLoop(c, inChannel, finished)

}

func (client *BattleLobbyClient) JoinBattleLobby(battleId primitive.ObjectID) {

	u := url.URL{Scheme: "ws", Host: client.BattlesAddr, Path: "/battles/join/" + battleId.Hex()}
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

	defer c.Close()
	client.conn = c

	inChannel := make(chan *string)
	finished := make(chan struct{})

	go WriteMessage(inChannel)
	go ReadMessages(c, finished)

	MainLoop(c, inChannel, finished)

}
