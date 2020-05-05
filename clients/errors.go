package clients

import "github.com/pkg/errors"

// AUTH CLIENT
// Auth client error messages
const (
	errorLogin    = "error logging in"
	errorRegister = "error registering"
)

// Auth client wrappers
func wrapLoginError(err error) error {
	return errors.Wrap(err, errorLogin)
}

func wrapRegisterError(err error) error {
	return errors.Wrap(err, errorRegister)
}

// BATTLES CLIENT
// Battles client error messages
const (
	errorGetBattleLobbies      = "error getting battle lobbies"
	errorQueueForBattle        = "error queueing for battle"
	errorChallengeForBattle    = "error challenging for battle"
	errorAcceptBattleChallenge = "error accepting battle challenge"
)

// Battles client wrappers
func wrapGetBattleLobbiesError(err error) error {
	return errors.Wrap(err, errorGetBattleLobbies)
}

func wrapQueueForBattleError(err error) error {
	return errors.Wrap(err, errorQueueForBattle)
}

func wrapChallengeForBattleError(err error) error {
	return errors.Wrap(err, errorChallengeForBattle)
}

func wrapAcceptBattleChallengeError(err error) error {
	return errors.Wrap(err, errorAcceptBattleChallenge)
}

// GYM CLIENT
// Gym client error messages
const (
	errorGetGymInfo = "error getting gym info"
	errorCreateGym  = "error creating gym"
	errorCreateRaid = "error creating raid"
	errorEnterRaid  = "error entering raid"
)

// Gym client wrappers
func wrapGetGymInfoError(err error) error {
	return errors.Wrap(err, errorGetGymInfo)
}

func wrapCreateGymError(err error) error {
	return errors.Wrap(err, errorCreateGym)
}

func wrapCreateRaidError(err error) error {
	return errors.Wrap(err, errorCreateRaid)
}

func wrapEnterRaidError(err error) error {
	return errors.Wrap(err, errorEnterRaid)
}

// LOCATION CLIENT
// Location client error messages
const (
	errorStartLocationUpdate = "error starting location updates"
	errorAddGymLocation      = "error adding gym location"
	errorCatchWildPokemon    = "error catching wild pokemon"
	errorConnect             = "error connecting"
	errorUpdateLocation      = "error updating location"
)

var (
	errorNoPokemonsVinicity = errors.New("no pokemons in vicinity")
	errorNoPokeballs        = errors.New("no pokeballs")
)

// Location client wrappers
func wrapStartLocationUpdatesError(err error) error {
	return errors.Wrap(err, errorStartLocationUpdate)
}

func wrapAddGymLocationError(err error) error {
	return errors.Wrap(err, errorAddGymLocation)
}

func wrapCatchWildPokemonError(err error) error {
	return errors.Wrap(err, errorCatchWildPokemon)
}

func wrapConnectError(err error) error {
	return errors.Wrap(err, errorConnect)
}

func wrapUpdateLocation(err error) error {
	return errors.Wrap(err, errorUpdateLocation)
}

// MICROTRANSACTIONS CLIENT
// Microtransactions client error messages
const (
	errorGetOffers             = "error getting offers"
	errorGetTransactionRecords = "error getting transaction records"
	errorPerformTransaction    = "error performing transaction"
)

// Microtransactions client wrappers
func wrapGetOffersError(err error) error {
	return errors.Wrap(err, errorGetOffers)
}

func wrapGetTransactionsRecordsError(err error) error {
	return errors.Wrap(err, errorGetTransactionRecords)
}

func wrapPerformTransactionError(err error) error {
	return errors.Wrap(err, errorPerformTransaction)
}

// NOTIFICATIONS CLIENT
// Notifications client error messages
const (
	errorListeningNotifications = "error listening to notifications"
	errorStopListening          = "error stopping listen on notifications"
	errorAddNotification        = "error adding notification"
	errorGetOthersListening     = "error getting others listening"
)

// Notifications client wrappers
func wrapListeningNotificationsError(err error) error {
	return errors.Wrap(err, errorListeningNotifications)
}

func wrapStopListeningError(err error) error {
	return errors.Wrap(err, errorStopListening)
}

func wrapAddNotificationError(err error) error {
	return errors.Wrap(err, errorAddNotification)
}

func wrapGetOthersListeningError(err error) error {
	return errors.Wrap(err, errorGetOthersListening)
}

// STORE CLIENT
// Store client error messages
const (
	errorGetItems = "error getting items"
	errorBuyItem  = "error buying item"
)

// Store client wrappers
func wrapGetItemsError(err error) error {
	return errors.Wrap(err, errorGetItems)
}

func wrapBuyItemError(err error) error {
	return errors.Wrap(err, errorBuyItem)
}

// TRADES CLIENT
// Trades client error messages
const (
	errorGetTradeLobbies          = "error getting trade lobbies"
	errorCreateTradeLobby         = "error creating trade lobby"
	errorJoinTradeLobby           = "error joining trade lobby"
	errorHandleMessagesTradeLobby = "error handling received messages"
)

// Trades client wrappers
func wrapGetTradeLobbiesError(err error) error {
	return errors.Wrap(err, errorGetTradeLobbies)
}

func wrapCreateTradeLobbyError(err error) error {
	return errors.Wrap(err, errorCreateTradeLobby)
}

func wrapJoinTradeLobbyError(err error) error {
	return errors.Wrap(err, errorJoinTradeLobby)
}

func wrapHandleMessagesTradeError(err error) error {
	return errors.Wrap(err, errorHandleMessagesTradeLobby)
}
