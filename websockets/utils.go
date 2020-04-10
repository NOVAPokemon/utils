package websockets

import (
	"errors"
	"strings"
	"time"
)

const PongWait = 10 * time.Second
const PingPeriod = (PongWait * 9) / 10

func ParseMessage(msg *string) (*Message, error) {
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
