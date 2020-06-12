package websockets

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	PongWait   = 2 * time.Second
	PingPeriod = (PongWait * 9) / 10
)

func ParseMessage(msg *string) (*Message, error) {
	if msg == nil {
		return nil, wrapMsgParsingError(ErrorMessageNil)
	}
	toReturn := &Message{}
	if err := json.Unmarshal([]byte(*msg), toReturn); err != nil {
		return nil, wrapMsgParsingError(ErrorMessageNil)
	}
	return toReturn, nil
}

func HandleSend(conn *websocket.Conn, outChannel chan GenericMsg, endConnection chan struct{}) error {
	defer log.Warn("Closing send routine")

	pingTicker := time.NewTicker(PingPeriod)
	conn.SetPongHandler(func(_ string) error {
		// log.Info("Received pong")
		return conn.SetReadDeadline(time.Now().Add(PongWait))
	})

	for {
		select {
		case <-pingTicker.C:
			// log.Warn("Pinging")
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				closeConnectionThroughChannel(conn, endConnection)
				return wrapHandleSendError(err)
			}
		case msg := <-outChannel:
			err := conn.WriteMessage(msg.MsgType, msg.Data)
			if err != nil {
				closeConnectionThroughChannel(conn, endConnection)
				return wrapHandleSendError(err)
			}
			log.Infof("Wrote %s into the channel", msg.Data)
		case <-endConnection:
			return nil
		}

	}
}

func HandleRecv(conn *websocket.Conn, inChannel chan *string, endConnection chan struct{}) error {
	defer close(inChannel)
	defer log.Warn("Closing receiving routine")

	for {
		select {
		case <-endConnection:
			return nil
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Warn("Closing finish channel and connection")
				closeConnectionThroughChannel(conn, endConnection)
				return wrapHandleReceiveError(err)
			} else {
				msg := strings.TrimSpace(string(message))
				log.Debugf("Message received: %s", msg)
				inChannel <- &msg
			}
		}
	}
}

func closeConnectionThroughChannel(conn *websocket.Conn, endConnection chan struct{}) {
	endChannel(endConnection)
	CloseConnection(conn)
}

func CloseConnection(conn *websocket.Conn) {
	if conn == nil {
		return
	}
	if err := conn.Close(); err != nil {
		log.Error(err)
	}
}

func endChannel(channel chan struct{}) {
	if channel == nil {
		return
	}

	select {
	case <-channel:
		return
	default:
		close(channel)
	}
}
