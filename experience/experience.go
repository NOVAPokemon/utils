package experience

import (
	"math"
	"math/rand"
)

const (
	MinExperiencePerBattle = 100
	MinExperiencePerRaid   = 150
	MaxBonusExperience     = 50
)

func GetTrainerExperienceGainFromBattle(winner bool) float64 {
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

func GetPokemonExperienceGainFromRaid(winner bool) float64 {
	xp := MinExperiencePerBattle
	if winner {
		xp += rand.Intn(MaxBonusExperience)
	}
	return float64(xp)
}

func CalculateLevel(xp float64) int {
	return int(math.Sqrt(xp))
}

// int(math.Floor(25+math.Sqrt(625+100*xp)) / 50)
func GetMinXpForLevel(level float64) float64 {
	if level == 0 {
		return 0
	}
	return level * level
}
