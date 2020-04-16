package trades

var (
	ErrorParsing = ErrorMessage{
		Info:  "error parsing message",
		Fatal: false,
	}.SerializeToWSMessage()

	NoneMessageConst = NoneMessage{}.SerializeToWSMessage()
)
