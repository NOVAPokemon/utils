package clients

import (
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/tokens"
	ws "github.com/NOVAPokemon/utils/websockets"
	notificationMessages "github.com/NOVAPokemon/utils/websockets/notifications"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"time"
)

type NotificationClient struct {
	NotificationsAddr    string
	httpClient           *http.Client
	NotificationsChannel chan *utils.Notification
	readChannel          chan *ws.Message
}

func NewNotificationClient(addr string, notificationsChannel chan *utils.Notification) *NotificationClient {
	return &NotificationClient{
		NotificationsAddr:    addr,
		httpClient:           &http.Client{},
		NotificationsChannel: notificationsChannel,
		readChannel:          make(chan *ws.Message),
	}
}

func (client *NotificationClient) ListenToNotifications(authToken string,
	receiveFinish chan struct{}, emitFinish chan bool) {
	u := url.URL{Scheme: "ws", Host: client.NotificationsAddr, Path: api.SubscribeNotificationPath}

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)

	conn, _, err := dialer.Dial(u.String(), header)
	defer ws.CloseConnection(conn)

	if err != nil {
		log.Fatal(err)
		return
	}

	go client.handleRecv(conn)

Loop:
	for {
		select {
		case msg := <-client.readChannel:
			client.parseToNotification(msg)
		case <-receiveFinish:
			break Loop
		}
	}

	log.Info("Stopped listening to notifications...")

	err = client.StopListening(authToken)
	if err != nil {
		log.Error(err)
	}

	emitFinish <- true
}

func (client *NotificationClient) StopListening(authToken string) error {
	req, err := BuildRequest("GET", client.NotificationsAddr, api.UnsubscribeNotificationPath, nil)
	if err != nil {
		return err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	_, err = DoRequest(client.httpClient, req, nil)
	return err
}

func (client *NotificationClient) AddNotification(notification utils.Notification, authToken string) error {
	req, err := BuildRequest("POST", client.NotificationsAddr, api.NotificationPath, notification)
	if err != nil {
		return err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	_, err = DoRequest(client.httpClient, req, nil)
	return err
}

func (client *NotificationClient) GetOthersListening(authToken string) ([]string, error) {
	req, err := BuildRequest("GET", client.NotificationsAddr, api.GetListenersPath, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var usernames []string
	_, err = DoRequest(client.httpClient, req, &usernames)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return usernames, nil
}

func (client *NotificationClient) handleRecv(conn *websocket.Conn) {
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Error(err)
			return
		}

		msgString := string(msgBytes)
		msg, err := ws.ParseMessage(&msgString)
		if err != nil {
			log.Error(err)
			return
		}

		client.readChannel <- msg
	}
}

func (client *NotificationClient) parseToNotification(msg *ws.Message) {
	notificationMsg := notificationMessages.Deserialize(msg).(*notificationMessages.NotificationMessage)
	client.NotificationsChannel <- &notificationMsg.Notification

	log.Infof("Received %s from the websocket", notificationMsg.Notification.Content)
}
