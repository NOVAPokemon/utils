package clients

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/clients/errors"
	"github.com/NOVAPokemon/utils/tokens"
	ws "github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/trades"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TradeLobbyClient struct {
	TradesAddr   string
	config       utils.TradesClientConfig
	conn         *websocket.Conn
	started      chan struct{}
	rejected     chan struct{}
	finished     chan struct{}
	finishOnce   sync.Once
	readChannel  chan *ws.WebsocketMsg
	writeChannel chan *ws.WebsocketMsg
	commsManager ws.CommunicationManager
}

var (
	defaultTradesURL = fmt.Sprintf("%s:%d", utils.Host, utils.TradesPort)
)

func NewTradesClient(config utils.TradesClientConfig, manager ws.CommunicationManager) *TradeLobbyClient {
	tradesURL, exists := os.LookupEnv(utils.TradesEnvVar)

	if !exists {
		log.Warn("missing ", utils.TradesEnvVar)
		tradesURL = defaultTradesURL
	}

	return &TradeLobbyClient{
		TradesAddr:   tradesURL,
		config:       config,
		commsManager: manager,
	}
}

func (client *TradeLobbyClient) GetAvailableLobbies() ([]utils.Lobby, error) {
	req, err := BuildRequest("GET", client.TradesAddr, api.GetTradesPath, nil)
	if err != nil {
		return nil, errors.WrapGetTradeLobbiesError(err)
	}

	var tradesArray []utils.Lobby
	_, err = DoRequest(&http.Client{}, req, &tradesArray, client.commsManager)
	if err != nil {
		return nil, errors.WrapGetTradeLobbiesError(err)
	}

	return tradesArray, nil
}

func (client *TradeLobbyClient) CreateTradeLobby(username string, authToken string,
	itemsToken string) (*primitive.ObjectID, *string, error) {
	body := api.CreateLobbyRequest{Username: username}
	req, err := BuildRequest("POST", client.TradesAddr, api.StartTradePath, &body)
	if err != nil {
		return nil, nil, errors.WrapCreateTradeLobbyError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)
	req.Header.Set(tokens.ItemsTokenHeaderName, itemsToken)
	trackInfo := ws.NewTrackedInfo(primitive.NewObjectID())
	trackInfo.Emit(ws.MakeTimestamp())
	trackInfo.LogEmit(trades.CreateTrade)
	req.Header.Set(ws.TrackInfoHeaderName, trackInfo.SerializeToJSON())

	var resp api.CreateLobbyResponse
	_, err = DoRequest(&http.Client{}, req, &resp, client.commsManager)
	if err != nil {
		return nil, nil, errors.WrapCreateTradeLobbyError(err)
	}

	lobbyId, err := primitive.ObjectIDFromHex(resp.LobbyId)
	if err != nil {
		return nil, nil, errors.WrapCreateTradeLobbyError(err)
	}

	return &lobbyId, &resp.ServerName, nil
}

func (client *TradeLobbyClient) JoinTradeLobby(tradeId *primitive.ObjectID,
	serverHostname string, authToken string, itemsToken string) (*string, error) {
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", serverHostname, utils.TradesPort),
		Path:   fmt.Sprintf(api.JoinTradePath, tradeId.Hex()),
	}
	log.Infof("Connecting to: %s", u.String())

	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)
	header.Set(tokens.ItemsTokenHeaderName, itemsToken)

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		return nil, errors.WrapJoinTradeLobbyError(err)
	}

	defer func() {
		if err = conn.Close(); err != nil {
			log.Error(err)
		}
	}()
	client.conn = conn

	items, err := tokens.ExtractItemsToken(itemsToken)
	if err != nil {
		return nil, errors.WrapJoinTradeLobbyError(err)
	}

	client.started = make(chan struct{})
	client.rejected = make(chan struct{})
	client.finished = make(chan struct{})
	client.finishOnce = sync.Once{}
	client.readChannel = make(chan *ws.WebsocketMsg, 10)
	client.writeChannel = make(chan *ws.WebsocketMsg, 10)

	go ReadMessagesFromConnToChan(conn, client.readChannel, client.finished, client.commsManager)

	itemIds := make([]string, len(items.Items))
	i := 0
	for k := range items.Items {
		itemIds[i] = k
		i++
	}

	client.WaitForStart()
	select {
	case <-client.rejected:
		log.Infof("trade was rejected")
		return nil, nil
	case <-client.finished:
		log.Warn("session finished before starting")
		return nil, nil
	default:
		break
	}

	go WriteTextMessagesFromChanToConn(conn, client.commsManager, client.writeChannel, client.finished)

	itemTokens, err := client.autoTrader(itemIds)

	log.Info("Finishing trade...")

	return itemTokens, errors.WrapJoinTradeLobbyError(err)
}

func (client *TradeLobbyClient) WaitForStart() {
	log.Info("waiting for start...")

	initialMessage, ok := <-client.readChannel

	if ok {
		_, err := client.HandleReceivedMessage(initialMessage)
		if err != nil {
			log.Error(err)
		}
	}

	select {
	case <-client.started:
	case <-client.rejected:
	case <-client.readChannel:
		client.finishOnce.Do(func() { close(client.finished) })
	case <-client.writeChannel:
		client.finishOnce.Do(func() { close(client.finished) })
	case <-client.finished:
	}

}

func (client *TradeLobbyClient) RejectTrade(lobbyId *primitive.ObjectID, serverHostname, authToken,
	itemsToken string) error {
	addr := fmt.Sprintf("%s:%d", serverHostname, utils.TradesPort)

	req, err := BuildRequest("POST", addr, fmt.Sprintf(api.RejectTradePath, lobbyId.Hex()), nil)
	if err != nil {
		return errors.WrapRejectTradeLobbyError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)
	req.Header.Set(tokens.ItemsTokenHeaderName, itemsToken)

	_, err = DoRequest(&http.Client{}, req, nil, client.commsManager)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("got status code %d", http.StatusNotFound)) {
			log.Warn(errors.WrapRejectTradeLobbyError(err))
			return nil
		}
		return errors.WrapRejectTradeLobbyError(err)
	}

	return nil
}

func (client *TradeLobbyClient) HandleReceivedMessage(wsMsg *ws.WebsocketMsg) (*string, error) {

	wsMsgContent := wsMsg.Content

	msgData := wsMsg.Content.Data

	switch wsMsgContent.AppMsgType {
	case trades.StartTrade:
		close(client.started)
	case trades.RejectTrade:
		close(client.rejected)
	case trades.Update:
		updateMsg := &trades.UpdateMessage{}
		if err := mapstructure.Decode(msgData, updateMsg); err != nil {
			panic(err)
		}
		log.Debugf("%+v ", updateMsg)
	case ws.SetToken:
		tokenMessage := &ws.SetTokenMessage{}
		if err := mapstructure.Decode(msgData, tokenMessage); err != nil {
			panic(err)
		}
		_, err := tokens.ExtractItemsToken(tokenMessage.TokensString[0])
		if err != nil {
			log.Error(errors.WrapHandleMessagesTradeError(err))
		}

		return &tokenMessage.TokensString[0], nil
	case ws.Finish:
		finishMsg := &ws.FinishMessage{}
		if err := mapstructure.Decode(msgData, finishMsg); err != nil {
			panic(err)
		}
		log.Info("Finished, Success: ", finishMsg.Success)
		client.finishOnce.Do(func() { close(client.finished) })
	}

	return nil, nil
}

func (client *TradeLobbyClient) autoTrader(availableItems []string) (*string, error) {
	var finalItemTokens *string

	numItems := len(availableItems)

	var maxItemsToTrade int
	if client.config.MaxItemsToTrade < 0 || client.config.MaxItemsToTrade > numItems {
		maxItemsToTrade = numItems
	} else if client.config.MaxItemsToTrade <= numItems {
		maxItemsToTrade = client.config.MaxItemsToTrade
	}

	var numItemsToAdd int
	if maxItemsToTrade == 0 {
		numItemsToAdd = 0
	} else {
		numItemsToAdd = rand.Intn(maxItemsToTrade)
	}

	log.Infof("will trade %d items", numItemsToAdd)

	syncChannel := make(chan struct{})

	go client.sendTradeMessages(numItemsToAdd, availableItems, syncChannel)

	for {
		select {
		case <-client.finished:
			<-syncChannel
			return finalItemTokens, nil
		case msgString, ok := <-client.readChannel:
			if !ok {
				select {
				case <-client.finished:
					return finalItemTokens, nil
				default:
					return nil, nil
				}
			}

			itemTokens, err := client.HandleReceivedMessage(msgString)
			if err != nil {
				return nil, err
			}

			if itemTokens != nil {
				log.Info("updated tokens")
				finalItemTokens = itemTokens
			}
		}
	}
}

func (client *TradeLobbyClient) sendTradeMessages(numItemsToAdd int, availableItems []string,
	syncChannel chan<- struct{}) {
	itemsTraded := 0

	timer := client.setTimerRandSleepTime(nil)
	if numItemsToAdd == 0 {
		if !timer.Stop() {
			<-timer.C
		}

		client.writeChannel <- trades.AcceptMessage{}.ConvertToWSMessage()
	}

	for {
		select {
		case <-client.finished:
			close(syncChannel)
			return
		case <-timer.C:
			randomItemIdx := rand.Intn(len(availableItems))

			client.writeChannel <- trades.TradeMessage{
				ItemId: availableItems[randomItemIdx],
			}.ConvertToWSMessage()

			log.Infof("adding %s to trade", availableItems[randomItemIdx])

			availableItems[randomItemIdx] = availableItems[len(availableItems)-1]
			availableItems = availableItems[:len(availableItems)-1]

			itemsTraded++

			if itemsTraded < numItemsToAdd {
				client.setTimerRandSleepTime(timer)
			} else {
				client.writeChannel <- trades.AcceptMessage{}.ConvertToWSMessage()
			}
		}
	}
}

func (client *TradeLobbyClient) setTimerRandSleepTime(timer *time.Timer) *time.Timer {
	var randSleep int
	if client.config.ThinkTime > 0 {
		randSleep = rand.Intn(client.config.ThinkTime)
	}

	log.Infof("sleeping %d milliseconds", randSleep)

	if timer == nil {
		timer = time.NewTimer(time.Duration(randSleep) * time.Millisecond)
		return timer
	} else {
		timer.Reset(time.Duration(randSleep) * time.Millisecond)
		return nil
	}

}
