package clients

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/clients/errors"
	"github.com/NOVAPokemon/utils/tokens"
	"github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/battles"

	"github.com/NOVAPokemon/utils/websockets/comms_manager"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type BattleLobbyClient struct {
	BattlesAddr  string
	httpClient   *http.Client
	commsManager websockets.CommunicationManager
	*BasicClient
}

const (
	chanSize = 10
)

var defaultBattleURL = fmt.Sprintf("%s:%d", utils.Host, utils.BattlesPort)

func NewBattlesClient(commsManager websockets.CommunicationManager, httpClient *http.Client,
	basicClient *BasicClient) *BattleLobbyClient {
	battlesURL, exists := os.LookupEnv(utils.BattlesEnvVar)

	if !exists {
		log.Warn("missing ", utils.BattlesEnvVar)
		battlesURL = defaultBattleURL
	}

	return &BattleLobbyClient{
		BattlesAddr:  battlesURL,
		httpClient:   httpClient,
		commsManager: commsManager,
		BasicClient:  basicClient,
	}
}

func (client *BattleLobbyClient) GetAvailableLobbies() ([]utils.Lobby, error) {
	req, err := client.BuildRequest(http.MethodGet, client.BattlesAddr, "/battles", nil)
	if err != nil {
		return nil, errors.WrapGetBattleLobbiesError(err)
	}

	var availableBattles []utils.Lobby
	_, err = DoRequest(client.httpClient, req, &availableBattles, client.commsManager)
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
		HandshakeTimeout: websockets.Timeout,
	}

	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)
	header.Set(tokens.StatsTokenHeaderName, statsToken)
	header.Set(tokens.ItemsTokenHeaderName, itemsToken)
	header[tokens.PokemonsTokenHeaderName] = pokemonsTokens
	websockets.AddTrackInfoToHeader(&header, battles.Queue)

	switch castedManager := client.commsManager.(type) {
	case *comms_manager.S2DelayedCommsManager:
		header.Set(comms_manager.LocationTagKey, castedManager.GetCellID().ToToken())
		header.Set(comms_manager.TagIsClientKey, strconv.FormatBool(true))
		header.Set(comms_manager.ClosestNodeKey, strconv.Itoa(castedManager.MyClosestNode))
	}

	c, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		err = errors.WrapQueueForBattleError(websockets.WrapDialingError(err, u.String()))
		return nil, nil, err
	}

	outChannel := make(chan *websockets.WebsocketMsg)
	inChannel := make(chan *websockets.WebsocketMsg, chanSize)
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
	itemsToken, targetPlayer string) (*websocket.Conn, *battles.BattleChannels, int64, error) {
	u := url.URL{Scheme: "ws", Host: client.BattlesAddr, Path: fmt.Sprintf(api.ChallengeToBattlePath, targetPlayer)}
	log.Infof("Connecting to: %s", u.String())

	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)
	header.Set(tokens.StatsTokenHeaderName, statsToken)
	header[tokens.PokemonsTokenHeaderName] = pokemonsTokens
	header.Set(tokens.ItemsTokenHeaderName, itemsToken)
	websockets.AddTrackInfoToHeader(&header, battles.Challenge)

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: websockets.Timeout,
	}

	requestTimestamp := websockets.MakeTimestamp()
	c, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		err = errors.WrapChallengeForBattleError(websockets.WrapDialingError(err, u.String()))
		return nil, nil, 0, err
	}

	inChannel := make(chan *websockets.WebsocketMsg, chanSize)
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
	itemsToken, battleId, serverHostname string) (*websocket.Conn, *battles.BattleChannels, error) {
	u := url.URL{
		Scheme: "ws",
		Host:   client.BattlesAddr,
		Path:   fmt.Sprintf(api.AcceptChallengePath, battleId),
	}
	log.Infof("Accepting challenge: %s", u.String())
	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)
	header.Set(tokens.StatsTokenHeaderName, statsToken)
	header[tokens.PokemonsTokenHeaderName] = pokemonsTokens
	header.Set("Host", serverHostname)
	header.Set(tokens.ItemsTokenHeaderName, itemsToken)

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: websockets.Timeout,
	}

	c, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		err = errors.WrapAcceptBattleChallengeError(websockets.WrapDialingError(err, u.String()))
		return nil, nil, err
	}

	outChannel := make(chan *websockets.WebsocketMsg)
	inChannel := make(chan *websockets.WebsocketMsg, chanSize)
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
	req, err := client.BuildRequestForHost("POST", client.BattlesAddr, serverHostname,
		fmt.Sprintf(api.RejectChallengePath, battleId), nil)
	if err != nil {
		return errors.WrapRejectBattleChallengeError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	_, err = DoRequest(client.httpClient, req, nil, client.commsManager)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("got status code %d", http.StatusNotFound)) {
			log.Warn(errors.WrapRejectBattleChallengeError(err))
			return nil
		}
		return errors.WrapRejectBattleChallengeError(err)
	}

	return nil
}
