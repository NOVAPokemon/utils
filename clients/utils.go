package clients

import (
	"bytes"
	"encoding/json"
	ws "github.com/NOVAPokemon/utils/websockets"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"time"
)

func Send(conn *websocket.Conn, msg *string) error {
	return ws.WrapWritingMessageError(conn.WriteMessage(websocket.TextMessage, []byte(*msg)))
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
			battleMsg, err := ws.ParseMessage(&msg)
			if err != nil {
				close(finished)
				return
			}

			if battleMsg.MsgType == ws.Finish {
				log.Info("Received finish message")
				close(finished)
				return
			}
			msgChan <- &msg
		}
	}
}

func Read(conn *websocket.Conn) (*ws.Message, error) {
	_, msgBytes, err := conn.ReadMessage()
	if err != nil {
		return nil, ws.WrapReadingMessageError(err)
	}

	msgString := string(msgBytes)
	log.Debugf("Received %s from the websocket", msgString)

	msg, err := ws.ParseMessage(&msgString)
	if err != nil {
		return nil, ws.WrapReadingMessageError(err)
	}

	return msg, nil
}

func WaitForStart(started, rejected, finish chan struct{}, requestTimestamp int64) int64 {
	var responseTimestamp int64

	select {
	case <-started:
		responseTimestamp = ws.MakeTimestamp()
	case <-rejected:
		responseTimestamp = ws.MakeTimestamp()
	case <-finish:
		return 0
	}

	return responseTimestamp - requestTimestamp
}

func MainLoop(conn *websocket.Conn, writeChannel chan ws.GenericMsg, finished chan struct{}) {
	defer log.Info("Closed out channel")
	defer close(writeChannel)

	_ = conn.SetReadDeadline(time.Now().Add(timeoutInDuration))
	conn.SetPingHandler(func(string) error {
		// log.Warn("Received ping, ponging...")
		writeChannel <- ws.GenericMsg{MsgType: websocket.PongMessage, Data: nil}
		return conn.SetReadDeadline(time.Now().Add(timeoutInDuration))
	})

	for {
		select {
		case <-finished:
			return
		case msg := <-writeChannel:
			if err := conn.WriteMessage(msg.MsgType, msg.Data); err != nil {
				log.Error(wrapMainLoopError(err))
				return
			}
		}
	}
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
