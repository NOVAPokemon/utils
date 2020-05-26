package api

import "fmt"

const (
	GetBattlesPath        = "/battles"
	ChallengeToBattlePath = "/battles/challenge/%s"
	AcceptChallengePath   = "/battles/accept/%s"
	QueueForBattlePath    = "/battles/queue"
	RejectChallengePath   = "/battles/reject/%s"
)

const (
	BattleIdPathVar       = "battleId"
	TargetPlayerIdPathvar = "targetPlayer"
)

var (
	ChallengeToBattleRoute = fmt.Sprintf(ChallengeToBattlePath, fmt.Sprintf("{%s}", TargetPlayerIdPathvar))
	AcceptChallengeRoute   = fmt.Sprintf(AcceptChallengePath, fmt.Sprintf("{%s}", BattleIdPathVar))
	RejectChallengeRoute   = fmt.Sprintf(RejectChallengePath, fmt.Sprintf("{%s}", BattleIdPathVar))
)
