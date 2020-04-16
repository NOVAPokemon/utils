package trades

var (
	ErrorOneItemAtATime = ErrorMessage{
		Info:  "can only add one item to trade at a time",
		Fatal: false,
	}.SerializeToWSMessage()

	ErrorParsing = ErrorMessage{
		Info:  "error parsing message",
		Fatal: false,
	}.SerializeToWSMessage()

	NoneMessageConst = NoneMessage{}.SerializeToWSMessage()
)
