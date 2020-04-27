package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/battles"
	"github.com/NOVAPokemon/utils/websockets/trades"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

func Send(conn *websocket.Conn, msg *string) {
	err := conn.WriteMessage(websocket.TextMessage, []byte(*msg))

	if err != nil {
		return
	}
}

func ReadMessagesToChan(conn *websocket.Conn, msgChan chan *string, finished chan struct{}) {
	defer log.Info("Closed in channel")
	defer close(msgChan)

	for {
		select {
		case <-finished:
			return
		default:
			_, message, err := conn.ReadMessage()

			if err != nil {
				close(finished)
				return
			}

			msg := string(message)
			battleMsg, err := websockets.ParseMessage(&msg)
			if err != nil {
				close(finished)
				return
			}

			if battleMsg.MsgType == battles.Finish {
				log.Info("Received finish message")
				close(finished)
				return
			}
			msgChan <- &msg
		}
	}
}

func ReadMessagesWithoutParse(conn *websocket.Conn) (*websockets.Message, error) {
	_, msgBytes, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	msgString := string(msgBytes)
	log.Debugf("Received %s from the websocket", msgString)

	msg, err := websockets.ParseMessage(&msgString)
	if err != nil {
		return trades.NoneMessageConst, nil
	}

	return msg, nil
}

func WaitForStart(started, finish chan struct{}) {
	select {
	case <-started:
		return
	case <-finish:
		return
	}
}

func MainLoop(conn *websocket.Conn, writeChannel chan *string, finished chan struct{}) {
	defer log.Info("Closed out channel")
	defer close(writeChannel)
	for {
		select {
		case <-finished:
			return
		case msg := <-writeChannel:
			Send(conn, msg)
		}
	}
}

// REQUESTS

func BuildRequest(method, host, path string, body interface{}) (request *http.Request, err error) {
	hostUrl := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   path,
	}

	if body != nil {
		jsonStr, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		request, err = http.NewRequest(method, hostUrl.String(), bytes.NewBuffer(jsonStr))
	} else {
		request, err = http.NewRequest(method, hostUrl.String(), nil)
	}

	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")

	return request, nil
}

// For now this function assumes that a response should always have 200 code
func DoRequest(httpClient *http.Client, request *http.Request, responseBody interface{}) (*http.Response, error) {
	log.Infof("Doing request: %s %s", request.Method, request.URL.String())

	if httpClient == nil {
		return nil, errors.New(fmt.Sprintf("httpclient is nil for: %s", request.URL.String()))
	}

	resp, err := httpClient.Do(request)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Warnf("got status code %d", resp.StatusCode)
		return nil, errors.New(fmt.Sprintf("got status code %d", resp.StatusCode))
	}

	if responseBody != nil {
		err = json.NewDecoder(resp.Body).Decode(responseBody)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}

	return resp, nil
}
