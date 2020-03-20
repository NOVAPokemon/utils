package trainer

import (
	"github.com/NOVAPokemon/utils"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
)

var trainerMockup = utils.Trainer{
	Username: "teste_trainer",
	Bag:      primitive.NewObjectID(),
	Pokemons: []primitive.ObjectID{},
	Level:    0,
	Coins:    0,
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

	err, trainer := GetTrainerById(trainerMockup.Username)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	t.Log(trainer)

}

func TestUpdate(t *testing.T) {

	toUpdate := utils.Trainer{
		Username: "teste_trainer_3",
		Bag:   primitive.NewObjectID(),
		Level: 10,
		Coins: 10,
	}

	err, _ := UpdateTrainer(trainerMockup.Username, toUpdate)

	if err != nil {
		log.Error(err)
		t.Fail()
	}

	err, updatedTrainer := GetTrainerById(trainerMockup.Username)

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
			t.Fail()
		}

	}
}

func TestAddPokemonToTrainer(t *testing.T) {

	pokemonId := primitive.NewObjectID()
	trainerMockup.Username = "teste_trainer_2"

	_, _ = AddTrainer(trainerMockup)
	_ = AddPokemonToTrainer(trainerMockup.Username, pokemonId)
	_, trainer := GetTrainerById(trainerMockup.Username)

	assert.Contains(t, trainer.Pokemons, pokemonId)

}

func TestRemovePokemonFromTrainer(t *testing.T) {

	pokemonId := primitive.NewObjectID()
	trainerMockup.Username = "test_trainer_2"

	// add trainer and pokemon
	_, _ = AddTrainer(trainerMockup)
	_ = AddPokemonToTrainer(trainerMockup.Username, pokemonId)
	_, trainer := GetTrainerById(trainerMockup.Username)
	assert.Contains(t, trainer.Pokemons, pokemonId)

	// remove pokemon from trainer
	_ = RemovePokemonFromTrainer(trainerMockup.Username, pokemonId)
	_, trainer = GetTrainerById(trainerMockup.Username)
	assert.NotContains(t, trainer.Pokemons, pokemonId)

}
