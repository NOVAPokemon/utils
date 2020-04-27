package websockets

import (
	"errors"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

const PongWait = 10 * time.Second
const PingPeriod = (PongWait * 9) / 10

func ParseMessage(msg *string) (*Message, error) {

	if msg == nil {
		return nil, errors.New("message is nil")
	}

	msgParts := strings.Split(*msg, " ")

	if len(msgParts) < 1 {
		return nil, errors.New("invalid msg format")
	}

	return &Message{
		MsgType: msgParts[0],
		MsgArgs: msgParts[1:],
	}, nil
}

func SendMessage(msg Message, channel chan *string) {
	toSend := msg.Serialize()
	channel <- &toSend
}

func HandleSend(conn *websocket.Conn, inChannel chan *string, endConnection chan struct{}) error {
	defer close(inChannel)
	defer log.Warn("Closing send routine")

	for {
		select {
		case msg := <-inChannel:
			err := conn.WriteMessage(websocket.TextMessage, []byte(*msg))
				if err != nil {
					closeConnectionThroughChannel(conn, endConnection)
					return err
				}
				log.Infof("Wrote %s into the channel", *msg)
		case <-endConnection:
			return nil
		}

	}
}

func HandleRecv(conn *websocket.Conn, outChannel chan *string, endConnection chan struct{}) error {
	defer close(outChannel)
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
				return err
			} else {
				msg := strings.TrimSpace(string(message))
				log.Debugf("Message received: %s", msg)
				outChannel <- &msg
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
