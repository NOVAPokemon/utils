package clients

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/websockets"
	"github.com/NOVAPokemon/utils/websockets/trades"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/cookiejar"
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
		log.Info(body)
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
func DoRequest(httpClient *http.Client, request *http.Request, responseBody interface{}) error {
	if httpClient == nil {
		return errors.New(fmt.Sprintf("httpclient is nil for: %s", request.URL.String()))
	}

	if httpClient.Jar == nil {
		return errors.New(fmt.Sprintf("jar is nil for: %s", request.URL.String()))
	}

	resp, err := httpClient.Do(request)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("got status code %d", resp.StatusCode))
	}

	if responseBody != nil {
		err = json.NewDecoder(resp.Body).Decode(responseBody)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateJarWithCookies(cookies ...*http.Cookie) (*cookiejar.Jar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	jar.SetCookies(&url.URL{
		Scheme: "http",
		Host:   utils.Host,
		Path:   "/",
	}, cookies)

	return jar, nil
}
