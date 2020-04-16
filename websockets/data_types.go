package websockets

import "strings"

type Message struct {
	MsgType string
	MsgArgs []string
}

func (msg Message) Serialize() string {
	builder := strings.Builder{}

	builder.WriteString(msg.MsgType)
	builder.WriteString(" ")

	for _, arg := range msg.MsgArgs {
		builder.WriteString(arg)
		builder.WriteString(" ")
	}

	return builder.String()
}

type GenericMsg struct {
	MsgType int
	Data []byte
}