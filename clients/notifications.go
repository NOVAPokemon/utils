package clients

import (
	"encoding/json"
	"errors"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

type NotificationClient struct {
	NotificationsAddr    string
	Jar                  *cookiejar.Jar
	client               *http.Client
	notificationHandlers map[string]utils.NotificationHandler
}

func NewNotificationClient(addr string, jar *cookiejar.Jar) *NotificationClient {
	return &NotificationClient{
		NotificationsAddr:    addr,
		Jar:                  jar,
		client:               &http.Client{},
		notificationHandlers: make(map[string]utils.NotificationHandler, 5),
	}
}

func (client *NotificationClient) ListenToNotifications() {
	u := url.URL{Scheme: "ws", Host: client.NotificationsAddr, Path: api.SubscribeNotificationPath}

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              client.Jar,
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

func (client *NotificationClient) RegisterHandler(notificationType string, handler utils.NotificationHandler) error {
	oldHandler := client.notificationHandlers[notificationType]
	if oldHandler != nil {
		return errors.New("notification already handled")
	}

	client.notificationHandlers[notificationType] = handler
	log.Infof("Registered handler for type: %s", notificationType)

	return nil
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

		handler := client.notificationHandlers[notification.Type]

		if handler == nil {
			log.Errorf("cant handle notification type: %s", notification.Type)
			log.Errorf("%+v", client.notificationHandlers)
			continue
		}

		err = handler(notification)

		if err != nil {
			log.Error(err)
		}

		log.Debugf("Received %s from the websocket", notification.Content)
	}
}

func (c *NotificationClient) AddNotification(notification utils.Notification) error {
	req, err := BuildRequest("POST", c.NotificationsAddr, api.NotificationPath, notification)
	if err != nil {
		return err
	}
	err = DoRequest(c.client, req, nil)
	return err
}
