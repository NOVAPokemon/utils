package pokemons

import (
	"github.com/NOVAPokemon/utils/experience"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
)

func GetOneWildPokemon(maxLevel float64, stdHPDeviation float64, maxHP float64, stdDamageDeviation float64,
	maxDamage float64, species string) *Pokemon {

	var level, hp, damage int
	level = rand.Intn(int(maxLevel-1)) + 1
	randNormal := rand.NormFloat64()*
		(stdHPDeviation*(float64(level)/maxLevel)) +
		(maxHP * (float64(level) / maxLevel))
	hp = int(randNormal)

	//safeguards
	if hp < 1 {
		hp = 1
	}

	randNormal = rand.NormFloat64()*
		(stdDamageDeviation*(float64(level)/maxLevel)) +
		(maxDamage * (float64(level) / maxLevel))

	damage = int(randNormal)

	//safeguards
	if damage < 1 {
		damage = 1
	}

	wildPokemon := &Pokemon{
		Id:      primitive.NewObjectID(),
		Species: species,
		Level:   level,
		XP:      experience.GetMinXpForLevel(float64(level)),
		HP:      hp,
		MaxHP:   hp,
		Damage:  damage,
	}
	return wildPokemon
}

func GenerateRaidBoss(maxLevel float64, stdHPDeviation float64, maxHP float64, stdDamageDeviation float64,
	maxDamage float64, species string) *Pokemon {
	generated := GetOneWildPokemon(maxLevel*2, stdHPDeviation, maxHP*3, stdDamageDeviation, maxDamage, species)
	return generated
}
