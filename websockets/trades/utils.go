package trades

import (
	"errors"
	"strings"
)

func ParseMessage(msg *string) (error, *TradeMessage) {
	msgParts := strings.Split(*msg, " ")

	if len(msgParts) < 1 {
		return errors.New("invalid msg format"), nil
	}

	return nil, &TradeMessage{
		MsgType: msgParts[0],
		MsgArgs: msgParts[1:],
	}
}

func SendMessage(msg *TradeMessage, channel chan *string) {
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
