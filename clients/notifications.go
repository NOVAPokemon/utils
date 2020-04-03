package clients

import (
	"encoding/json"
	"fmt"
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
	jar                  *cookiejar.Jar
	httpClient           *http.Client
	NotificationsChannel chan *utils.Notification
}

func NewNotificationClient(addr string, jar *cookiejar.Jar, notificationsChannel chan *utils.Notification) *NotificationClient {
	return &NotificationClient{
		NotificationsAddr: addr,
		jar:               jar,
		httpClient: &http.Client{
			Jar: jar,
		},
		NotificationsChannel: notificationsChannel,
	}
}

func (client *NotificationClient) ListenToNotifications() {
	u := url.URL{Scheme: "ws", Host: client.NotificationsAddr, Path: api.SubscribeNotificationPath}

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		Jar:              client.jar,
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

		log.Info(client.NotificationsChannel)

		client.NotificationsChannel <- &notification

		log.Infof("Received %s from the websocket", notification.Content)
	}
}

func (client *NotificationClient) AddNotification(notification utils.Notification) error {
	req, err := BuildRequest("POST", client.NotificationsAddr, api.NotificationPath, notification)
	if err != nil {
		return err
	}
	err = DoRequest(client.httpClient, req, nil)
	return err
}

func (client *NotificationClient) GetOthersListening(myUsername string) ([]string, error) {
	req, err := BuildRequest("GET", client.NotificationsAddr, fmt.Sprintf(api.GetListenersPath, myUsername), nil)
	if err != nil {
		return nil, err
	}

	var usernames []string
	err = DoRequest(client.httpClient, req, &usernames)
	return usernames, err
}

func (client *NotificationClient) SetJar(jar *cookiejar.Jar) {
	client.jar = jar
	client.httpClient.Jar = jar
}
