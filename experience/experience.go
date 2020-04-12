package experience

import (
	"math"
	"math/rand"
)

const (
	MinExperiencePerBattle = 100
	MaxBonusExperience     = 50
)

func GetPokemonTrainerGainFromBattle(winner bool) float64 {
	xp := MinExperiencePerBattle
	if winner {
		xp += rand.Intn(MaxBonusExperience)
	}
	return float64(xp)
}

func GetPokemonExperienceGainFromBattle(winner bool) float64 {
	xp := MinExperiencePerBattle
	if winner {
		xp += rand.Intn(MaxBonusExperience)
	}
	return float64(xp)
}

func CalculateLevel(xp float64) int {
	return int(math.Floor(25+math.Sqrt(625+100*xp)) / 50)
}
