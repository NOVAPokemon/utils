package clients

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	ws "github.com/NOVAPokemon/utils/websockets"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	RequestTimeout = 5 * time.Second
)

func Send(conn *websocket.Conn, msg *ws.WebsocketMsg, writer ws.CommunicationManager) error {
	return ws.WrapWritingMessageError(writer.WriteGenericMessageToConn(conn, msg))
}

func ReadMessagesFromConnToChan(conn *websocket.Conn, msgChan chan *ws.WebsocketMsg, finished chan struct{},
	commsManager ws.CommunicationManager) {
	defer func() {
		log.Info("closing read routine")
		close(msgChan)
	}()
	for {
		select {
		case <-finished:
			return
		default:
			wsMsg, err := commsManager.ReadMessageFromConn(conn)
			if err != nil {
				log.Warn(err)
				return
			}
			if wsMsg != nil {
				msgChan <- wsMsg
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
	for {
		select {
		case <-finished:
			return
		default:
			wsMsg, err := manager.ReadMessageFromConn(conn)
			if err != nil {
				log.Warn(err)
				return
			}
			if wsMsg != nil {
				msgChan <- wsMsg
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

func Read(conn *websocket.Conn, manager ws.CommunicationManager) (*ws.WebsocketMsg, error) {
	wsMsg, err := manager.ReadMessageFromConn(conn)
	if err != nil {
		return nil, ws.WrapReadingMessageError(err)
	}

	return wsMsg, nil
}

// REQUESTS

func BuildRequest(method, host, path string, body interface{}) (*http.Request, error) {
	hostUrl := url.URL{
		Scheme: "http",
		Host:   host,
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

// For now this function assumes that a response should always have 200 code
func DoRequest(httpClient *http.Client, request *http.Request, responseBody interface{},
	manager ws.CommunicationManager) (*http.Response, error) {
	log.Infof("Doing request: %s %s", request.Method, request.URL.String())

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
