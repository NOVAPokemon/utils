package clients

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	locationMessages "github.com/NOVAPokemon/utils/messages/location"
	"github.com/NOVAPokemon/utils/tokens"
	"github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/location"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

const (
	bufferSize = 10
)

type LocationClient struct {
	LocationAddr string
	Gyms         []utils.Gym
	HttpClient   *http.Client
}

func NewLocationClient(addr string) *LocationClient {
	return &LocationClient{
		LocationAddr: addr,
		HttpClient:   &http.Client{},
	}
}

func (c *LocationClient) StartLocationUpdates(locationParameters utils.LocationParameters, authToken string) {
	inChan := make(chan *websockets.Message)
	outChan := make(chan websockets.GenericMsg, bufferSize)
	finish := make(chan struct{})

	conn, err := c.connect(outChan, authToken)
	if err != nil {
		log.Error(err)
		return
	}

	go readMessages(conn, inChan, finish)
	go updateLocation(locationParameters, conn, outChan)

	for {
		select {
		case msg := <-inChan:
			switch msg.MsgType {
			case location.Gyms:
				log.Info("updating gyms")
				log.Info(locationMessages.Deserialize(msg).(locationMessages.GymsMessage).Gyms)
			default:
				log.Warn("got message type ", msg.MsgType)
			}
		case <-finish:
			log.Warn("Trainer stopped updating location")
			return
		case msg := <-outChan:
			if err := conn.WriteMessage(msg.MsgType, msg.Data); err != nil {
				log.Error(err)
				return
			}
		}
	}
}

func (c *LocationClient) connect(outChan chan websockets.GenericMsg, authToken string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: c.LocationAddr, Path: fmt.Sprintf(api.UserLocationPath)}
	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	log.Info("Dialing: ", u.String())
	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	_ = conn.SetReadDeadline(time.Now().Add(location.Timeout))
	conn.SetPingHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(location.Timeout)); err != nil {
			return err
		}
		outChan <- websockets.GenericMsg{MsgType: websocket.PongMessage, Data: nil}

		return nil
	})

	return conn, err
}

func updateLocation(locationParameters utils.LocationParameters, conn *websocket.Conn, outChan chan websockets.GenericMsg) {
	updateTicker := time.NewTicker(location.UpdateCooldown)

	for {
		select {
		case <-updateTicker.C:
			loc := utils.Location{
				Latitude:  rand.Float64(),
				Longitude: rand.Float64(),
			}

			locationMsg := locationMessages.UpdateLocationMessage{
				Location: loc,
			}
			wsMsg := locationMsg.Serialize()
			genericMsg := websockets.GenericMsg{
				MsgType: websocket.TextMessage,
				Data:    []byte(wsMsg.Serialize()),
			}

			log.Info("updating location")

			outChan <- genericMsg

			err := conn.SetReadDeadline(time.Now().Add(location.Timeout))
			if err != nil {
				log.Error(err)
				return
			}
		}
	}
}

func readMessages(conn *websocket.Conn, inChan chan *websockets.Message, finish chan struct{}) {
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Error(err)
		close(finish)
		return
	} else {
		stringMsg := string(msg)
		msg, err := websockets.ParseMessage(&stringMsg)
		if err != nil {
			log.Error(err)
			return
		}

		inChan <- msg
	}
}
