package clients

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	http "github.com/bruno-anjos/archimedesHTTPClient"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/clients/errors"
	"github.com/NOVAPokemon/utils/tokens"
	"github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/battles"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type BattleLobbyClient struct {
	BattlesAddr  string
	httpClient   *http.Client
	commsManager websockets.CommunicationManager
}

var defaultBattleURL = fmt.Sprintf("%s:%d", utils.Host, utils.BattlesPort)

func NewBattlesClient(commsManager websockets.CommunicationManager, httpClient *http.Client) *BattleLobbyClient {
	battlesURL, exists := os.LookupEnv(utils.BattlesEnvVar)

	if !exists {
		log.Warn("missing ", utils.BattlesEnvVar)
		battlesURL = defaultBattleURL
	}

	return &BattleLobbyClient{
		BattlesAddr:  battlesURL,
		httpClient:   httpClient,
		commsManager: commsManager,
	}
}

func (client *BattleLobbyClient) GetAvailableLobbies() ([]utils.Lobby, error) {
	resolvedAddr, _, err := client.httpClient.ResolveServiceInArchimedes(client.BattlesAddr)
	if err != nil {
		log.Panic(err)
	}

	u := url.URL{Scheme: "http", Host: resolvedAddr, Path: "/battles"}

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
	resolvedAddr, _, err := client.httpClient.ResolveServiceInArchimedes(client.BattlesAddr)

	u := url.URL{Scheme: "ws", Host: resolvedAddr, Path: api.QueueForBattlePath}
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
	websockets.AddTrackInfoToHeader(&header, battles.Queue)

	c, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		err = errors.WrapQueueForBattleError(websockets.WrapDialingError(err, u.String()))
		return nil, nil, err
	}

	outChannel := make(chan *websockets.WebsocketMsg)
	inChannel := make(chan *websockets.WebsocketMsg)
	rejectedChannel := make(chan struct{})
	finished := make(chan struct{})

	go ReadMessagesFromConnToChan(c, inChannel, finished, client.commsManager)
	go WriteTextMessagesFromChanToConn(c, client.commsManager, outChannel, finished)

	battleChannels := battles.BattleChannels{
		OutChannel:      outChannel,
		InChannel:       inChannel,
		RejectedChannel: rejectedChannel,
		FinishChannel:   finished,
	}

	return c, &battleChannels, nil
}

func (client *BattleLobbyClient) ChallengePlayerToBattle(authToken string, pokemonsTokens []string, statsToken string,
	itemsToken string, targetPlayer string) (*websocket.Conn, *battles.BattleChannels, int64, error) {
	resolvedAddr, _, err := client.httpClient.ResolveServiceInArchimedes(client.BattlesAddr)

	u := url.URL{Scheme: "ws", Host: resolvedAddr, Path: fmt.Sprintf(api.ChallengeToBattlePath, targetPlayer)}
	log.Infof("Connecting to: %s", u.String())

	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)
	header.Set(tokens.StatsTokenHeaderName, statsToken)
	header[tokens.PokemonsTokenHeaderName] = pokemonsTokens
	header.Set(tokens.ItemsTokenHeaderName, itemsToken)
	websockets.AddTrackInfoToHeader(&header, battles.Challenge)

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	requestTimestamp := websockets.MakeTimestamp()
	c, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		err = errors.WrapChallengeForBattleError(websockets.WrapDialingError(err, u.String()))
		return nil, nil, 0, err
	}

	inChannel := make(chan *websockets.WebsocketMsg)
	outChannel := make(chan *websockets.WebsocketMsg)
	rejectedChannel := make(chan struct{})
	finished := make(chan struct{})

	go ReadMessagesFromConnToChan(c, inChannel, finished, client.commsManager)
	go WriteTextMessagesFromChanToConn(c, client.commsManager, outChannel, finished)

	battleChannels := battles.BattleChannels{
		OutChannel:      outChannel,
		InChannel:       inChannel,
		RejectedChannel: rejectedChannel,
		FinishChannel:   finished,
	}

	return c, &battleChannels, requestTimestamp, nil
}

func (client *BattleLobbyClient) AcceptChallenge(authToken string, pokemonsTokens []string, statsToken string,
	itemsToken string, battleId string, serverHostname string) (*websocket.Conn, *battles.BattleChannels, error) {
	resolvedAddr, _, err := client.httpClient.ResolveServiceInArchimedes(fmt.Sprintf("%s:%d", serverHostname,
		utils.BattlesPort))
	if err != nil {
		log.Panic(err)
	}

	u := url.URL{Scheme: "ws", Host: resolvedAddr, Path: fmt.Sprintf(api.AcceptChallengePath, battleId)}
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

	outChannel := make(chan *websockets.WebsocketMsg)
	inChannel := make(chan *websockets.WebsocketMsg)
	rejectedChannel := make(chan struct{})
	finished := make(chan struct{})

	go ReadMessagesFromConnToChan(c, inChannel, finished, client.commsManager)
	go WriteTextMessagesFromChanToConn(c, client.commsManager, outChannel, finished)

	battleChannels := battles.BattleChannels{
		OutChannel:      outChannel,
		InChannel:       inChannel,
		RejectedChannel: rejectedChannel,
		FinishChannel:   finished,
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

	_, err = DoRequest(&http.Client{}, req, nil, client.commsManager)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("got status code %d", http.StatusNotFound)) {
			log.Warn(errors.WrapRejectBattleChallengeError(err))
			return nil
		}
		return errors.WrapRejectBattleChallengeError(err)
	}

	return nil
}
