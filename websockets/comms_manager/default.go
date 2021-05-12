package comms_manager

import (
	"math/rand"
	"time"

	http "github.com/bruno-anjos/archimedesHTTPClient"
	log "github.com/sirupsen/logrus"

	"github.com/NOVAPokemon/utils/websockets"
	"github.com/gorilla/websocket"
)

const (
	minTimeBetweenRetries = 5
	maxTimeBetweenRetries = 10
)

type DefaultCommsManager struct {
	websockets.CommsManagerWithCounter
}

func (d *DefaultCommsManager) ApplyReceiveLogic(msg *websockets.WebsocketMsg) *websockets.WebsocketMsg {
	if msg.MsgType == websocket.TextMessage && msg.Content.MsgKind == websockets.Reply {
		msg.Content.RequestTrack.Receive(websockets.MakeTimestamp())
		msg.Content.RequestTrack.LogReceive(msg.Content.AppMsgType)
	}
	return msg
}

func (d *DefaultCommsManager) ApplySendLogic(msg *websockets.WebsocketMsg) *websockets.WebsocketMsg {
	if msg.MsgType == websocket.TextMessage && msg.Content.MsgKind == websockets.Request {
		msg.Content.RequestTrack.Emit(websockets.MakeTimestamp())
		msg.Content.RequestTrack.LogEmit(msg.Content.AppMsgType)
	}
	return msg
}

func (d *DefaultCommsManager) WriteGenericMessageToConn(conn *websocket.Conn, msg *websockets.WebsocketMsg) error {
	msg = d.ApplySendLogic(msg)
	if msg.Content == nil {
		return conn.WriteMessage(msg.MsgType, nil)
	}
	return conn.WriteMessage(msg.MsgType, msg.Content.Serialize())
}

func (d *DefaultCommsManager) ReadMessageFromConn(conn *websocket.Conn) (<-chan *websockets.WebsocketMsg, error) {
	msgType, p, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	content := websockets.ParseContent(p)
	msg := &websockets.WebsocketMsg{
		MsgType: msgType,
		Content: content,
	}

	msgChan := make(chan *websockets.WebsocketMsg)
	go func() {
		msg = d.ApplyReceiveLogic(msg)
		msgChan <- msg
	}()

	return msgChan, nil
}

func (d *DefaultCommsManager) DoHTTPRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	log.Debugf("doing req: %+v", req)
	log.Debug("host in dohttprequest: ", req.Host)

	var (
		resp    *http.Response
		err     error
		timer   time.Timer
		retried = false
	)

	for {
		resp, err = client.Do(req)
		ts := websockets.MakeTimestamp()
		if d.CommsManagerWithCounter.LogRequestAndRetry(resp, err, ts) {
			break
		}

		retried = true
		waitingTime := time.Duration(rand.Int31n(maxTimeBetweenRetries-minTimeBetweenRetries+1)+minTimeBetweenRetries) * time.Second
		timer.Reset(waitingTime)
		<-timer.C
	}

	if retried {
		if !timer.Stop() {
			<-timer.C
		}
	}

	return resp, err
}

func (d *DefaultCommsManager) HTTPRequestInterceptor(next http.Handler) http.Handler {
	return next
}
