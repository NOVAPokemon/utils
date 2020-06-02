package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	errors2 "github.com/NOVAPokemon/utils/clients/errors"
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

	Gyms     []utils.GymWithServer
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
	timeoutInDuration  time.Duration
	defaultLocationURL = fmt.Sprintf("%s:%d", utils.Host, utils.LocationPort)
	outChan            chan websockets.GenericMsg
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

		Gyms:                []utils.GymWithServer{},
		HttpClient:          &http.Client{},
		CurrentLocation:     config.Parameters.StartingLocation,
		LocationParameters:  config.Parameters,
		DistanceToStartLat:  0.0,
		DistanceToStartLong: 0.0,
	}
}

func (c *LocationClient) StartLocationUpdates(authToken string, trainersCLient *TrainersClient) error {
	inChan := make(chan *string)
	outChan = make(chan websockets.GenericMsg, bufferSize)
	finish := make(chan struct{})

	conn, err := c.connect(outChan, authToken)
	if err != nil {
		return errors2.WrapStartLocationUpdatesError(err)
	}
	go func() {
		if err := websockets.HandleRecv(conn, inChan, finish); err != nil {
			log.Error(errors2.WrapStartLocationUpdatesError(err))
		}
	}()
	go func() {
		if err := c.updateLocationLoop(conn, outChan); err != nil {
			log.Error(errors2.WrapStartLocationUpdatesError(err))
		}
	}()

	for {
		select {
		case msgString, ok := <-inChan:
			if !ok {
				continue
			}
			// log.Infof("Received message: %s", *msgString)
			msg, err := websockets.ParseMessage(msgString)
			if err != nil {
				return errors2.WrapStartLocationUpdatesError(err)
			}

			switch msg.MsgType {
			case location.Gyms:
				desMsg, err := location.DeserializeLocationMsg(msg)
				if err != nil {
					return errors2.WrapStartLocationUpdatesError(err)
				}
				c.Gyms = desMsg.(*location.GymsMessage).Gyms
			case location.Pokemon:
				desMsg, err := location.DeserializeLocationMsg(msg)
				if err != nil {
					return errors2.WrapStartLocationUpdatesError(err)
				}

				c.Pokemons = desMsg.(*location.PokemonMessage).Pokemon

			case location.CatchPokemonResponse:
				desMsg, err := location.DeserializeLocationMsg(msg)
				if err != nil {
					return errors2.WrapStartLocationUpdatesError(err)
				}

				cwpMsg := desMsg.(*location.CatchWildPokemonMessageResponse)

				if !cwpMsg.Caught {
					log.Info("pokemon got away")
				}

				err = trainersCLient.AppendPokemonTokens(cwpMsg.PokemonTokens)
			case websockets.Error:
				desMsg, err := websockets.DeserializeMsg(msg)
				if err != nil {
					return errors2.WrapStartLocationUpdatesError(err)
				}

				errMsg := desMsg.(*websockets.ErrorMessage)
				log.Error(errMsg.Info)

				if errMsg.Fatal {
					return errors2.WrapStartLocationUpdatesError(errors.New(errMsg.Info))
				}

			default:
				log.Warn("got message type ", msg.MsgType)
			}
		case <-finish:
			log.Warn("Trainer stopped updating location")
			return nil
		case msg := <-outChan:
			err = conn.WriteMessage(msg.MsgType, msg.Data)
			if err != nil {
				return errors2.WrapStartLocationUpdatesError(err)
			}
		}
	}
}

func (c *LocationClient) connect(outChan chan websockets.GenericMsg, authToken string) (*websocket.Conn, error) {

	serverUrl, err := c.GetServerForLocation(c.CurrentLocation)
	if err != nil {
		return nil, errors2.WrapConnectError(err)
	}
	u := url.URL{Scheme: "ws", Host: *serverUrl, Path: api.UserLocationPath}
	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	log.Info("Dialing: ", u.String())
	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		return nil, errors2.WrapConnectError(err)
	}

	err = conn.SetReadDeadline(time.Now().Add(timeoutInDuration))
	conn.SetPingHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(timeoutInDuration)); err != nil {
			return err
		}
		outChan <- websockets.GenericMsg{MsgType: websocket.PongMessage, Data: nil}
		return nil
	})

	return conn, errors2.WrapConnectError(err)
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
		return errors2.WrapUpdateLocation(err)
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

func (c *LocationClient) AddGymLocation(gym utils.GymWithServer) error {

	serverUrl, err := c.GetServerForLocation(gym.Gym.Location)
	if err != nil {
		return errors2.WrapAddGymLocationError(err)
	}

	req, err := BuildRequest("POST", fmt.Sprintf("%s.%s", *serverUrl, c.LocationAddr), api.GymLocationRoute, gym)
	if err != nil {
		return errors2.WrapAddGymLocationError(err)
	}

	_, err = DoRequest(c.HttpClient, req, nil)
	return errors2.WrapAddGymLocationError(err)
}

func (c *LocationClient) GetServerForLocation(loc utils.Location) (*string, error) {
	u := url.URL{Scheme: "http", Host: c.LocationAddr, Path: fmt.Sprintf(api.GetServerForLocationPath)}
	q := u.Query()
	q.Set(api.LatitudeQueryParam, fmt.Sprintf("%f", loc.Latitude))
	q.Set(api.LongitudeQueryParam, fmt.Sprintf("%f", loc.Longitude))
	u.RawQuery = q.Encode()
	log.Info(u.String())
	req, err := http.NewRequest("GET", u.String(), nil)
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, errors2.WrapGetServerForLocation(err)
	}

	var servername string
	err = json.NewDecoder(resp.Body).Decode(&servername)
	if err != nil {
		return nil, errors2.WrapGetServerForLocation(err)
	}
	return &servername, nil
}

func (c *LocationClient) CatchWildPokemon(itemsTokenString string) error {

	itemsToken, err := tokens.ExtractItemsToken(itemsTokenString)
	if err != nil {
		return errors2.WrapCatchWildPokemonError(err)
	}

	pokeball, err := getRandomPokeball(itemsToken.Items)
	if err != nil {
		return errors2.WrapCatchWildPokemonError(err)
	}

	if len(c.Pokemons) <= 0 {
		return errors2.WrapCatchWildPokemonError(errors2.ErrorNoPokemonsVinicity)
	}
	if outChan == nil {
		return errors2.WrapCatchWildPokemonError(errors2.ErrorNotConnected)
	}

	catchingPokemon := c.Pokemons[rand.Intn(len(c.Pokemons))]

	log.Info("will try to catch ", catchingPokemon.Pokemon.Species)

	catchPokemonMsg := location.CatchWildPokemonMessage{
		Pokeball: *pokeball,
		Pokemon:  catchingPokemon.Pokemon.Id.Hex(),
	}
	genericMsg := websockets.GenericMsg{
		MsgType: websocket.TextMessage,
		Data:    []byte(catchPokemonMsg.SerializeToWSMessage().Serialize()),
	}
	outChan <- genericMsg

	return nil
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
		return nil, errors2.ErrorNoPokeballs
	}

	return pokeballs[rand.Intn(len(pokeballs))], nil
}
