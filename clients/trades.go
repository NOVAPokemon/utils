package clients

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/tokens"
	"github.com/NOVAPokemon/utils/websockets"
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

var defaultTradesURL = fmt.Sprintf("%s:%d", utils.Host, utils.TradesPort)

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
		return nil, wrapGetTradeLobbiesError(err)
	}

	var tradesArray []utils.Lobby
	_, err = DoRequest(&http.Client{}, req, &tradesArray)
	if err != nil {
		return nil, wrapGetTradeLobbiesError(err)
	}

	return tradesArray, nil
}

func (client *TradeLobbyClient) CreateTradeLobby(username string, authToken string,
	itemsToken string) (*primitive.ObjectID, error) {
	body := api.CreateLobbyRequest{Username: username}
	req, err := BuildRequest("POST", client.TradesAddr, api.StartTradePath, &body)
	if err != nil {
		return nil, wrapCreateTradeLobbyError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)
	req.Header.Set(tokens.ItemsTokenHeaderName, itemsToken)

	var lobbyIdHex string
	_, err = DoRequest(&http.Client{}, req, &lobbyIdHex)
	if err != nil {
		return nil, wrapCreateTradeLobbyError(err)
	}

	lobbyId, err := primitive.ObjectIDFromHex(lobbyIdHex)
	if err != nil {
		return nil, wrapCreateTradeLobbyError(err)
	}

	return &lobbyId, nil
}

func (client *TradeLobbyClient) JoinTradeLobby(tradeId *primitive.ObjectID, authToken string,
	itemsToken string) (*string, error) {
	u := url.URL{Scheme: "ws", Host: client.TradesAddr, Path: fmt.Sprintf(api.JoinTradePath, tradeId.Hex())}
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
		return nil, wrapJoinTradeLobbyError(err)
	}

	defer websockets.CloseConnection(conn)
	client.conn = conn

	items, err := tokens.ExtractItemsToken(itemsToken)
	if err != nil {
		return nil, wrapJoinTradeLobbyError(err)
	}

	started := make(chan struct{})
	finished := make(chan struct{})
	setItemsToken := make(chan *string)
	writeChannel := make(chan *string)

	go func() {
		if err := client.HandleReceivedMessages(conn, started, finished, setItemsToken); err != nil{
			log.Error(err)
		}
	}()

	itemIds := make([]string, len(items.Items))
	i := 0
	for k := range items.Items {
		itemIds[i] = k
		i++
	}

	WaitForStart(started, finished)

	go client.autoTrader(itemIds, writeChannel, finished)

	MainLoop(conn, writeChannel, finished)

	log.Info("Finishing trade...")

	return <-setItemsToken, nil
}

func (client *TradeLobbyClient) HandleReceivedMessages(conn *websocket.Conn, started, finished chan struct{},
	setItemsToken chan *string) error {
	var itemsToken *string = nil

	for {
		msg, err := Read(conn)
		if err != nil {
			return wrapHandleMessagesTradeError(err)
		}

		switch msg.MsgType {
		case trades.Start:
			close(started)
		case trades.Update:
			updateMsg := trades.DeserializeTradeMessage(msg).(*trades.UpdateMessage)
			updateMsg.Receive(websockets.MakeTimestamp())
			updateMsg.LogReceive(trades.Update)
		case trades.SetToken:
			tokenMessage := trades.DeserializeTradeMessage(msg).(*trades.SetTokenMessage)
			token, err := tokens.ExtractItemsToken(tokenMessage.TokenString)
			if err != nil {
				log.Error(wrapHandleMessagesTradeError(err))
			}

			itemsToken = &tokenMessage.TokenString
			log.Info(token.ItemsHash)
		case trades.Finish:
			finishMsg := trades.DeserializeTradeMessage(msg).(*trades.FinishMessage)
			log.Info("Finished, Success: ", finishMsg.Success)
			close(finished)
			setItemsToken <- itemsToken
			return wrapHandleMessagesTradeError(err)
		}
	}
}

func (client *TradeLobbyClient) autoTrader(availableItems []string, writeChannel chan *string,
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
			randomItemIdx := rand.Intn(numItems)
			tradeMsg := trades.NewTradeMessage(availableItems[randomItemIdx])
			//TODO Is it too soon to emit the log?
			tradeMsg.LogEmit(trades.Trade)
			msg := tradeMsg.SerializeToWSMessage()
			s := (*msg).Serialize()
			writeChannel <- &s

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
		//TODO Is it too soon to emit the log?
		acceptMsg.LogEmit(trades.Accept)
		msg := acceptMsg.SerializeToWSMessage()
		s := (*msg).Serialize()
		writeChannel <- &s

		<-finished
	}
}
