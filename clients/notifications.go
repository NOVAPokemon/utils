package clients

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/clients/errors"
	"github.com/NOVAPokemon/utils/tokens"
	ws "github.com/NOVAPokemon/utils/websockets"
	notificationMessages "github.com/NOVAPokemon/utils/websockets/notifications"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
)

type NotificationClient struct {
	NotificationsAddr    string
	httpClient           *http.Client
	NotificationsChannel chan utils.Notification
	readChannel          chan *ws.WebsocketMsg
	commsManager         ws.CommunicationManager
}

var (
	defaultNotificationsURL = fmt.Sprintf("%s:%d", utils.Host, utils.NotificationsPort)
)

func NewNotificationClient(notificationsChannel chan utils.Notification,
	manager ws.CommunicationManager) *NotificationClient {
	notificationsURL, exists := os.LookupEnv(utils.NotificationsEnvVar)

	if !exists {
		log.Warn("missing ", utils.NotificationsEnvVar)
		notificationsURL = defaultNotificationsURL
	}

	return &NotificationClient{
		NotificationsAddr:    notificationsURL,
		httpClient:           &http.Client{},
		NotificationsChannel: notificationsChannel,
		readChannel:          make(chan *ws.WebsocketMsg),
		commsManager:         manager,
	}
}

func (client *NotificationClient) ListenToNotifications(authToken string,
	receiveFinish chan struct{}, emitFinish chan bool) error {
	u := url.URL{Scheme: "ws", Host: client.NotificationsAddr, Path: api.SubscribeNotificationPath}

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	log.Info("Dialing: ", u.String())
	header := http.Header{}
	header.Set(tokens.AuthTokenHeaderName, authToken)

	conn, _, err := dialer.Dial(u.String(), header)
	defer func() {
		if err = conn.Close(); err != nil {
			log.Error(err)
		}
	}()

	if err != nil {
		err = errors.WrapListeningNotificationsError(ws.WrapDialingError(err, u.String()))
		return err
	}

	conn.SetPingHandler(func(string) error {
		return client.commsManager.WriteGenericMessageToConn(conn, ws.NewControlMsg(websocket.PongMessage))
	})

	go ReadMessagesFromConnToChan(conn, client.readChannel, receiveFinish, client.commsManager)

Loop:
	for {
		select {
		case wsMsg, ok := <-client.readChannel:
			if ok {
				client.parseToNotification(wsMsg)
			}
		case <-receiveFinish:
			break Loop
		}
	}

	log.Info("Stopped listening to notifications...")

	err = client.StopListening(authToken)
	if err != nil {
		return errors.WrapListeningNotificationsError(err)
	}

	emitFinish <- true
	return nil
}

func (client *NotificationClient) StopListening(authToken string) error {
	req, err := BuildRequest("GET", client.NotificationsAddr, api.UnsubscribeNotificationPath, nil)
	if err != nil {
		return errors.WrapStopListeningError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	_, err = DoRequest(client.httpClient, req, nil, client.commsManager)
	return errors.WrapStopListeningError(err)
}

func (client *NotificationClient) AddNotification(notificationMsg *notificationMessages.NotificationMessage,
	authToken string) error {
	req, err := BuildRequest("POST", client.NotificationsAddr, api.NotificationPath, notificationMsg)
	if err != nil {
		return errors.WrapAddNotificationError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	_, err = DoRequest(client.httpClient, req, nil, client.commsManager)
	return errors.WrapAddNotificationError(err)
}

func (client *NotificationClient) GetOthersListening(authToken string) ([]string, error) {
	req, err := BuildRequest("GET", client.NotificationsAddr, api.GetListenersPath, nil)
	if err != nil {
		return nil, errors.WrapGetOthersListeningError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var usernames []string
	_, err = DoRequest(client.httpClient, req, &usernames, client.commsManager)
	if err != nil {
		return nil, errors.WrapGetOthersListeningError(err)
	}

	return usernames, nil
}

func (client *NotificationClient) parseToNotification(wsMsg *ws.WebsocketMsg) {
	notificationMsg := notificationMessages.NotificationMessage{}
	err := mapstructure.Decode(wsMsg.Content.Data, &notificationMsg)
	if err != nil {
		panic(err)
	}

	client.NotificationsChannel <- notificationMsg.Notification
	log.Infof("Received %s from the websocket", notificationMsg.Notification.Content)
}
