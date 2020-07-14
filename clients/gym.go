package clients

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/clients/errors"
	"github.com/NOVAPokemon/utils/comms_manager"
	"github.com/NOVAPokemon/utils/tokens"
	"github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/battles"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type GymClient struct {
	GymAddr      string
	HttpClient   *http.Client
	commsManager comms_manager.CommunicationManager
}

var defaultGymURL = fmt.Sprintf("%s:%d", utils.Host, utils.GymPort)

func NewGymClient(httpClient *http.Client, commsManager comms_manager.CommunicationManager) *GymClient {
	gymURL, exists := os.LookupEnv(utils.GymEnvVar)

	if !exists {
		log.Warn("missing ", utils.GymEnvVar)
		gymURL = defaultGymURL
	}

	return &GymClient{
		GymAddr:      gymURL,
		HttpClient:   httpClient,
		commsManager: commsManager,
	}
}

func (g *GymClient) GetGymInfo(serverHostname string, gymName string) (*utils.Gym, error) {
	req, err := BuildRequest("GET", fmt.Sprintf("%s:%d", serverHostname, utils.GymPort), fmt.Sprintf(api.GetGymInfoPath, gymName), nil)
	if err != nil {
		return nil, errors.WrapGetGymInfoError(err)
	}

	gym := &utils.Gym{}
	_, err = DoRequest(g.HttpClient, req, gym, g.commsManager)
	return gym, errors.WrapGetGymInfoError(err)
}

func (g *GymClient) CreateGym(toCreate utils.Gym) (*utils.Gym, error) {
	req, err := BuildRequest("POST", g.GymAddr, api.CreateGymPath, toCreate)
	if err != nil {
		return nil, errors.WrapCreateGymError(err)
	}

	createdGym := &utils.Gym{}
	_, err = DoRequest(g.HttpClient, req, createdGym, g.commsManager)
	return createdGym, errors.WrapCreateGymError(err)
}

func (g *GymClient) CreateRaid(serverHostname string, gymName string) error {
	req, err := BuildRequest("POST", fmt.Sprintf("%s:%d", serverHostname, utils.GymPort), fmt.Sprintf(api.CreateRaidPath, gymName), nil)
	if err != nil {
		return errors.WrapCreateRaidError(err)
	}

	_, err = DoRequest(g.HttpClient, req, nil, g.commsManager)
	log.Info("Finish createRaid")

	return errors.WrapCreateRaidError(err)
}

func (g *GymClient) EnterRaid(authToken string, pokemonsTokens []string, statsToken string, itemsToken string,
	gymId string, serverHostname string) (*websocket.Conn, *battles.BattleChannels, error) {
	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%d", serverHostname, utils.GymPort), Path: fmt.Sprintf(api.JoinRaidPath, gymId)}
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
		err = errors.WrapEnterRaidError(websockets.WrapDialingError(err, u.String()))
		return nil, nil, err
	}

	outChannel := make(chan websockets.GenericMsg)
	inChannel := make(chan string)
	finished := make(chan struct{})

	SetDefaultPingHandler(c, outChannel)

	go ReadMessagesFromConnToChan(c, inChannel, finished)
	go WriteMessagesFromChanToConn(c, g.commsManager, outChannel, finished)

	return c, &battles.BattleChannels{OutChannel: outChannel, InChannel: inChannel, FinishChannel: finished}, nil
}
