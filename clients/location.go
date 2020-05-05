package clients

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/gps"
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/tokens"
	"github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/location"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
)

type CaughtPokemonMessage struct {
	Caught bool
}

type LocationClient struct {
	LocationAddr string
	config       utils.LocationClientConfig
	HttpClient   *http.Client

	Gyms     []utils.Gym
	Pokemons []utils.WildPokemon

	CurrentLocation     utils.Location
	LocationParameters  utils.LocationParameters
	DistanceToStartLat  float64
	DistanceToStartLong float64
}

const (
	bufferSize = 10
)

var (
	timeoutInDuration time.Duration

	defaultLocationURL = fmt.Sprintf("%s:%d", utils.Host, utils.LocationPort)
)

func NewLocationClient(config utils.LocationClientConfig) *LocationClient {
	locationURL, exists := os.LookupEnv(utils.LocationEnvVar)

	if !exists {
		log.Warn("missing ", utils.LocationEnvVar)
		locationURL = defaultLocationURL
	}

	timeoutInDuration = time.Duration(config.Timeout) * time.Second
	return &LocationClient{
		LocationAddr: locationURL,
		config:       config,

		Gyms:                []utils.Gym{},
		HttpClient:          &http.Client{},
		CurrentLocation:     config.Parameters.StartingLocation,
		LocationParameters:  config.Parameters,
		DistanceToStartLat:  0.0,
		DistanceToStartLong: 0.0,
	}
}

func (c *LocationClient) StartLocationUpdates(authToken string) error {
	inChan := make(chan *string)
	outChan := make(chan websockets.GenericMsg, bufferSize)
	finish := make(chan struct{})

	conn, err := c.connect(outChan, authToken)
	if err != nil {
		return wrapStartLocationUpdatesError(err)
	}

	go func() {
		if err := websockets.HandleRecv(conn, inChan, finish); err != nil {
			log.Error(wrapStartLocationUpdatesError(err))
		}
	}()
	go func() {
		if err := c.updateLocationLoop(conn, outChan); err!= nil {
			log.Error(wrapStartLocationUpdatesError(err))
		}
	}()

	for {
		select {
		case msgString, ok := <-inChan:
			if !ok {
				continue
			}

			msg, err := websockets.ParseMessage(msgString)
			if err != nil {
				return wrapStartLocationUpdatesError(err)
			}

			switch msg.MsgType {
			case location.Gyms:
				c.Gyms = location.Deserialize(msg).(*location.GymsMessage).Gyms
			case location.Pokemon:
				c.Pokemons = location.Deserialize(msg).(*location.PokemonMessage).Pokemon
			default:
				log.Warn("got message type ", msg.MsgType)
			}
		case <-finish:
			log.Warn("Trainer stopped updating location")
			return nil
		case msg := <-outChan:
			err = conn.WriteMessage(msg.MsgType, msg.Data)
			if err != nil {
				return wrapStartLocationUpdatesError(err)
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
		return nil, wrapConnectError(err)
	}

	err = conn.SetReadDeadline(time.Now().Add(timeoutInDuration))
	conn.SetPingHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(timeoutInDuration)); err != nil {
			return err
		}
		outChan <- websockets.GenericMsg{MsgType: websocket.PongMessage, Data: nil}

		return nil
	})

	return conn, wrapConnectError(err)
}

func (c *LocationClient) updateLocationLoop(conn *websocket.Conn, outChan chan websockets.GenericMsg) error {
	err := c.updateLocation(conn, outChan)
	if err != nil {
		return err
	}

	updateTicker := time.NewTicker(time.Duration(c.config.UpdateInterval) * time.Second)

	for {
		select {
		case <-updateTicker.C:
			err = c.updateLocation(conn, outChan)
			if err != nil {
				return err
			}
		}
	}
}

func (c *LocationClient) updateLocation(conn *websocket.Conn, outChan chan websockets.GenericMsg) error {
	locationMsg := location.UpdateLocationMessage{
		Location: c.CurrentLocation,
	}
	wsMsg := locationMsg.SerializeToWSMessage()
	genericMsg := websockets.GenericMsg{
		MsgType: websocket.TextMessage,
		Data:    []byte(wsMsg.Serialize()),
	}

	log.Info("updating location: ", c.CurrentLocation)

	outChan <- genericMsg

	err := conn.SetReadDeadline(time.Now().Add(timeoutInDuration))
	if err != nil {
		return wrapUpdateLocation(err)
	}

	if rand.Float64() <= c.LocationParameters.MovingProbability {
		c.CurrentLocation = c.move(c.config.UpdateInterval)
	}

	return nil
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
		return wrapAddGymLocationError(err)
	}

	_, err = DoRequest(c.HttpClient, req, nil)
	return wrapAddGymLocationError(err)
}

func (c *LocationClient) CatchWildPokemon(authToken, itemsTokenString string) (caught bool,
	header http.Header, err error) {
	itemsToken, err := tokens.ExtractItemsToken(itemsTokenString)
	if err != nil {
		return false, nil, wrapCatchWildPokemonError(err)
	}

	pokeball, err := getRandomPokeball(itemsToken.Items)
	if err != nil {
		return false, nil, wrapCatchWildPokemonError(err)
	}

	if len(c.Pokemons) <= 0 {
		err = wrapCatchWildPokemonError(errorNoPokemonsVinicity)
		return false, nil, err
	}

	catchingPokemon := c.Pokemons[rand.Intn(len(c.Pokemons))]

	log.Info("will try to catch ", catchingPokemon.Pokemon.Species)

	requestBody := api.CatchWildPokemonRequest{
		Pokeball: *pokeball,
		Pokemon:  catchingPokemon.Pokemon,
	}

	req, err := BuildRequest("GET", c.LocationAddr, api.CatchWildPokemonPath, requestBody)
	if err != nil {
		return false, nil, wrapCatchWildPokemonError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var msg CaughtPokemonMessage
	resp, err := DoRequest(c.HttpClient, req, &msg)
	if err != nil {
		return false, nil, wrapCatchWildPokemonError(err)
	}

	if !msg.Caught {
		return false, nil, nil
	}

	for i, pokemon := range c.Pokemons {
		if pokemon.Pokemon.Id.Hex() == catchingPokemon.Pokemon.Id.Hex() {
			log.Info("Removing caught pokemon...")
			c.Pokemons = append(c.Pokemons[:i], c.Pokemons[i+1:]...)
			log.Info(c.Pokemons)
			break
		}
	}

	return true, resp.Header, nil
}

func getRandomPokeball(itemsFromToken map[string]items.Item) (*items.Item, error) {
	var pokeballs []*items.Item
	for _, item := range itemsFromToken {
		if item.IsPokeBall() {
			toAdd := item
			pokeballs = append(pokeballs, &toAdd)
		}
	}

	if pokeballs == nil {
		return nil, errorNoPokeballs
	}

	return pokeballs[rand.Intn(len(pokeballs))], nil
}
