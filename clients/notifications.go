package clients

import (
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	"github.com/NOVAPokemon/utils/clients/errors"
	"github.com/NOVAPokemon/utils/tokens"
	ws "github.com/NOVAPokemon/utils/websockets"
	notificationMessages "github.com/NOVAPokemon/utils/websockets/notifications"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"time"
)

type NotificationClient struct {
	NotificationsAddr    string
	httpClient           *http.Client
	NotificationsChannel chan *utils.Notification
	readChannel          chan *string
}

var defaultNotificationsURL = fmt.Sprintf("%s:%d", utils.Host, utils.NotificationsPort)

func NewNotificationClient(notificationsChannel chan *utils.Notification) *NotificationClient {
	notificationsURL, exists := os.LookupEnv(utils.NotificationsEnvVar)

	if !exists {
		log.Warn("missing ", utils.NotificationsEnvVar)
		notificationsURL = defaultNotificationsURL
	}

	return &NotificationClient{
		NotificationsAddr:    notificationsURL,
		httpClient:           &http.Client{},
		NotificationsChannel: notificationsChannel,
		readChannel:          make(chan *string),
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
	defer ws.CloseConnection(conn)
	if err != nil {
		err = errors.WrapListeningNotificationsError(ws.WrapDialingError(err, u.String()))
		return err
	}

	go func() {
		if err := ws.HandleRecv(conn, client.readChannel, nil); err != nil {
			log.Error(errors.WrapListeningNotificationsError(err))
			return
		}
	}()

	conn.SetPingHandler(func(string) error {
		//log.Warn("Received ping from notifications, sending pong")
		return conn.WriteMessage(websocket.PongMessage, nil)
	})

Loop:
	for {
		select {
		case msgString, ok := <-client.readChannel:
			if ok {
				msg, err := ws.ParseMessage(msgString)
				if err != nil {
					log.Error(errors.WrapListeningNotificationsError(err))
					break Loop
				}
				client.parseToNotification(msg)
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

	_, err = DoRequest(client.httpClient, req, nil)
	return errors.WrapStopListeningError(err)
}

func (client *NotificationClient) AddNotification(notificationMsg *notificationMessages.NotificationMessage,
	authToken string) error {
	req, err := BuildRequest("POST", client.NotificationsAddr, api.NotificationPath, notificationMsg)
	if err != nil {
		return errors.WrapAddNotificationError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	_, err = DoRequest(client.httpClient, req, nil)
	return errors.WrapAddNotificationError(err)
}

func (client *NotificationClient) GetOthersListening(authToken string) ([]string, error) {
	req, err := BuildRequest("GET", client.NotificationsAddr, api.GetListenersPath, nil)
	if err != nil {
		return nil, errors.WrapGetOthersListeningError(err)
	}

	req.Header.Set(tokens.AuthTokenHeaderName, authToken)

	var usernames []string
	_, err = DoRequest(client.httpClient, req, &usernames)
	if err != nil {
		return nil, errors.WrapGetOthersListeningError(err)
	}

	return usernames, nil
}

func (client *NotificationClient) parseToNotification(msg *ws.Message) {
	desMsg, err := notificationMessages.DeserializeNotificationMessage(msg)
	if err != nil {
		log.Error(err)
		return
	}

	notificationMsg := desMsg.(*notificationMessages.NotificationMessage)
	notificationMsg.Receive(ws.MakeTimestamp())
	notificationMsg.LogReceive(notificationMessages.Notification)
	client.NotificationsChannel <- &notificationMsg.Notification

	log.Infof("Received %s from the websocket", notificationMsg.Notification.Content)
}
