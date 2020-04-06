package clients

import (
	"encoding/json"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/tokens"
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
	readChannel          chan []byte
}

func NewNotificationClient(addr string, notificationsChannel chan *utils.Notification) *NotificationClient {
	return &NotificationClient{
		NotificationsAddr:    addr,
		httpClient:           &http.Client{},
		NotificationsChannel: notificationsChannel,
		readChannel:          make(chan []byte),
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

	c, _, err := dialer.Dial(u.String(), header)
	defer c.Close()

	if err != nil {
		log.Fatal(err)
		return
	}

	go client.handleRecv(c)

Loop:
	for {
		select {
		case jsonBytes := <-client.readChannel:
			client.parseToNotification(jsonBytes)
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
		_, jsonBytes, err := conn.ReadMessage()
		if err != nil {
			log.Error(err)
			return
		}

		client.readChannel <- jsonBytes
	}
}

func (client *NotificationClient) parseToNotification(jsonBytes []byte) {
	var notification utils.Notification
	err := json.Unmarshal(jsonBytes, &notification)
	if err != nil {
		log.Error(err)
		return
	}

	client.NotificationsChannel <- &notification

	log.Infof("Received %s from the websocket", notification.Content)
}
