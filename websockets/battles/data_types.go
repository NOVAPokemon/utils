package battles

// Message Types
const (
	ATTACK   = "ATTACK"
	USE_ITEM = "USE_ITEM"
	DEFEND   = "DEFEND"

	// Update Types
	UPDATE_PLAYER_POKEMON    = "UPDATE_PLAYER_POKEMON"
	UPDATE_ADVERSARY_POKEMON = "UPDATE_ADVERSARY_POKEMON"

	UPDATE_PLAYER    = "UPDATE_PLAYER"
	UPDATE_ADVERSARY = "UPDATE_ADVERSARY"

	// Setup Message Types
	SELECT_POKEMON = "SELECT_POKEMON"

	STATUS = "STATUS"

	// Error
	ERROR = "ERROR"
)

type BattleMessage struct {
	MsgType string
	MsgArgs []string
}
