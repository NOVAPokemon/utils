package trainer

import (
	"github.com/NOVAPokemon/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var trainerMockup = utils.Trainer{
	Username: "trainer1",
	Pokemons: map[string]utils.Pokemon{},
	Items:    map[string]utils.Item{},
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
	res, err := AddTrainer(trainerMockup)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	t.Log("added: " + res)

	_ = DeleteTrainer(res)
}

func TestGetAll(t *testing.T) {
	res := GetAllTrainers()
	for i, item := range res {
		t.Log(i, item)
	}
}

func TestGetByUsername(t *testing.T) {

	_, _ = AddTrainer(trainerMockup)
	trainer, err := GetTrainerByUsername(trainerMockup.Username)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	t.Log(trainer)
	_ = DeleteTrainer(trainer.Username)

}

func TestUpdate(t *testing.T) {

	_, _ = AddTrainer(trainerMockup)

	toUpdate := utils.Trainer{
		Username: trainerMockup.Username,
		Level:    10,
		Coins:    10,
	}

	_, err := UpdateTrainerStats(trainerMockup.Username, toUpdate)

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	updatedTrainer, err := GetTrainerByUsername(trainerMockup.Username)

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	assert.Equal(t, toUpdate.Level, updatedTrainer.Level)
	assert.Equal(t, toUpdate.Coins, updatedTrainer.Coins)

	_ = DeleteTrainer(toUpdate.Username)
}

func TestDelete(t *testing.T) {

	_, _ = AddTrainer(trainerMockup)
	err := DeleteTrainer(trainerMockup.Username)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	trainers := GetAllTrainers()

	for _, trainer := range trainers {
		if trainer.Username == trainerMockup.Username {
			t.Error(trainer.Username)
			t.Fail()
		}

	}
}

func TestAddPokemonToTrainer(t *testing.T) {

	pokemon := utils.Pokemon{}
	_, _ = AddTrainer(trainerMockup)

	pokemonNew, _ := AddPokemonToTrainer(trainerMockup.Username, pokemon)
	trainer, _ := GetTrainerByUsername(trainerMockup.Username)

	assert.Contains(t, trainer.Pokemons, pokemonNew.Id.Hex())

	_ = DeleteTrainer(trainerMockup.Username)

}

func TestRemovePokemonFromTrainer(t *testing.T) {

	// add trainer and pokemon
	_, _ = AddTrainer(trainerMockup)
	pokemon, _ := AddPokemonToTrainer(trainerMockup.Username, utils.Pokemon{})
	trainer, _ := GetTrainerByUsername(trainerMockup.Username)

	assert.Contains(t, trainer.Pokemons, pokemon.Id.Hex())

	// remove pokemon from trainer
	_ = RemovePokemonFromTrainer(trainerMockup.Username, pokemon.Id)
	trainer, _ = GetTrainerByUsername(trainerMockup.Username)
	assert.NotContains(t, trainer.Pokemons, pokemon.Id.Hex())

	_ = DeleteTrainer(trainerMockup.Username)

}

func TestAppendAndRemove(t *testing.T) {

	userName, err := AddTrainer(trainerMockup)

	toAppend := utils.Item{
		Name: "Soup",
	}

	toAppend2 := utils.Item{
		Name: "Soup",
	}

	// add item, verify that it is in trainer
	item, err := AddItemToTrainer(userName, toAppend)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	trainer, err := GetTrainerByUsername(userName)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.True(t, len(trainer.Items) == 1)
	assert.Contains(t, trainer.Items, item.Id.Hex())

	// add another item, verify that trainer has both items
	item2, err := AddItemToTrainer(userName, toAppend2)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	trainer, err = GetTrainerByUsername(userName)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.True(t, len(trainer.Items) == 2)
	assert.Contains(t, trainer.Items, item.Id.Hex())
	assert.Contains(t, trainer.Items, item2.Id.Hex())

	// delete one item, verify that trainer has  the remaining Item
	_, err = RemoveItemFromTrainer(userName, item.Id)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	trainer, err = GetTrainerByUsername(userName)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.True(t, len(trainer.Items) == 1)
	assert.NotContains(t, trainer.Items, item.Id.Hex())
	assert.Contains(t, trainer.Items, item2.Id.Hex())

	// Remove remaining item, assure there are no items remaining
	_, err = RemoveItemFromTrainer(userName, item2.Id)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	trainer, err = GetTrainerByUsername(userName)

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	assert.True(t, len(trainer.Items) == 0)
	_ = DeleteTrainer(userName)

}
