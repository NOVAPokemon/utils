package websockets

import (
	"errors"
	"strings"
)

func ParseMessage(msg *string) (error, *Message) {
	msgParts := strings.Split(*msg, " ")

	if len(msgParts) < 1 {
		return errors.New("invalid msg format"), nil
	}

	return nil, &Message{
		MsgType: msgParts[0],
		MsgArgs: msgParts[1:],
	}
}

func SendMessage(msg Message, channel chan *string) {
	builder := strings.Builder{}

	builder.WriteString(msg.MsgType)
	builder.WriteString(" ")

	for _, arg := range msg.MsgArgs {
		builder.WriteString(arg)
		builder.WriteString(" ")
	}

	toSend := builder.String()
	channel <- &toSend
}
