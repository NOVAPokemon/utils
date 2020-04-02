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
}

func NewNotificationClient(addr string, notificationsChannel chan *utils.Notification) *NotificationClient {
	return &NotificationClient{
		NotificationsAddr:    addr,
		httpClient:           &http.Client{},
		NotificationsChannel: notificationsChannel,
	}
}

func (client *NotificationClient) ListenToNotifications() {
	u := url.URL{Scheme: "ws", Host: client.NotificationsAddr, Path: api.SubscribeNotificationPath}

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	c, _, err := dialer.Dial(u.String(), nil)
	defer c.Close()

	if err != nil {
		log.Fatal(err)
		return
	}

	client.readNotifications(c)

	log.Info("Stopped listening to notifications...")
}

func (client *NotificationClient) readNotifications(conn *websocket.Conn) {
	for {
		_, jsonBytes, err := conn.ReadMessage()

		if err != nil {
			log.Error(err)
			return
		}

		var notification utils.Notification
		err = json.Unmarshal(jsonBytes, &notification)
		if err != nil {
			log.Error(err)
			return
		}

		client.NotificationsChannel <- &notification

		log.Debugf("Received %s from the websocket", notification.Content)
	}
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
