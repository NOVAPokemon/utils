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
	"time"
)

type TradeLobbyClient struct {
	TradesAddr string

	conn *websocket.Conn
}

func NewTradesClient(addr string) *TradeLobbyClient {
	return &TradeLobbyClient{
		TradesAddr: addr,
	}
}

func (client *TradeLobbyClient) GetAvailableLobbies() []utils.Lobby {
	req, err := BuildRequest("GET", client.TradesAddr, api.GetTradesPath, nil)
	if err != nil {
		log.Error(err)
		return nil
	}

	var tradesArray []utils.Lobby
	_, err = DoRequest(&http.Client{}, req, &tradesArray)
	if err != nil {
		log.Error(err)
		return nil
	}

	return tradesArray
}

func (client *TradeLobbyClient) CreateTradeLobby(username string, authToken string, itemsToken string) *primitive.ObjectID {
	body := api.CreateLobbyRequest{Username: username}
	req, err := BuildRequest("POST", client.TradesAddr, api.StartTradePath, &body)
	if err != nil {
		log.Error(err)
		return nil
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)
	req.Header.Set(tokens.ItemsTokenHeaderName, itemsToken)

	var lobbyIdHex string
	_, err = DoRequest(&http.Client{}, req, &lobbyIdHex)
	if err != nil {
		log.Error(err)
		return nil
	}

	lobbyId, err := primitive.ObjectIDFromHex(lobbyIdHex)
	if err != nil {
		log.Error(err)
		return nil
	}

	return &lobbyId
}

func (client *TradeLobbyClient) JoinTradeLobby(tradeId *primitive.ObjectID, authToken string, itemsToken string) *string {
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
		log.Fatal(err)
	}

	defer websockets.CloseConnection(conn)
	client.conn = conn

	items, err := tokens.ExtractItemsToken(itemsToken)
	if err != nil {
		log.Error(err)
		return nil
	}

	started := make(chan struct{})
	finished := make(chan struct{})
	setItemsToken := make(chan *string)
	writeChannel := make(chan *string)

	go client.HandleReceivedMessages(conn, started, finished, setItemsToken)

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

	return <-setItemsToken
}

func (client *TradeLobbyClient) HandleReceivedMessages(conn *websocket.Conn, started, finished chan struct{},
	setItemsToken chan *string) {
	var itemsToken *string = nil

	for {
		msg, err := ReadMessagesWithoutParse(conn)
		if err != nil {
			log.Error(err)
			return
		}

		log.Infof("Message: %s", msg)

		switch msg.MsgType {
		case trades.START:
			close(started)
		case trades.SETTOKEN:
			tokenMessage := trades.Deserialize(msg).(*trades.SetTokenMessage)
			token, err := tokens.ExtractItemsToken(tokenMessage.TokenString)
			if err != nil {
				log.Error(err)
			}
			itemsToken = &tokenMessage.TokenString
			log.Info(token.ItemsHash)
		case trades.FINISH:
			finishMsg := trades.Deserialize(msg).(*trades.FinishMessage)
			log.Info("Finished, Success: ", finishMsg.Success)
			close(finished)
			setItemsToken <- itemsToken
			return
		}
	}
}

func (client *TradeLobbyClient) autoTrader(availableItems []string, writeChannel chan *string, finished chan struct{}) {
	select {
	case <-finished:
		return
	default:
		log.Infof("got %d items", len(availableItems))

		numItemsToAdd := rand.Intn(len(availableItems))
		log.Infof("will trade %d items", numItemsToAdd)

		for i := 0; i < numItemsToAdd; i++ {
			randomItemIdx := rand.Intn(len(availableItems))
			msg := trades.TradeMessage{ItemId: availableItems[randomItemIdx]}.SerializeToWSMessage()
			s := (*msg).Serialize()
			writeChannel <- &s

			log.Infof("adding %s to trade", availableItems[randomItemIdx])

			availableItems[randomItemIdx] = availableItems[len(availableItems)-1]
			availableItems = availableItems[:len(availableItems)-1]

			randSleep := rand.Intn(1000) + 1000
			time.Sleep(time.Duration(randSleep) * time.Millisecond)

			log.Infof("sleeping %d milliseconds", randSleep)
		}

		msg := trades.AcceptMessage{}.SerializeToWSMessage()
		s := (*msg).Serialize()
		writeChannel <- &s

		<-finished
	}
}
