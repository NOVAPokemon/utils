package battles

// Battle message types
// Message Types
const PokemonsPerBattle = 3

var (
	ErrInvalidMessageFormat   = "ErrInvalidMessageFormat"
	ErrInvalidMessageType     = "ErrInvalidMessageType"
	ErrPokemonSelectionPhase  = "ErrPokemonSelectionPhase"
	ErrNoPokemonSelected      = "ErrNoPokemonSelected"
	ErrInvalidPokemonSelected = "ErrInvalidPokemonSelected"
	ErrPokemonNoHP            = "ErrPokemonNoHP"
	ErrCooldown               = "ErrCooldown"

	StatusDefended            = "StatusDefended"
	StatusOpponentedDeffended = "StatusOpponentedDeffended"
)

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
	ERROR  = "ERROR"
	FINISH = "FINISH_BATTLE"
	START  = "START"

	SET_TOKEN = "SET_TOKEN"
)
