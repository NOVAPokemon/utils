package api

import "fmt"

const CreateGymPath = "/gym"
const GetGymInfoPath = "/gym/%s"
const CreateRaidPath = "/gym/%s/raid/create"
const JoinRaidPath = "/gym/%s/raid/join"

const GymIdPathVar = "gymId"

var CreateGymRoute = CreateGymPath
var GetGymInfoRoute = fmt.Sprintf(GetGymInfoPath, fmt.Sprintf("{%s}", GymIdPathVar))
var CreateRaidRoute = fmt.Sprintf(CreateRaidPath, fmt.Sprintf("{%s}", GymIdPathVar))
var JoinRaidRoute = fmt.Sprintf(JoinRaidPath, fmt.Sprintf("{%s}", GymIdPathVar))
