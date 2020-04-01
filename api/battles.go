package api

import "fmt"

const GetBattlesPath = "/battles"
const ChallengeToBattlePath = "/battles/challenge/%s"
const AcceptChallengePath = "/battles/accept/%s"
const QueueForBattlePath = "/battles/queue"

const BattleIdPathVar = "battleId"
const TargetPlayerIdPathvar = "targetPlayer"

var ChallengeToBattleRoute = fmt.Sprintf(ChallengeToBattlePath, fmt.Sprintf("{%s}", TargetPlayerIdPathvar))
var AcceptChallengeRoute = fmt.Sprintf(AcceptChallengePath, fmt.Sprintf("{%s}", BattleIdPathVar))
