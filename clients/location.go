package clients

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/gps"
	"github.com/NOVAPokemon/utils/tokens"
	"github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/location"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

const (
	bufferSize = 10
)

var (
	timeoutInDuration time.Duration
)

type LocationClient struct {
	LocationAddr string
	config       utils.LocationClientConfig

	Gyms                []utils.Gym
	HttpClient          *http.Client
	CurrentLocation     utils.Location
	LocationParameters  utils.LocationParameters
	DistanceToStartLat  float64
	DistanceToStartLong float64
}

func NewLocationClient(addr string, config utils.LocationClientConfig) *LocationClient {
	timeoutInDuration = time.Duration(config.Timeout) * time.Second
	return &LocationClient{
		LocationAddr: addr,
		config:       config,

		Gyms:                []utils.Gym{},
		HttpClient:          &http.Client{},
		CurrentLocation:     config.Parameters.StartingLocation,
		LocationParameters:  config.Parameters,
		DistanceToStartLat:  0.0,
		DistanceToStartLong: 0.0,
	}
}

func (c *LocationClient) StartLocationUpdates(authToken string) {
	inChan := make(chan *websockets.Message)
	outChan := make(chan websockets.GenericMsg, bufferSize)
	finish := make(chan struct{})

	conn, err := c.connect(outChan, authToken)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("timeout: ", c.config.Timeout)

	go readMessages(conn, inChan, finish)
	go c.updateLocation(conn, outChan)

	for {
		select {
		case msg := <-inChan:
			switch msg.MsgType {
			case location.Gyms:
				c.Gyms = location.Deserialize(msg).(*location.GymsMessage).Gyms
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

	_ = conn.SetReadDeadline(time.Now().Add(timeoutInDuration))
	conn.SetPingHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(timeoutInDuration)); err != nil {
			return err
		}
		outChan <- websockets.GenericMsg{MsgType: websocket.PongMessage, Data: nil}

		return nil
	})

	return conn, err
}

func (c *LocationClient) updateLocation(conn *websocket.Conn, outChan chan websockets.GenericMsg) {
	updateTicker := time.NewTicker(time.Duration(c.config.UpdateInterval) * time.Second)

	for {
		select {
		case <-updateTicker.C:
			locationMsg := location.UpdateLocationMessage{
				Location: c.CurrentLocation,
			}
			wsMsg := locationMsg.SerializeToWSMessage()
			genericMsg := websockets.GenericMsg{
				MsgType: websocket.TextMessage,
				Data:    []byte(wsMsg.Serialize()),
			}

			//log.Info("updating location: ", c.CurrentLocation)

			outChan <- genericMsg

			err := conn.SetReadDeadline(time.Now().Add(timeoutInDuration))
			if err != nil {
				log.Error(err)
				return
			}

			if rand.Float64() <= c.LocationParameters.MovingProbability {
				c.CurrentLocation = c.move(c.config.UpdateInterval)
			}

			// log.Info(c.DistanceToStartLat, c.DistanceToStartLong)
		}
	}
}

func (c *LocationClient) move(timePassed int) utils.Location {
	randAngle := rand.Float64() * 2 * math.Pi
	distanceTraveled := rand.Float64() * c.LocationParameters.MaxMovingSpeed * float64(timePassed)

	dLat := randAngle * math.Sin(distanceTraveled)
	if math.Abs(c.DistanceToStartLat+dLat) > float64(c.LocationParameters.MaxDistanceFromStart) {
		dLat *= -1
	}

	dLong := randAngle * math.Cos(distanceTraveled)
	if math.Abs(c.DistanceToStartLong+dLong) > float64(c.LocationParameters.MaxDistanceFromStart) {
		dLong *= -1
	}

	c.DistanceToStartLat += dLat
	c.DistanceToStartLong += dLong

	return gps.CalcLocationPlusDistanceTraveled(c.CurrentLocation, dLat, dLong)
}

func (c *LocationClient) AddGymLocation(gym utils.Gym) error {
	req, err := BuildRequest("POST", c.LocationAddr, api.GymLocationRoute, gym)
	if err != nil {
		return err
	}

	_, err = DoRequest(c.HttpClient, req, nil)
	return err
}

func readMessages(conn *websocket.Conn, inChan chan *websockets.Message, finish chan struct{}) {
	for {
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
}
