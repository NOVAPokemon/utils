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

func Send(conn *websocket.Conn, msg *string) error {
	return ws.WrapWritingMessageError(conn.WriteMessage(websocket.TextMessage, []byte(*msg)))
}

func ReadMessagesFromConnToChan(conn *websocket.Conn, msgChan chan string, finished chan struct{}) {
	defer func() {
		log.Info("closing read routine")
		close(msgChan)
	}()
	for {
		select {
		case <-finished:
			return
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Warn(err)
				return
			}
			msgChan <- string(message)
		}
	}
}

func WriteMessagesFromChanToConn(conn *websocket.Conn, writeChannel <-chan ws.GenericMsg, finished chan struct{}) {
	defer log.Info("closing write routine")

	for {
		select {
		case <-finished:
			return
		case msg := <-writeChannel:
			if err := conn.WriteMessage(msg.MsgType, msg.Data); err != nil {
				log.Warn(err)
				return
			}
		}
	}
}

func ReadMessagesFromConnToChanWithoutClosing(conn *websocket.Conn, msgChan chan string, finished chan struct{}) {
	for {
		select {
		case <-finished:
			return
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Warn(err)
				return
			}
			msgChan <- string(message)
		}
	}
}

func SetDefaultPingHandler(conn *websocket.Conn, writeChannel chan ws.GenericMsg) {
	_ = conn.SetReadDeadline(time.Now().Add(timeoutInDuration))
	conn.SetPingHandler(func(string) error {
		// log.Warn("Received ping, ponging...")
		writeChannel <- ws.GenericMsg{MsgType: websocket.PongMessage, Data: nil}
		return conn.SetReadDeadline(time.Now().Add(timeoutInDuration))
	})
}

func Read(conn *websocket.Conn) (*ws.Message, error) {
	_, msgBytes, err := conn.ReadMessage()
	if err != nil {
		return nil, ws.WrapReadingMessageError(err)
	}

	msgStr := string(msgBytes)
	log.Debugf("Received %s from the websocket", msgStr)
	msgParsed, err := ws.ParseMessage(msgStr)
	if err != nil {
		return nil, ws.WrapReadingMessageError(err)
	}
	return msgParsed, nil
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
func DoRequest(httpClient *http.Client, request *http.Request, responseBody interface{}) (*http.Response, error) {
	log.Infof("Doing request: %s %s", request.Method, request.URL.String())

	if httpClient == nil {
		return nil, wrapDoRequestError(newHttpClientNilError(request.URL.String()))
	}

	resp, err := httpClient.Do(request)
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
