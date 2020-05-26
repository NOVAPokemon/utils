package clients

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/clients/errors"
	"github.com/NOVAPokemon/utils/tokens"
	ws "github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/trades"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
)

type TradeLobbyClient struct {
	TradesAddr string
	config     utils.TradesClientConfig
	conn       *websocket.Conn
}

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
		TradesAddr: tradesURL,
		config:     config,
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

	started := make(chan struct{})
	rejected := make(chan struct{})
	finished := make(chan struct{})
	setItemsToken := make(chan *string)
	writeChannel := make(chan ws.GenericMsg)

	go func() {
		if err := client.HandleReceivedMessages(conn, started, rejected, finished, setItemsToken); err != nil {
			log.Error(err)
		}
	}()

	itemIds := make([]string, len(items.Items))
	i := 0
	for k := range items.Items {
		itemIds[i] = k
		i++
	}

	timeTook := WaitForStart(started, rejected, finished, requestTimestamp)
	log.Infof("time took to initiate interaction: %d ms", timeTook)

	numberMeasuresStart++
	totalTimeTookStart += timeTook
	log.Infof("average time starting: %f ms", float64(totalTimeTookStart)/float64(numberMeasuresStart))

	select {
	case <-rejected:
		log.Infof("trade was rejected")
		return nil, nil
	default:
	}

	go client.autoTrader(itemIds, writeChannel, finished)

	MainLoop(conn, writeChannel, finished)

	log.Info("Finishing trade...")

	return <-setItemsToken, nil
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

func (client *TradeLobbyClient) HandleReceivedMessages(conn *websocket.Conn, started, rejected, finished chan struct{},
	setItemsToken chan *string) error {
	var itemsToken *string

	for {
		msg, err := Read(conn)
		if err != nil {
			return errors.WrapHandleMessagesTradeError(err)
		}

		switch msg.MsgType {
		case ws.Start:
			close(started)
		case ws.Reject:
			close(rejected)
			return nil
		case trades.Update:
			desMsg, err := trades.DeserializeTradeMessage(msg)
			if err != nil {
				log.Error(errors.WrapHandleMessagesTradeError(err))
				continue
			}

			updateMsg := desMsg.(*trades.UpdateMessage)
			updateMsg.Receive(ws.MakeTimestamp())

			timeTook, ok := updateMsg.TimeTook()
			if ok {
				totalTimeTookTradeMsgs += timeTook
				numberMeasuresTradeMsgs++

				log.Infof("time took: %d ms", timeTook)
				log.Infof("average time for trade msgs: %f ms",
					float64(totalTimeTookTradeMsgs)/float64(numberMeasuresTradeMsgs))
			}

			updateMsg.LogReceive(trades.Update)
		case ws.SetToken:
			desMsg, err := trades.DeserializeTradeMessage(msg)
			if err != nil {
				log.Error(errors.WrapHandleMessagesTradeError(err))
				continue
			}

			tokenMessage := desMsg.(*ws.SetTokenMessage)
			token, err := tokens.ExtractItemsToken(tokenMessage.TokensString[0])
			if err != nil {
				log.Error(errors.WrapHandleMessagesTradeError(err))
				continue
			}

			itemsToken = &tokenMessage.TokensString[0]
			log.Info(token.ItemsHash)
		case ws.Finish:
			desMsg, err := trades.DeserializeTradeMessage(msg)
			if err != nil {
				log.Error(errors.WrapHandleMessagesTradeError(err))
				continue
			}

			finishMsg := desMsg.(*ws.FinishMessage)
			log.Info("Finished, Success: ", finishMsg.Success)
			close(finished)
			setItemsToken <- itemsToken
			return errors.WrapHandleMessagesTradeError(err)
		}
	}
}

func (client *TradeLobbyClient) autoTrader(availableItems []string, writeChannel chan ws.GenericMsg,
	finished chan struct{}) {
	select {
	case <-finished:
		return
	default:
		numItems := len(availableItems)

		var maxItemsToTrade int
		if client.config.MaxItemsToTrade < 0 || client.config.MaxItemsToTrade > numItems {
			maxItemsToTrade = numItems
		} else if client.config.MaxItemsToTrade <= numItems {
			maxItemsToTrade = client.config.MaxItemsToTrade
		}

		numItemsToAdd := rand.Intn(maxItemsToTrade)
		log.Infof("will trade %d items", numItemsToAdd)

		for i := 0; i < numItemsToAdd; i++ {
			randomItemIdx := rand.Intn(len(availableItems))
			tradeMsg := trades.NewTradeMessage(availableItems[randomItemIdx])
			// TODO Is it too soon to emit the log?
			tradeMsg.LogEmit(trades.Trade)
			msg := tradeMsg.SerializeToWSMessage()
			s := (*msg).Serialize()
			writeChannel <- ws.GenericMsg{MsgType: websocket.TextMessage, Data: []byte(s)}

			log.Infof("adding %s to trade", availableItems[randomItemIdx])

			availableItems[randomItemIdx] = availableItems[len(availableItems)-1]
			availableItems = availableItems[:len(availableItems)-1]

			if client.config.MaxSleepTime > 0 {
				randSleep := rand.Intn(client.config.MaxSleepTime)
				time.Sleep(time.Duration(randSleep) * time.Millisecond)
				log.Infof("sleeping %d milliseconds", randSleep)
			}

		}

		acceptMsg := trades.NewAcceptMessage()
		// TODO Is it too soon to emit the log?
		acceptMsg.LogEmit(trades.Accept)
		msg := acceptMsg.SerializeToWSMessage()
		s := (*msg).Serialize()
		writeChannel <- ws.GenericMsg{MsgType: websocket.TextMessage, Data: []byte(s)}
		<-finished
	}
}
