package clients

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/clients/errors"
	"github.com/NOVAPokemon/utils/tokens"
	ws "github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/comms_manager"
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
	client       *http.Client
	*BasicClient
}

var defaultTradesURL = fmt.Sprintf("%s:%d", utils.Host, utils.TradesPort)

func NewTradesClient(config utils.TradesClientConfig, manager ws.CommunicationManager,
	httpClient *http.Client, client *BasicClient) *TradeLobbyClient {
	tradesURL, exists := os.LookupEnv(utils.TradesEnvVar)

	if !exists {
		log.Warn("missing ", utils.TradesEnvVar)
		tradesURL = defaultTradesURL
	}

	return &TradeLobbyClient{
		TradesAddr:   tradesURL,
		config:       config,
		commsManager: manager,
		client:       httpClient,
		BasicClient:  client,
	}
}

func (t *TradeLobbyClient) GetAvailableLobbies() ([]utils.Lobby, error) {
	req, err := t.BuildRequest("GET", t.TradesAddr, api.GetTradesPath, nil)
	if err != nil {
		return nil, errors.WrapGetTradeLobbiesError(err)
	}

	var tradesArray []utils.Lobby
	_, err = DoRequest(t.client, req, &tradesArray, t.commsManager)
	if err != nil {
		return nil, errors.WrapGetTradeLobbiesError(err)
	}

	return tradesArray, nil
}

func (t *TradeLobbyClient) CreateTradeLobby(username, authToken string,
	itemsToken string) (*primitive.ObjectID, *string, error) {
	body := api.CreateLobbyRequest{Username: username}
	req, err := t.BuildRequest("POST", t.TradesAddr, api.StartTradePath, &body)
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
	_, err = DoRequest(t.client, req, &resp, t.commsManager)
	if err != nil {
		return nil, nil, errors.WrapCreateTradeLobbyError(err)
	}

	lobbyId, err := primitive.ObjectIDFromHex(resp.LobbyId)
	if err != nil {
		return nil, nil, errors.WrapCreateTradeLobbyError(err)
	}

	return &lobbyId, &resp.ServerName, nil
}

func (t *TradeLobbyClient) JoinTradeLobby(tradeId *primitive.ObjectID,
	serverHostname, authToken, itemsToken string) (*string, error) {
	u := url.URL{
		Scheme: "ws",
		Host:   t.TradesAddr,
		Path:   fmt.Sprintf(api.JoinTradePath, tradeId.Hex()),
	}
	log.Infof("Connecting to: %s", u.String())

	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)
	header.Set(tokens.ItemsTokenHeaderName, itemsToken)
	header.Set("Host", serverHostname)

	switch castedManager := t.commsManager.(type) {
	case *comms_manager.S2DelayedCommsManager:
		header.Set(comms_manager.LocationTagKey, castedManager.GetCellID().ToToken())
		header.Set(comms_manager.TagIsClientKey, strconv.FormatBool(true))
		header.Set(comms_manager.ClosestNodeKey, strconv.Itoa(castedManager.MyClosestNode))
	}

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	trackInfo := ws.NewTrackedInfo(primitive.NewObjectID())
	trackInfo.Emit(ws.MakeTimestamp())
	trackInfo.LogEmit(trades.JoinTrade)
	header.Set(ws.TrackInfoHeaderName, trackInfo.SerializeToJSON())

	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		return nil, errors.WrapJoinTradeLobbyError(err)
	}

	defer func() {
		if err = conn.Close(); err != nil {
			log.Error(err)
		}
	}()
	t.conn = conn

	items, err := tokens.ExtractItemsToken(itemsToken)
	if err != nil {
		return nil, errors.WrapJoinTradeLobbyError(err)
	}

	t.started = make(chan struct{})
	t.rejected = make(chan struct{})
	t.finished = make(chan struct{})
	t.finishOnce = sync.Once{}
	t.readChannel = make(chan *ws.WebsocketMsg, chanSize)
	t.writeChannel = make(chan *ws.WebsocketMsg, chanSize)

	go ReadMessagesFromConnToChan(conn, t.readChannel, t.finished, t.commsManager)

	itemIds := make([]string, len(items.Items))
	i := 0
	for k := range items.Items {
		itemIds[i] = k
		i++
	}

	t.WaitForStart()
	select {
	case <-t.rejected:
		log.Infof("trade was rejected")
		return nil, nil
	case <-t.finished:
		log.Warn("session finished before starting")
		return nil, nil
	default:
		break
	}

	go WriteTextMessagesFromChanToConn(conn, t.commsManager, t.writeChannel, t.finished)

	itemTokens, err := t.autoTrader(itemIds)

	log.Info("Finishing trade...")

	return itemTokens, errors.WrapJoinTradeLobbyError(err)
}

func (t *TradeLobbyClient) WaitForStart() {
	log.Info("waiting for start...")

	initialMessage, ok := <-t.readChannel

	if ok {
		_, err := t.HandleReceivedMessage(initialMessage)
		if err != nil {
			log.Error(err)
		}
	}

	select {
	case <-t.started:
	case <-t.rejected:
	case <-t.readChannel:
		t.finishOnce.Do(func() { close(t.finished) })
	case <-t.writeChannel:
		t.finishOnce.Do(func() { close(t.finished) })
	case <-t.finished:
	}
}

func (t *TradeLobbyClient) RejectTrade(lobbyId *primitive.ObjectID, serverHostname, authToken,
	itemsToken string) error {
	req, err := t.BuildRequestForHost("POST", t.TradesAddr, serverHostname, fmt.Sprintf(api.RejectTradePath,
		lobbyId.Hex()), nil)
	if err != nil {
		return errors.WrapRejectTradeLobbyError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)
	req.Header.Set(tokens.ItemsTokenHeaderName, itemsToken)

	_, err = DoRequest(t.client, req, nil, t.commsManager)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("got status code %d", http.StatusNotFound)) {
			log.Warn(errors.WrapRejectTradeLobbyError(err))
			return nil
		}
		return errors.WrapRejectTradeLobbyError(err)
	}

	return nil
}

func (t *TradeLobbyClient) HandleReceivedMessage(wsMsg *ws.WebsocketMsg) (*string, error) {
	wsMsgContent := wsMsg.Content

	msgData := wsMsg.Content.Data

	switch wsMsgContent.AppMsgType {
	case trades.StartTrade:
		close(t.started)
	case trades.RejectTrade:
		close(t.rejected)
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
		t.finishOnce.Do(func() { close(t.finished) })
	}

	return nil, nil
}

func (t *TradeLobbyClient) autoTrader(availableItems []string) (*string, error) {
	var finalItemTokens *string

	numItems := len(availableItems)

	var maxItemsToTrade int
	if t.config.MaxItemsToTrade < 0 || t.config.MaxItemsToTrade > numItems {
		maxItemsToTrade = numItems
	} else if t.config.MaxItemsToTrade <= numItems {
		maxItemsToTrade = t.config.MaxItemsToTrade
	}

	var numItemsToAdd int
	if maxItemsToTrade == 0 {
		numItemsToAdd = 0
	} else {
		numItemsToAdd = rand.Intn(maxItemsToTrade)
	}

	log.Infof("will trade %d items", numItemsToAdd)

	syncChannel := make(chan struct{})

	go t.sendTradeMessages(numItemsToAdd, availableItems, syncChannel)

	for {
		select {
		case <-t.finished:
			<-syncChannel
			return finalItemTokens, nil
		case msgString, ok := <-t.readChannel:
			if !ok {
				select {
				case <-t.finished:
					return finalItemTokens, nil
				default:
					return nil, nil
				}
			}

			itemTokens, err := t.HandleReceivedMessage(msgString)
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

func (t *TradeLobbyClient) sendTradeMessages(numItemsToAdd int, availableItems []string,
	syncChannel chan<- struct{}) {
	itemsTraded := 0

	timer := t.setTimerRandSleepTime(nil)
	if numItemsToAdd == 0 {
		if !timer.Stop() {
			<-timer.C
		}

		t.writeChannel <- trades.AcceptMessage{}.ConvertToWSMessage()
	}

	for {
		select {
		case <-t.finished:
			close(syncChannel)
			return
		case <-timer.C:
			if itemsTraded == numItemsToAdd {
				t.writeChannel <- trades.AcceptMessage{}.ConvertToWSMessage()
				break
			}

			randomItemIdx := rand.Intn(len(availableItems))

			t.writeChannel <- trades.TradeMessage{
				ItemId: availableItems[randomItemIdx],
			}.ConvertToWSMessage()

			log.Infof("adding %s to trade", availableItems[randomItemIdx])

			availableItems[randomItemIdx] = availableItems[len(availableItems)-1]
			availableItems = availableItems[:len(availableItems)-1]

			itemsTraded++

			t.setTimerRandSleepTime(timer)
		}
	}
}

func (t *TradeLobbyClient) setTimerRandSleepTime(timer *time.Timer) *time.Timer {
	var randSleep int
	if t.config.ThinkTime > 0 {
		randSleep = rand.Intn(t.config.ThinkTime)
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
