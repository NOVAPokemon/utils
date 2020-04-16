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
		return nil, errors.New("message is null")
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
	//logrus.Infof("Sending: %s", toSend)
	channel <- &toSend
}

func handleSend(conn *websocket.Conn, inChannel chan *string, endConnection chan struct{}, finished *bool) {
	defer close(inChannel)
	defer log.Warn("Closing send routine")

	for {
		select {
		case msg := <-inChannel:
			err := conn.WriteMessage(websocket.TextMessage, []byte(*msg))
			if err != nil {

				if err != nil {
					if *finished {
						log.Info("lobby write routine finished properly")
					} else {
						log.Warn(err)
						log.Warn("closed lobby because could not read")
					}
					closeConnectionThroughChannel(conn, endConnection)
					return
				}
				log.Infof("Wrote %s into the channel", *msg)
			}
		case <-endConnection:
			return
		}

	}
}

func handleRecv(conn *websocket.Conn, outChannel chan *string, endConnection chan struct{}, finished *bool) {
	defer close(outChannel)
	for {
		select {
		case <-endConnection:
			return
		default:
			_, message, err := conn.ReadMessage()

			if err != nil {
				if *finished {
					log.Info("lobby read routine finished properly")
				} else {
					log.Warn(err)
					log.Warn("closed lobby because could not read")
				}
				closeConnectionThroughChannel(conn, endConnection)
				return
			} else {
				msg := strings.TrimSpace(string(message))
				log.Infof("Message received: %s", msg)
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
	select {
	case <-channel:
		return
	default:
		close(channel)
	}
}
