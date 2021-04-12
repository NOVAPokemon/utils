package clients

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/NOVAPokemon/utils/websockets"
	ws "github.com/NOVAPokemon/utils/websockets"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	RequestTimeout = 5 * time.Second
)

type BasicClient struct {
	usingIngress bool
	ingress      string
}

func NewBasicClient(usingIngress bool, ingressURL string) *BasicClient {
	return &BasicClient{
		usingIngress: usingIngress,
		ingress:      ingressURL,
	}
}

func (b *BasicClient) BuildRequest(method, serverAddr, path string, body interface{}) (*http.Request, error) {
	var addr string
	if b.usingIngress {
		addr = b.ingress
	} else {
		addr = serverAddr
	}

	hostUrl := url.URL{
		Scheme: "http",
		Host:   addr,
		Path:   path,
	}

	var (
		err        error
		request    *http.Request
		bodyBuffer *bytes.Buffer
	)

	if body != nil {
		var jsonStr []byte
		jsonStr, err = json.Marshal(body)
		if err != nil {
			return nil, wrapBuildRequestError(err)
		}
		bodyBuffer = bytes.NewBuffer(jsonStr)
	} else {
		bodyBuffer = new(bytes.Buffer)
	}

	request, err = http.NewRequest(method, hostUrl.String(), bodyBuffer)
	if err != nil {
		return nil, wrapBuildRequestError(err)
	}

	request.Header.Set("Content-Type", "application/json")

	return request, nil
}

func (b *BasicClient) BuildRequestForHost(method, serverAddr, host, path string, body interface{}) (*http.Request,
	error) {
	var addr string
	if b.usingIngress {
		addr = b.ingress
	} else {
		addr = serverAddr
	}

	hostUrl := url.URL{
		Scheme: "http",
		Host:   addr,
		Path:   path,
	}

	var (
		err        error
		request    *http.Request
		bodyBuffer *bytes.Buffer
	)

	if body != nil {
		var jsonStr []byte
		jsonStr, err = json.Marshal(body)
		if err != nil {
			return nil, wrapBuildRequestError(err)
		}
		bodyBuffer = bytes.NewBuffer(jsonStr)
	} else {
		bodyBuffer = new(bytes.Buffer)
	}

	request, err = http.NewRequest(method, hostUrl.String(), bodyBuffer)
	if err != nil {
		return nil, wrapBuildRequestError(err)
	}

	request.Host = host
	request.Header.Set("Content-Type", "application/json")

	return request, nil
}

func Send(conn *websocket.Conn, msg *ws.WebsocketMsg, writer ws.CommunicationManager) error {
	return ws.WrapWritingMessageError(writer.WriteGenericMessageToConn(conn, msg))
}

func ReadMessagesFromConnToChan(conn *websocket.Conn, msgChan chan *ws.WebsocketMsg, finished chan struct{},
	commsManager ws.CommunicationManager) {
	messagesQueue := make(chan (<-chan *websockets.WebsocketMsg), 10)

	cancel := make(chan struct{})

	go func() {
		defer func() {
			close(msgChan)
		}()
		for {
			select {
			case <-finished:
				log.Info("finished message queue routine")
				return
			case <-cancel:
				log.Info("canceled message queue routine")
				return
			case chanToWait := <-messagesQueue:
				log.Info("waiting for message")
				msg := <-chanToWait
				log.Info("got message")

				select {
				case <-finished:
					log.Info("finished message queue routine while waiting")
				case msgChan <- msg:
				}
				log.Info("wrote message")
			}
		}
	}()

	defer func() {
		close(cancel)
		log.Info("closing read routine")
	}()

	for {
		msgChan, err := commsManager.ReadMessageFromConn(conn)
		if err != nil {
			log.Warn(err)
			return
		}

		select {
		case <-finished:
			return
		default:
			if msgChan != nil {
				messagesQueue <- msgChan
			}
		}
	}
}

func WriteTextMessagesFromChanToConn(conn *websocket.Conn, commsManager ws.CommunicationManager,
	writeChannel <-chan *ws.WebsocketMsg, finished chan struct{}) {
	defer log.Info("closing write routine")

	for {
		select {
		case <-finished:
			return
		case msg, ok := <-writeChannel:
			if !ok {
				return
			}

			err := commsManager.WriteGenericMessageToConn(conn, msg)
			if err != nil {
				log.Warn(err)
				return
			}
		}
	}
}

func ReadMessagesFromConnToChanWithoutClosing(conn *websocket.Conn, msgChan chan *ws.WebsocketMsg,
	finished chan struct{}, manager ws.CommunicationManager) {
	messagesQueue := make(chan (<-chan *websockets.WebsocketMsg), 10)

	cancel := make(chan struct{})

	go func() {
		for {
			select {
			case <-cancel:
				return
			case chanToWait := <-messagesQueue:
				msgChan <- (<-chanToWait)
			case <-finished:
				return
			}
		}
	}()

	for {
		select {
		case <-finished:
			return
		default:
			msgChan, err := manager.ReadMessageFromConn(conn)
			if err != nil {
				log.Warn(err)
				close(cancel)
				return
			}

			if msgChan != nil {
				messagesQueue <- msgChan
			}
		}
	}
}

func SetDefaultPingHandler(conn *websocket.Conn, writeChannel chan *ws.WebsocketMsg) {
	_ = conn.SetReadDeadline(time.Now().Add(timeoutInDuration))
	conn.SetPingHandler(func(string) error {
		writeChannel <- ws.NewControlMsg(websocket.PongMessage)
		return conn.SetReadDeadline(time.Now().Add(timeoutInDuration))
	})
}

// REQUESTS

func Read(conn *websocket.Conn, manager ws.CommunicationManager) (*ws.WebsocketMsg, error) {
	msgChan, err := manager.ReadMessageFromConn(conn)
	if err != nil {
		return nil, ws.WrapReadingMessageError(err)
	}

	msg := <-msgChan

	return msg, nil
}

// For now this function assumes that a response should always have 200 code
func DoRequest(httpClient *http.Client, request *http.Request, responseBody interface{},
	manager ws.CommunicationManager) (*http.Response, error) {
	log.Infof("Doing request: %s %s %s", request.Method, request.URL.String(),
		request.Header.Get("Host"))

	if httpClient == nil {
		return nil, wrapDoRequestError(newHttpClientNilError(request.URL.String()))
	}

	resp, err := manager.DoHTTPRequest(httpClient, request)
	if err != nil {
		log.Error(err)
		return nil, wrapDoRequestError(err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, wrapDoRequestError(newUnexpectedResponseStatusError(resp.StatusCode))
	}

	if responseBody != nil {
		err = json.NewDecoder(resp.Body).Decode(responseBody)
		if err != nil {
			return nil, wrapDoRequestError(err)
		}
	}

	return resp, nil
}
