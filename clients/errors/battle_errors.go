package errors

import "github.com/pkg/errors"

const (
	errorGetBattleLobbies      = "error getting battle lobbies"
	errorQueueForBattle        = "error queueing for battle"
	errorChallengeForBattle    = "error challenging for battle"
	errorAcceptBattleChallenge = "error accepting battle challenge"
)

func WrapGetBattleLobbiesError(err error) error {
	return errors.Wrap(err, errorGetBattleLobbies)
}

func WrapQueueForBattleError(err error) error {
	return errors.Wrap(err, errorQueueForBattle)
}

func WrapChallengeForBattleError(err error) error {
	return errors.Wrap(err, errorChallengeForBattle)
}

func WrapAcceptBattleChallengeError(err error) error {
	return errors.Wrap(err, errorAcceptBattleChallenge)
}