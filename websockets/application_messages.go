package websockets

const (
	Start    = "START"
	Reject   = "REJECT"
	SetToken = "SETTOKEN"
	Finish   = "FINISH"
	Error    = "ERROR"
)

type StartMessage struct{}

func (s StartMessage) ConvertToWSMessage() *WebsocketMsg {
	return NewStandardMsg(Start, nil)
}

func (s StartMessage) ConvertToWSMessageWithInfo(info TrackedInfo) *WebsocketMsg {
	return NewReplyMsg(Start, nil, info)
}

type RejectMessage struct{}

func (r RejectMessage) ConvertToWSMessage(info TrackedInfo) *WebsocketMsg {
	return NewReplyMsg(Reject, nil, info)
}

type FinishMessage struct {
	Success bool
}

func (f FinishMessage) ConvertToWSMessage() *WebsocketMsg {
	return NewStandardMsg(Finish, f)
}

type SetTokenMessage struct {
	TokenField   string
	TokensString []string
}

func (s SetTokenMessage) ConvertToWSMessage() *WebsocketMsg {
	return NewStandardMsg(SetToken, s)
}

type ErrorMessage struct {
	Info  string
	Fatal bool
}

func (e ErrorMessage) ConvertToWSMessageWithInfo(info TrackedInfo) *WebsocketMsg {
	return NewReplyMsg(Error, e, info)
}

func (e ErrorMessage) ConvertToWSMessage() *WebsocketMsg {
	return NewStandardMsg(Error, e)
}
