package api

import "fmt"

const (
	// GetBattlesPath path to get all battles
	GetBattlesPath = "/battles"

	// ChallengeToBattlePath path to challenge user for a battle
	ChallengeToBattlePath = "/battles/challenge/%s"

	// AcceptChallengePath path to accept challenge
	AcceptChallengePath = "/battles/accept/%s"

	// QueueForBattlePath path to queue for battle
	QueueForBattlePath = "/battles/queue"

	// RejectChallengePath path to reject battle challenge
	RejectChallengePath = "/battles/reject/%s"
)

const (
	// BattleIdPathVar battleId variable in path
	BattleIdPathVar = "battleId"

	// TargetPlayerIdPathvar targetPlayer variable in path
	TargetPlayerIdPathvar = "targetPlayer"
)

var (
	ChallengeToBattleRoute = fmt.Sprintf(ChallengeToBattlePath, fmt.Sprintf("{%s}", TargetPlayerIdPathvar))
	AcceptChallengeRoute   = fmt.Sprintf(AcceptChallengePath, fmt.Sprintf("{%s}", BattleIdPathVar))
	RejectChallengeRoute   = fmt.Sprintf(RejectChallengePath, fmt.Sprintf("{%s}", BattleIdPathVar))
)
