package trainer

import (
	"github.com/NOVAPokemon/utils"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
	"testing"
)

var trainerMockup = utils.Trainer{
	Username: "trainer1",
	Bag:      primitive.NewObjectID(),
	Pokemons: []primitive.ObjectID{},
	Level:    0,
	Coins:    0,
}

func TestMain(m *testing.M) {
	_ = removeAll()

	res := m.Run()

	_ = removeAll()

	os.Exit(res)
}

func TestAddTrainer(t *testing.T) {
	err, res := AddTrainer(trainerMockup)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	t.Log(res)
}

func TestGetAll(t *testing.T) {
	res := GetAllTrainers()
	for i, item := range res {
		t.Log(i, item)
	}
}

func TestGetByID(t *testing.T) {

	err, trainer := GetTrainerByUsername(trainerMockup.Username)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	t.Log(trainer)

}

func TestUpdate(t *testing.T) {
	toUpdate := utils.Trainer{
		Username: trainerMockup.Username,
		Bag:   primitive.NewObjectID(),
		Level: 10,
		Coins: 10,
	}

	err, _ := UpdateTrainer(trainerMockup.Username, toUpdate)

	if err != nil {
		log.Error(err)
		t.Fail()
	}

	err, updatedTrainer := GetTrainerByUsername(trainerMockup.Username)

	assert.Equal(t, toUpdate.Level, updatedTrainer.Level)
	assert.Equal(t, toUpdate.Coins, updatedTrainer.Coins)
}

func TestDelete(t *testing.T) {
	err := DeleteTrainer(trainerMockup.Username)

	if err != nil {
		log.Error(err)
		t.Fail()
	}

	trainers := GetAllTrainers()

	for _, trainer := range trainers {
		if trainer.Username == trainerMockup.Username {
			log.Error(trainer.Username)
			t.Fail()
		}

	}
}

func TestAddPokemonToTrainer(t *testing.T) {

	pokemonId := primitive.NewObjectID()

	_, _ = AddTrainer(trainerMockup)
	_ = AddPokemonToTrainer(trainerMockup.Username, pokemonId)
	_, trainer := GetTrainerByUsername(trainerMockup.Username)

	assert.Contains(t, trainer.Pokemons, pokemonId)

}

func TestRemovePokemonFromTrainer(t *testing.T) {

	pokemonId := primitive.NewObjectID()

	// add trainer and pokemon
	_, _ = AddTrainer(trainerMockup)
	_ = AddPokemonToTrainer(trainerMockup.Username, pokemonId)
	_, trainer := GetTrainerByUsername(trainerMockup.Username)
	assert.Contains(t, trainer.Pokemons, pokemonId)

	// remove pokemon from trainer
	_ = RemovePokemonFromTrainer(trainerMockup.Username, pokemonId)
	_, trainer = GetTrainerByUsername(trainerMockup.Username)
	assert.NotContains(t, trainer.Pokemons, pokemonId)

}
