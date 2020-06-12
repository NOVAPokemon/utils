package clients

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
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
	readChannel  chan *string
	writeChannel chan ws.GenericMsg
}

const (
	logTimeTookStartTrade    = "time start trade: %d ms"
	logAverageTimeStartTrade = "average start trade: %f ms"

	logTimeTookTradeMsg    = "time trade: %d ms"
	logAverageTimeTradeMsg = "average trade: %f ms"
)

var (
	defaultTradesURL = fmt.Sprintf("%s:%d", utils.Host, utils.TradesPort)

	totalTimeTookStart  int64 = 0
	numberMeasuresStart       = 0

	totalTimeTookTradeMsgs  int64 = 0
	numberMeasuresTradeMsgs       = 0
)

func NewTradesClient(config utils.TradesClientConfig) *TradeLobbyClient {
	tradesURL, exists := os.LookupEnv(utils.TradesEnvVar)

	if !exists {
		log.Warn("missing ", utils.TradesEnvVar)
		tradesURL = defaultTradesURL
	}

	return &TradeLobbyClient{
		TradesAddr:   tradesURL,
		config:       config,
		started:      make(chan struct{}),
		rejected:     make(chan struct{}),
		finished:     make(chan struct{}),
		readChannel:  make(chan *string),
		writeChannel: make(chan ws.GenericMsg),
	}
}

func (client *TradeLobbyClient) GetAvailableLobbies() ([]utils.Lobby, error) {
	req, err := BuildRequest("GET", client.TradesAddr, api.GetTradesPath, nil)
	if err != nil {
		return nil, errors.WrapGetTradeLobbiesError(err)
	}

	var tradesArray []utils.Lobby
	_, err = DoRequest(&http.Client{}, req, &tradesArray)
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

	var resp api.CreateLobbyResponse
	_, err = DoRequest(&http.Client{}, req, &resp)
	if err != nil {
		return nil, nil, errors.WrapCreateTradeLobbyError(err)
	}

	lobbyId, err := primitive.ObjectIDFromHex(resp.LobbyId)
	if err != nil {
		return nil, nil, errors.WrapCreateTradeLobbyError(err)
	}

	return &lobbyId, &resp.ServerName, nil
}

func (client *TradeLobbyClient) JoinTradeLobby(tradeId *primitive.ObjectID, serverHostname string, authToken string,
	itemsToken string) (*string, error) {
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

	requestTimestamp := ws.MakeTimestamp()

	defer ws.CloseConnection(conn)
	client.conn = conn

	items, err := tokens.ExtractItemsToken(itemsToken)
	if err != nil {
		return nil, errors.WrapJoinTradeLobbyError(err)
	}

	go ReadMessagesFromConnToChan(conn, client.readChannel, client.finished)

	itemIds := make([]string, len(items.Items))
	i := 0
	for k := range items.Items {
		itemIds[i] = k
		i++
	}

	timeTook := client.WaitForStart(client.started, client.rejected, client.finished, client.readChannel,
		requestTimestamp)
	if timeTook != -1 {
		log.Infof(logTimeTookStartTrade, timeTook)

		numberMeasuresStart++
		totalTimeTookStart += timeTook
		log.Infof(logAverageTimeStartTrade, float64(totalTimeTookStart)/float64(numberMeasuresStart))
	}

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

	go WriteMessagesFromChanToConn(conn, client.writeChannel, client.finished)

	itemTokens, err := client.autoTrader(itemIds)

	log.Info("Finishing trade...")

	return itemTokens, errors.WrapJoinTradeLobbyError(err)
}

func (client *TradeLobbyClient) WaitForStart(started, rejected, finished chan struct{}, readChannel chan *string,
	requestTimestamp int64) int64 {
	var responseTimestamp int64

	log.Info("waiting for start...")

	initialMessage, ok := <-readChannel

	if ok {
		_, err := client.HandleReceivedMessage(initialMessage)
		if err != nil {
			close(finished)
		}
	} else {
		close(finished)
	}

	select {
	case <-started:
		responseTimestamp = ws.MakeTimestamp()
	case <-rejected:
		responseTimestamp = ws.MakeTimestamp()
	case <-finished:
		return -1
	}

	return responseTimestamp - requestTimestamp
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

	_, err = DoRequest(&http.Client{}, req, nil)
	if err != nil {
		return errors.WrapRejectTradeLobbyError(err)
	}

	return nil
}

func (client *TradeLobbyClient) HandleReceivedMessage(msgString *string) (*string, error) {

	msg, err := ws.ParseMessage(msgString)
	if err != nil {
		return nil, errors.WrapHandleMessagesTradeError(err)
	}

	switch msg.MsgType {
	case ws.Start:
		close(client.started)
	case ws.Reject:
		close(client.rejected)
	case trades.Update:
		desMsg, err := trades.DeserializeTradeMessage(msg)
		if err != nil {
			log.Error(errors.WrapHandleMessagesTradeError(err))
			break
		}

		updateMsg := desMsg.(*trades.UpdateMessage)
		updateMsg.Receive(ws.MakeTimestamp())

		timeTook, ok := updateMsg.TimeTook()
		if ok {
			totalTimeTookTradeMsgs += timeTook
			numberMeasuresTradeMsgs++
			log.Infof(logTimeTookTradeMsg, timeTook)
			log.Infof(logAverageTimeTradeMsg,
				float64(totalTimeTookTradeMsgs)/float64(numberMeasuresTradeMsgs))
		}

		updateMsg.LogReceive(trades.Update)
	case ws.SetToken:
		desMsg, err := trades.DeserializeTradeMessage(msg)
		if err != nil {
			log.Error(errors.WrapHandleMessagesTradeError(err))
			break
		}

		tokenMessage := desMsg.(*ws.SetTokenMessage)
		token, err := tokens.ExtractItemsToken(tokenMessage.TokensString[0])
		if err != nil {
			log.Error(errors.WrapHandleMessagesTradeError(err))

		}

		log.Info(token.ItemsHash)
		return &tokenMessage.TokensString[0], nil
	case ws.Finish:
		desMsg, err := trades.DeserializeTradeMessage(msg)
		if err != nil {
			log.Error(errors.WrapHandleMessagesTradeError(err))
			break
		}

		finishMsg := desMsg.(*ws.FinishMessage)
		log.Info("Finished, Success: ", finishMsg.Success)
		close(client.finished)
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

	itemsTraded := 0

	timer := client.setTimerRandSleepTime(nil)

	for {
		select {
		case <-client.finished:
			return finalItemTokens, nil
		case msgString, ok := <-client.readChannel:
			if !ok {
				close(client.finished)
				break
			}

			itemTokens, err := client.HandleReceivedMessage(msgString)
			if err != nil {
				return nil, err
			}

			if itemTokens != nil {
				finalItemTokens = itemTokens
			}
		case <-timer.C:
			randomItemIdx := rand.Intn(len(availableItems))
			tradeMsg := trades.NewTradeMessage(availableItems[randomItemIdx])
			tradeMsg.LogEmit(trades.Trade)
			msg := tradeMsg.SerializeToWSMessage()
			s := (*msg).Serialize()

			client.writeChannel <- ws.GenericMsg{MsgType: websocket.TextMessage, Data: []byte(s)}

			log.Infof("adding %s to trade", availableItems[randomItemIdx])

			availableItems[randomItemIdx] = availableItems[len(availableItems)-1]
			availableItems = availableItems[:len(availableItems)-1]

			if client.config.MaxSleepTime > 0 {
				randSleep := rand.Intn(client.config.MaxSleepTime)
				time.Sleep(time.Duration(randSleep) * time.Millisecond)
				log.Infof("sleeping %d milliseconds", randSleep)
			}

			itemsTraded++

			if itemsTraded < numItemsToAdd {
				client.setTimerRandSleepTime(timer)
			} else {
				acceptMsg := trades.NewAcceptMessage()
				acceptMsg.LogEmit(trades.Accept)
				msg := acceptMsg.SerializeToWSMessage()
				s := (*msg).Serialize()

				client.writeChannel <- ws.GenericMsg{MsgType: websocket.TextMessage, Data: []byte(s)}
			}
		}
	}
}

func (client *TradeLobbyClient) setTimerRandSleepTime(timer *time.Timer) *time.Timer {
	var randSleep int
	if client.config.MaxSleepTime > 0 {
		randSleep = rand.Intn(client.config.MaxSleepTime)
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
