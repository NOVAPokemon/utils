package clients

import (
	"encoding/json"
	"errors"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/routes"
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
	notificationHandlers map[string]utils.NotificationHandler
}

func (client *NotificationClient) ListenToNotifications() {
	u := url.URL{Scheme: "ws", Host: client.NotificationsAddr, Path: routes.SubscribeNotificationPath}

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
