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
		msgType: msgParts[0],
		msgArgs: msgParts[1:],
	}
}

func SendMessage(msg *TradeMessage, channel chan *string) {
	builder := strings.Builder{}

	builder.WriteString(msg.msgType)
	builder.WriteString(" ")

	for _, arg := range msg.msgArgs {
		builder.WriteString(arg)
		builder.WriteString(" ")
	}

	toSend := builder.String()
	select {
	case channel <- &toSend:
	default:
		close(channel)
	}
}
