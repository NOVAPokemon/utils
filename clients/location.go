package clients

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/tokens"
	"github.com/NOVAPokemon/utils/websockets/location"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type LocationClient struct {
	LocationAddr string
	HttpClient   *http.Client
}

func NewLocationClient(addr string) *LocationClient {
	return &LocationClient{
		LocationAddr: addr,
		HttpClient:   &http.Client{},
	}
}

func (c *LocationClient) StartLocationUpdates(authToken string) {
	u := url.URL{Scheme: "ws", Host: c.LocationAddr, Path: fmt.Sprintf(api.UserLocationPath)}
	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)
	writeMut := sync.Mutex{}

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	log.Info("Dialing: ", u.String())
	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(location.Timeout))
	conn.SetPingHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(location.Timeout))
		writeMut.Lock()
		_ = conn.WriteMessage(websocket.PongMessage, nil)
		writeMut.Unlock()
		return nil
	})

	var updateTicker = time.NewTicker(location.UpdateCooldown)
	var inChan = make(chan string)
	finish := make(chan *struct{})
	go func() {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Error(err)
			finish <- nil
			return
		} else {
			inChan <- string(msg)
		}
	}()

	for {
		select {
		case <-updateTicker.C:
			loc, err := c.getLocation()
			if err != nil {
				log.Error(err)
				return
			}

			writeMut.Lock()
			err = conn.WriteJSON(*loc)
			writeMut.Unlock()
			if err != nil {
				log.Error(err)
				return
			}

			err = conn.SetReadDeadline(time.Now().Add(location.Timeout))
			if err != nil {
				log.Error(err)
				return
			}
		case msg := <-inChan:
			log.Info(msg)
		case <-finish:
			log.Warn("Trainer stopped updating location")
		}
	}
}

func (c *LocationClient) getLocation() (*utils.Location, error) {
	return &utils.Location{
		Latitude: rand.Float64(), Longitude: rand.Float64(),
	}, nil // TODO
}
