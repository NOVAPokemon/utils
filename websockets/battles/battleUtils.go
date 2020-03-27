package battles

import (
	"errors"
	"strings"
)

func ParseMessage(msg *string) (error, *BattleMessage) {
	msgParts := strings.Split(*msg, " ")

	if len(msgParts) < 1 {
		return errors.New("invalid msg format"), nil
	}

	return nil, &BattleMessage{
		MsgType: msgParts[0],
		MsgArgs: msgParts[1:],
	}
}

func SendMessage(msg *BattleMessage, channel chan *string) {
	builder := strings.Builder{}

	builder.WriteString(msg.MsgType)
	builder.WriteString(" ")

	for _, arg := range msg.MsgArgs {
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
