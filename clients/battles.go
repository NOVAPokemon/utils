package clients

import (
	"encoding/json"
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/clients/errors"
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

type BattleLobbyClient struct {
	BattlesAddr string
	httpClient  http.Client
}

var defaultBattleURL = fmt.Sprintf("%s:%d", utils.Host, utils.BattlesPort)

func NewBattlesClient() *BattleLobbyClient {
	battlesURL, exists := os.LookupEnv(utils.BattlesEnvVar)

	if !exists {
		log.Warn("missing ", utils.BattlesEnvVar)
		battlesURL = defaultBattleURL
	}

	return &BattleLobbyClient{
		BattlesAddr: battlesURL,
		httpClient:  http.Client{},
	}
}

func (client *BattleLobbyClient) GetAvailableLobbies() ([]utils.Lobby, error) {
	u := url.URL{Scheme: "http", Host: client.BattlesAddr, Path: "/battles"}

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, errors.WrapGetBattleLobbiesError(err)
	}

	var availableBattles []utils.Lobby
	err = json.NewDecoder(resp.Body).Decode(&availableBattles)
	if err != nil {
		return nil, errors.WrapGetBattleLobbiesError(err)
	}

	return availableBattles, nil
}

func (client *BattleLobbyClient) QueueForBattle(authToken string, pokemonsTokens []string, statsToken string,
	itemsToken string) (*websocket.Conn, *battles.BattleChannels, error) {
	u := url.URL{Scheme: "ws", Host: client.BattlesAddr, Path: api.QueueForBattlePath}
	log.Infof("Queuing for battle: %s", u.String())

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)
	header.Set(tokens.StatsTokenHeaderName, statsToken)
	header.Set(tokens.ItemsTokenHeaderName, itemsToken)
	header[tokens.PokemonsTokenHeaderName] = pokemonsTokens

	c, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		err = errors.WrapQueueForBattleError(websockets.WrapDialingError(err, u.String()))
		return nil, nil, err
	}

	outChannel := make(chan websockets.GenericMsg)
	inChannel := make(chan *string)
	rejectedChannel := make(chan struct{})
	finished := make(chan struct{})

	go ReadMessagesToChan(c, inChannel, finished)
	go MainLoop(c, outChannel, finished)

	battleChannels := battles.BattleChannels{
		OutChannel:      outChannel,
		InChannel:       inChannel,
		RejectedChannel: rejectedChannel,
		FinishChannel:   finished,
	}

	return c, &battleChannels, nil
}

func (client *BattleLobbyClient) ChallengePlayerToBattle(authToken string, pokemonsTokens []string, statsToken string,
	itemsToken string, targetPlayer string) (*websocket.Conn, *battles.BattleChannels, error) {
	u := url.URL{Scheme: "ws", Host: client.BattlesAddr, Path: fmt.Sprintf(api.ChallengeToBattlePath, targetPlayer)}
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
		err = errors.WrapChallengeForBattleError(websockets.WrapDialingError(err, u.String()))
		return nil, nil, err
	}

	inChannel := make(chan *string)
	outChannel := make(chan websockets.GenericMsg)
	rejectedChannel := make(chan struct{})
	finished := make(chan struct{})

	go ReadMessagesToChan(c, inChannel, finished)
	go MainLoop(c, outChannel, finished)

	battleChannels := battles.BattleChannels{
		OutChannel:      outChannel,
		InChannel:       inChannel,
		RejectedChannel: rejectedChannel,
		FinishChannel:   finished,
	}

	return c, &battleChannels, nil
}

func (client *BattleLobbyClient) AcceptChallenge(authToken string, pokemonsTokens []string, statsToken string,
	itemsToken string, battleId string, serverHostname string) (*websocket.Conn, *battles.BattleChannels, error) {

	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%d", serverHostname, utils.BattlesPort), Path: fmt.Sprintf(api.AcceptChallengePath, battleId)}
	log.Infof("Accepting challenge: %s", u.String())
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
		err = errors.WrapAcceptBattleChallengeError(websockets.WrapDialingError(err, u.String()))
		return nil, nil, err
	}

	outChannel := make(chan websockets.GenericMsg)
	inChannel := make(chan *string)
	rejectedChannel := make(chan struct{})
	finished := make(chan struct{})

	go ReadMessagesToChan(c, inChannel, finished)
	go MainLoop(c, outChannel, finished)

	battleChannels := battles.BattleChannels{
		OutChannel: outChannel,
		InChannel: inChannel,
		RejectedChannel: rejectedChannel,
		FinishChannel: finished,
	}

	return c, &battleChannels, nil
}

func (client *BattleLobbyClient) RejectChallenge(authToken, battleId, serverHostname string) error {
	addr := fmt.Sprintf("%s:%d", serverHostname, utils.BattlesPort)

	req, err := BuildRequest("POST", addr, fmt.Sprintf(api.RejectChallengePath, battleId), nil)
	if err != nil {
		return errors.WrapRejectBattleChallengeError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	_, err = DoRequest(&http.Client{}, req, nil)
	if err != nil {
		return errors.WrapRejectBattleChallengeError(err)
	}

	return nil
}
