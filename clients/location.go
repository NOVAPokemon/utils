package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	errors2 "github.com/NOVAPokemon/utils/clients/errors"
	"github.com/NOVAPokemon/utils/gps"
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/tokens"
	ws "github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/location"
	"github.com/golang/geo/s2"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type (
	CaughtPokemonMessage struct {
		Caught bool
	}

	toConnChansValueType     = chan *ws.WebsocketMsg
	finishConnChansValueType = chan struct{}
	connectionsValueType     = *websocket.Conn
	gymsValueType            = []utils.GymWithServer
)

type LocationClient struct {
	LocationAddr string
	config       utils.LocationClientConfig
	HttpClient   *http.Client

	gyms sync.Map

	pokemonsLock sync.Mutex
	pokemons     []utils.WildPokemonWithServer

	CurrentLocation     s2.LatLng
	LocationParameters  utils.LocationParameters
	DistanceToStartLat  float64
	DistanceToStartLong float64

	serversConnected []string
	fromConnChan     chan *ws.WebsocketMsg
	toConnsChans     sync.Map
	finishConnChans  sync.Map
	connections      sync.Map

	commsManager ws.CommunicationManager

	updateConnectionsLock sync.Mutex
}

const (
	bufferSize                   = 10
	defaultRegionsToAreaFilename = "regions_to_area.json"
)

var (
	timeoutInDuration     time.Duration
	defaultLocationURL    = fmt.Sprintf("%s:%d", utils.Host, utils.LocationPort)
	catchPokemonResponses chan *location.CatchWildPokemonMessageResponse
)

func NewLocationClient(config utils.LocationClientConfig, region string,
	manager ws.CommunicationManager, httpClient *http.Client) *LocationClient {
	locationURL, exists := os.LookupEnv(utils.LocationEnvVar)

	if !exists {
		log.Warn("missing ", utils.LocationEnvVar)
		locationURL = defaultLocationURL
	}

	timeoutInDuration = time.Duration(config.Timeout) * time.Second

	var startingLocation s2.LatLng
	if config.Parameters.StartingLocation {
		log.Info("starting location enabled")
		startingLocation = s2.LatLngFromDegrees(config.Parameters.StartingLocationLat,
			config.Parameters.StartingLocationLon)
	} else if region != "" {
		log.Info("starting location disabled")
		startingLocation = getRandomLatLng(region)
	}

	return &LocationClient{
		LocationAddr:        locationURL,
		config:              config,
		pokemonsLock:        sync.Mutex{},
		gyms:                sync.Map{},
		HttpClient:          httpClient,
		CurrentLocation:     startingLocation,
		LocationParameters:  config.Parameters,
		DistanceToStartLat:  0.0,
		DistanceToStartLong: 0.0,
		serversConnected:    []string{},
		fromConnChan:        make(chan *ws.WebsocketMsg, bufferSize),
		finishConnChans:     sync.Map{},
		connections:         sync.Map{},
		toConnsChans:        sync.Map{},
		commsManager:        manager,
	}
}

func (c *LocationClient) StartLocationUpdates(authToken string) error {
	catchPokemonResponses = make(chan *location.CatchWildPokemonMessageResponse)

	serverUrl, err := c.GetServerForLocation(c.CurrentLocation)
	if err != nil {
		return errors2.WrapStartLocationUpdatesError(err)
	}

	err = c.handleLocationConnection(*serverUrl, authToken)
	if err != nil {
		return errors2.WrapStartLocationUpdatesError(err)
	}

	go c.updateLocationLoop()

	for {
		msgString, ok := <-c.fromConnChan
		if !ok {
			break
		}
		if err = c.handleLocationMsg(msgString, authToken); err != nil {
			return errors2.WrapStartLocationUpdatesError(err)
		}
	}

	return errors.New("stopped updating location")
}

func (c *LocationClient) handleLocationConnection(serverUrl, authToken string) error {
	outChan := make(toConnChansValueType, bufferSize)

	conn, err := c.connect(serverUrl, outChan, authToken)
	if err != nil {
		return errors2.WrapStartLocationUpdatesError(err)
	}

	finishChan := make(chan struct{})

	c.toConnsChans.Store(serverUrl, outChan)
	c.finishConnChans.Store(serverUrl, finishChan)
	c.connections.Store(serverUrl, conn)
	c.serversConnected = append(c.serversConnected, serverUrl)

	SetDefaultPingHandler(conn, outChan)

	go ReadMessagesFromConnToChanWithoutClosing(conn, c.fromConnChan, finishChan, c.commsManager)
	go WriteTextMessagesFromChanToConn(conn, c.commsManager, outChan, finishChan)

	log.Info("handling connection to ", serverUrl)

	return nil
}

func (c *LocationClient) handleLocationMsg(wsMsg *ws.WebsocketMsg, authToken string) error {
	// log.Infof("Received message: %s", *msgString)
	msgData := wsMsg.Content.Data

	switch wsMsg.Content.AppMsgType {
	case location.Gyms:
		gymsMsg := &location.GymsMessage{}
		if err := mapstructure.Decode(msgData, gymsMsg); err != nil {
			panic(err)
		}
		gyms := gymsMsg.Gyms
		if len(gyms) > 0 {
			server := gyms[0].ServerName
			c.SetGyms(server, gyms)
		}
	case location.Pokemon:
		pokemonMsg := &location.PokemonMessage{}
		if err := mapstructure.Decode(msgData, pokemonMsg); err != nil {
			panic(err)
		}
		c.SetPokemons(pokemonMsg.Pokemon)
	case location.CatchPokemonResponse:

		cwpMsg := &location.CatchWildPokemonMessageResponse{}
		if err := mapstructure.Decode(msgData, cwpMsg); err != nil {
			panic(err)
		}

		catchPokemonResponses <- cwpMsg
	case location.ServersResponse:

		serversMsg := &location.ServersMessage{}
		if err := mapstructure.Decode(msgData, serversMsg); err != nil {
			panic(err)
		}

		log.Info("received servers ", serversMsg.Servers)
		err := c.updateConnections(serversMsg.Servers, authToken)
		if err != nil {
			return errors2.WrapHandleLocationMsgError(err)
		}
	case location.CellsResponse:
		cellsMsg := &location.CellsPerServerMessage{}
		if err := mapstructure.Decode(msgData, cellsMsg); err != nil {
			panic(err)
		}

		c.updateLocationWithCells(cellsMsg.CellsPerServer, cellsMsg.OriginServer)
	case ws.Error:
		errMsg := &ws.ErrorMessage{}
		if err := mapstructure.Decode(msgData, errMsg); err != nil {
			panic(err)
		}
		log.Error(errMsg.Info)

		if errMsg.Fatal {
			return errors2.WrapHandleLocationMsgError(errors.New(errMsg.Info))
		}
	default:
		log.Warn("got message type ", wsMsg.Content.AppMsgType)
	}

	return nil
}

func (c *LocationClient) updateConnections(servers []string, authToken string) error {
	c.updateConnectionsLock.Lock()
	defer c.updateConnectionsLock.Unlock()

	var (
		isNewServer bool
	)

	var toRemove []string
	var newServers []string

	// Add new connections
	for i := range servers {
		isNewServer = true

		for j := range c.serversConnected {
			if servers[i] == c.serversConnected[j] {
				isNewServer = false
			}
		}

		if isNewServer {
			err := c.handleLocationConnection(servers[i], authToken)
			if err != nil {
				return errors2.WrapUpdateConnectionsError(err)
			}
			newServers = append(newServers, c.serversConnected[i])
		}
	}

	// Remove unused connections
	remove := true
	for i := range c.serversConnected {
		remove = true
		for j := range servers {
			if c.serversConnected[i] == servers[j] {
				remove = false
				break
			}
		}

		if remove {
			finishChanValue, ok := c.finishConnChans.Load(c.serversConnected[i])
			if !ok {
				return errors.New("tried to finish location connection without a finish chan")
			}
			log.Info("finishing connection to ", c.serversConnected[i])
			close(finishChanValue.(finishConnChansValueType))
			c.toConnsChans.Delete(c.serversConnected[i])
			toRemove = append(toRemove, c.serversConnected[i])
		}
	}

	var add bool
	for i := range c.serversConnected {
		add = true
		for j := range toRemove {
			if c.serversConnected[i] == toRemove[j] {
				add = false
			}
		}

		if add {
			newServers = append(newServers, c.serversConnected[i])
		}
	}

	c.serversConnected = newServers

	return nil
}

func (c *LocationClient) connect(serverUrl string, outChan chan *ws.WebsocketMsg,
	authToken string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%d", serverUrl, utils.LocationPort), Path: api.UserLocationPath}
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
		if err = conn.SetReadDeadline(time.Now().Add(timeoutInDuration)); err != nil {
			return err
		}
		outChan <- ws.NewControlMsg(websocket.PongMessage)
		return nil
	})

	return conn, errors2.WrapConnectError(err)
}

func (c *LocationClient) updateLocationLoop() {
	c.updateLocation()
	updateTicker := time.NewTicker(time.Duration(c.config.UpdateInterval) * time.Second)

	for {
		select {
		case <-updateTicker.C:
			c.updateLocation()

			if rand.Float64() <= c.LocationParameters.MovingProbability {
				c.CurrentLocation = c.move(c.config.UpdateInterval)
			}
		}
	}
}

func (c *LocationClient) updateLocation() {
	locationMsg := location.UpdateLocationMessage{
		Location: c.CurrentLocation,
	}

	c.HttpClient.SetLocation(c.CurrentLocation)
	log.Info("updating location: ", c.CurrentLocation)

	// Only runs once
	c.toConnsChans.Range(func(serverUrl, toConnChanValue interface{}) bool {
		log.Info("updating location to ", serverUrl)

		toConnChan := toConnChanValue.(toConnChansValueType)
		log.Infof("sending location msg %v", locationMsg)
		toConnChan <- locationMsg.ConvertToWSMessage()

		connValue, ok := c.connections.Load(serverUrl)
		if !ok {
			panic("tried to write to a connection that is not in the map")
		}

		conn := connValue.(connectionsValueType)
		err := conn.SetReadDeadline(time.Now().Add(timeoutInDuration))
		if err != nil {
			panic("error setting deadline")
		}

		return false
	})
}

func (c *LocationClient) updateLocationWithCells(tilesPerServer map[string]s2.CellUnion, excludeServer string) {
	locationMsg := location.UpdateLocationWithTilesMessage{
		CellsPerServer: tilesPerServer,
	}

	log.Infof("updating location with tiles %v", tilesPerServer)

	c.toConnsChans.Range(func(serverUrl, toConnChanValue interface{}) bool {
		if serverUrl == excludeServer {
			return true
		}

		toConnChan := toConnChanValue.(toConnChansValueType)
		toConnChan <- locationMsg.ConvertToWSMessage()

		connValue, ok := c.connections.Load(serverUrl)
		if !ok {
			panic("tried to write to a connection that is not in the map")
		}

		log.Info("sending precomputed tiles to ", serverUrl)

		conn := connValue.(connectionsValueType)
		err := conn.SetReadDeadline(time.Now().Add(timeoutInDuration))
		if err != nil {
			panic("error setting deadline")
		}

		return true
	})
}

func (c *LocationClient) move(timePassed int) s2.LatLng {
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
	req, err := BuildRequest("POST", c.LocationAddr, api.GymLocationRoute, gym)
	if err != nil {
		return errors2.WrapAddGymLocationError(err)
	}
	_, err = DoRequest(c.HttpClient, req, nil, c.commsManager)
	return errors2.WrapAddGymLocationError(err)
}

func (c *LocationClient) GetServerForLocation(loc s2.LatLng) (*string, error) {
	u := url.URL{Scheme: "http", Host: c.LocationAddr, Path: fmt.Sprintf(api.GetServerForLocationPath)}
	q := u.Query()
	q.Set(api.LatitudeQueryParam, fmt.Sprintf("%f", loc.Lat.Degrees()))
	q.Set(api.LongitudeQueryParam, fmt.Sprintf("%f", loc.Lng.Degrees()))
	u.RawQuery = q.Encode()
	log.Info(u.String())
	req, err := http.NewRequest("GET", u.String(), nil)
	var servername string
	_, err = DoRequest(http.DefaultClient, req, &servername, c.commsManager)
	if err != nil {
		return nil, errors2.WrapGetServerForLocationError(err)
	}
	return &servername, nil
}

func (c *LocationClient) CatchWildPokemon(trainersClient *TrainersClient) error {
	itemsTokenString := trainersClient.ItemsToken
	itemsToken, err := tokens.ExtractItemsToken(itemsTokenString)
	if err != nil {
		return errors2.WrapCatchWildPokemonError(err)
	}

	pokeball, err := getRandomPokeball(itemsToken.Items)
	if err != nil {
		return errors2.WrapCatchWildPokemonError(err)
	}

	c.pokemonsLock.Lock()
	pokemonsLen := len(c.pokemons)

	if pokemonsLen <= 0 {
		c.pokemonsLock.Unlock()
		log.Warn(errors2.WrapCatchWildPokemonError(errors2.ErrorNoPokemonsVinicity))
		return nil
	}

	toCatch := c.pokemons[rand.Intn(pokemonsLen)]
	c.pokemonsLock.Unlock()

	log.Info("will try to catch ", toCatch.Pokemon.Species)

	catchPokemonMsg := location.CatchWildPokemonMessage{
		Pokeball:    *pokeball,
		WildPokemon: toCatch,
	}

	toConnChanValue, ok := c.toConnsChans.Load(toCatch.Server)
	if !ok {
		return errors2.WrapCatchWildPokemonError(
			errors.New(fmt.Sprintf("no connection for pokemon %s", toCatch.Server)))
	}

	toConnChan := toConnChanValue.(toConnChansValueType)
	toConnChan <- catchPokemonMsg.ConvertToWSMessage()
	catchResponse := <-catchPokemonResponses

	if catchResponse.Error != "" {
		log.Error(catchResponse.Error)
		return nil
	}

	if !catchResponse.Caught {
		log.Info("pokemon got away")
		return nil
	}

	return trainersClient.AppendPokemonTokens(catchResponse.PokemonTokens)
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

func getRandomLatLng(region string) s2.LatLng {
	regionsToArea := loadRegionsToArea(defaultRegionsToAreaFilename)
	areas := regionsToArea.Regions[region]
	randArea := areas[rand.Intn(len(areas))]

	deltaLat := math.Abs(randArea.TopLeft.Lat - randArea.BotRight.Lat)
	deltaLng := math.Abs(randArea.TopLeft.Lng - randArea.BotRight.Lng)
	randomLat := randArea.TopLeft.Lat - rand.Float64()*(deltaLat)
	randomLng := randArea.BotRight.Lng - rand.Float64()*(deltaLng)

	randomLatLng := s2.LatLngFromDegrees(randomLat, randomLng)

	return randomLatLng
}

func loadRegionsToArea(regionsFilename string) *utils.RegionsToAreas {
	fileData, err := ioutil.ReadFile(regionsFilename)
	if err != nil {
		log.Error("error loading regions filename")
		panic(err)
	}

	var regionsToArea utils.RegionsToAreas
	err = json.Unmarshal(fileData, &regionsToArea)
	if err != nil {
		panic(err)
		return nil
	}

	return &regionsToArea
}

func (c *LocationClient) GetWildPokemons() []utils.WildPokemonWithServer {
	c.pokemonsLock.Lock()
	pokemonsCopy := c.pokemons
	defer c.pokemonsLock.Unlock()
	return pokemonsCopy
}

func (c *LocationClient) SetPokemons(newPokemons []utils.WildPokemonWithServer) {
	c.pokemonsLock.Lock()
	c.pokemons = newPokemons
	c.pokemonsLock.Unlock()
}

func (c *LocationClient) GetGyms() []utils.GymWithServer {
	var allGyms []utils.GymWithServer
	c.gyms.Range(func(_, gymsValue interface{}) bool {
		gyms := gymsValue.(gymsValueType)
		allGyms = append(allGyms, gyms...)
		return true
	})

	return allGyms
}

func (c *LocationClient) SetGyms(server string, newGyms []utils.GymWithServer) {
	c.gyms.Store(server, newGyms)
}
