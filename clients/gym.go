package clients

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/tokens"
	"github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/battles"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"time"
)

type GymClient struct {
	GymAddr    string
	HttpClient *http.Client
}

var defaultGymURL = fmt.Sprintf("%s:%d", utils.Host, utils.GymPort)

func NewGymClient(httpClient *http.Client) *GymClient {
	gymURL, exists := os.LookupEnv(utils.GymEnvVar)

	if !exists {
		log.Warn("missing ", utils.GymEnvVar)
		gymURL = defaultGymURL
	}

	return &GymClient{
		GymAddr:    gymURL,
		HttpClient: httpClient,
	}
}

func (g *GymClient) GetGymInfo(gymName string) (*utils.Gym, error) {
	req, err := BuildRequest("GET", g.GymAddr, fmt.Sprintf(api.GetGymInfoPath, gymName), nil)
	if err != nil {
		return nil, wrapGetGymInfoError(err)
	}

	gym := &utils.Gym{}
	_, err = DoRequest(g.HttpClient, req, gym)
	return gym, wrapGetGymInfoError(err)
}

func (g *GymClient) CreateGym(toCreate utils.Gym) (*utils.Gym, error) {
	req, err := BuildRequest("POST", g.GymAddr, api.CreateGymPath, toCreate)
	if err != nil {
		return nil, wrapCreateGymError(err)
	}

	createdGym := &utils.Gym{}
	_, err = DoRequest(g.HttpClient, req, createdGym)
	return createdGym, wrapCreateGymError(err)
}

func (g *GymClient) CreateRaid(gymName string) error {
	req, err := BuildRequest("POST", g.GymAddr, fmt.Sprintf(api.CreateRaidPath, gymName), nil)
	if err != nil {
		return wrapCreateRaidError(err)
	}

	_, err = DoRequest(g.HttpClient, req, nil)
	log.Info("Finish createRaid")

	return wrapCreateRaidError(err)
}

func (g *GymClient) EnterRaid(authToken string, pokemonsTokens []string, statsToken string, itemsToken string,
	gymId string) (*websocket.Conn, *battles.BattleChannels, error) {

	log.Infof("Dialing: %s %s", g.GymAddr, fmt.Sprintf(api.GetGymInfoPath, gymId))
	u := url.URL{Scheme: "ws", Host: g.GymAddr, Path: fmt.Sprintf(api.JoinRaidPath, gymId)}
	log.Infof("Connecting to: %s", u.String())

	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)
	header.Set(tokens.StatsTokenHeaderName, statsToken)
	header[tokens.PokemonsTokenHeaderName] = pokemonsTokens
	header.Set(tokens.ItemsTokenHeaderName, itemsToken)

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	c, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		err = wrapEnterRaidError(websockets.WrapDialingError(err, u.String()))
		return nil, nil, err
	}

	outChannel := make(chan *string)
	inChannel := make(chan *string)
	finished := make(chan struct{})

	go ReadMessagesToChan(c, inChannel, finished)
	go MainLoop(c, outChannel, finished)

	return c, &battles.BattleChannels{OutChannel: outChannel, InChannel: inChannel, FinishChannel: finished}, nil
}
