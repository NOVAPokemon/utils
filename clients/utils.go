package clients

import (
	"bufio"
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
	"os"
)

func Send(conn *websocket.Conn, msg *string) {
	err := conn.WriteMessage(websocket.TextMessage, []byte(*msg))

	if err != nil {
		return
	} else {
		log.Debugf("Wrote %s into the channel", *msg)
	}
}

func ReadMessages(conn *websocket.Conn, finished chan struct{}) {
	defer close(finished)

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Error(err)
			return
		}

		msg := string(message)
		log.Debugf("Received %s from the websocket", msg)

		err, tradeMsg := websockets.ParseMessage(&msg)
		if err != nil {
			log.Error(err)
			continue
		}

		log.Infof("Message: %s", msg)

		if tradeMsg.MsgType == trades.FINISH {
			log.Info("Finished trade.")
			return
		}

		if tradeMsg.MsgType == battles.FINISH {
			log.Info("Finished trade.")
			_ = conn.Close()
			return
		}
	}
}

func WriteMessage(writeChannel chan *string) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter text: ")
		text, _ := reader.ReadString('\n')
		writeChannel <- &text
	}
}

func Finish() {
	log.Info("Finishing connection...")
}

func MainLoop(conn *websocket.Conn, writeChannel chan *string, finished chan struct{}) {
	for {
		select {
		case <-finished:
			Finish()
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

	log.Info("requesting: ", hostUrl.String())

	request.Header.Set("Content-Type", "application/json")

	return request, nil
}

// For now this function assumes that a response should always have 200 code
func DoRequest(httpClient *http.Client, request *http.Request, responseBody interface{}) (*http.Response, error) {
	log.Infof("Requesting: %s", request.URL.String())

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
