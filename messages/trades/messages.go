package trades

var (
	ErrorOneItemAtATime = ErrorMessage{
		Info:  "can only add one item to trade at a time",
		Fatal: false,
	}.Serialize()

	ErrorParsing = ErrorMessage{
		Info:  "error parsing message",
		Fatal: false,
	}.Serialize()
)
