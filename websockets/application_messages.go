package websockets

const (
	SetToken = "SETTOKEN"
	Finish   = "FINISH"
	Error    = "ERROR"
)

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

func (e ErrorMessage) ConvertToWSMessage() *WebsocketMsg {
	return NewStandardMsg(Error, e)
}
