package pokemon

import (
	"NOVAPokemon/utils"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

var pokemonMockup = utils.Pokemon {
	Id : primitive.NewObjectIDFromTimestamp(time.Now()),
	Owner : primitive.NewObjectIDFromTimestamp(time.Now()),
	Species: "toDelete",
	Level:   10,
	HP:      100,
	Damage:  1000,
}

func TestAddPokemon(t *testing.T) {

	err, res := AddPokemon(pokemonMockup)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	t.Log(res)
}

func TestGetAll(t *testing.T) {
	res := GetAllPokemons()
	for i, item := range res {
		t.Log(i, item)
	}
}

func TestGetByID(t *testing.T) {

	err, pokemon := GetPokemonById(pokemonMockup.Id)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	var eq = pokemonMockup == pokemon
	assert.True(t, eq)
	t.Log(pokemon)

}

func TestUpdate(t *testing.T) {

	toUpdate := utils.Pokemon{
		Species: "charizard:updated",
		Level:   10111,
		HP:      100111,
		Damage:  1000111,
	}

	err, _ := UpdatePokemon(pokemonMockup.Id, toUpdate)

	if err != nil {
		log.Error(err)
		t.Fail()
	}
	err, updatedPokemon := GetPokemonById(pokemonMockup.Id)

	assert.Equal(t, toUpdate.Owner, updatedPokemon.Owner)
	assert.Equal(t, toUpdate.Species, updatedPokemon.Species)
	assert.Equal(t, toUpdate.Damage, updatedPokemon.Damage)
	assert.Equal(t, toUpdate.HP, updatedPokemon.HP)
	assert.Equal(t, toUpdate.Level, updatedPokemon.Level)
}

func TestDelete(t *testing.T) {
	err := DeletePokemon(pokemonMockup.Id)

	if err != nil {
		log.Error(err)
		t.Fail()
	}

	pokemons := GetAllPokemons()

	for _, pokemon := range pokemons {
		if pokemon.Id == pokemonMockup.Id {
			t.Fail()
		}

	}
}
