package trainer

import (
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/pokemons"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var trainerMockup = utils.Trainer{
	Username: "trainer1",
	Pokemons: map[string]pokemons.Pokemon{},
	Items:    map[string]items.Item{},
	Stats: utils.TrainerStats{
		Level: 0,
		Coins: 0,
	},
}

var mockupLocation = utils.Location{
	RegionName: "/Europe/England/London/Greenwich",
	Latitude:   59.6,
	Longitude:  61.6,
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
	res, err := GetAllTrainers()
	if err != nil {
		log.Error(err)
		t.Fail()
	}

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

	trainer, _ := AddTrainer(trainerMockup)

	toUpdate := utils.TrainerStats{
		Level: 10,
		Coins: 10,
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

	assert.Equal(t, toUpdate.Level, updatedTrainer.Stats.Level)
	assert.Equal(t, toUpdate.Coins, updatedTrainer.Stats.Coins)

	_ = DeleteTrainer(trainer)
}

func TestUpdateRegion(t *testing.T) {

	trainer, _ := AddTrainer(trainerMockup)

	_, err := UpdateUserLocation(trainerMockup.Username, mockupLocation)

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	updatedTrainer, err := GetTrainerByUsername(trainerMockup.Username)

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	assert.Equal(t, mockupLocation.Longitude, updatedTrainer.Location.Longitude)
	assert.Equal(t, mockupLocation.Longitude, updatedTrainer.Location.Longitude)
	assert.Equal(t, mockupLocation.RegionName, updatedTrainer.Location.RegionName)

	_ = DeleteTrainer(trainer)
}

func TestDelete(t *testing.T) {

	_, _ = AddTrainer(trainerMockup)
	err := DeleteTrainer(trainerMockup.Username)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	trainers, err := GetAllTrainers()

	if err != nil {
		log.Error(err)
		t.Fail()
	}

	for _, trainer := range trainers {
		if trainer.Username == trainerMockup.Username {
			t.Error(trainer.Username)
			t.Fail()
		}

	}
}

func TestAddPokemonToTrainer(t *testing.T) {

	pokemon := pokemons.Pokemon{}
	_, _ = AddTrainer(trainerMockup)

	pokemonNew, _ := AddPokemonToTrainer(trainerMockup.Username, pokemon)
	trainer, _ := GetTrainerByUsername(trainerMockup.Username)

	assert.Contains(t, trainer.Pokemons, pokemonNew.Id.Hex())

	_ = DeleteTrainer(trainerMockup.Username)

}

func TestRemovePokemonFromTrainer(t *testing.T) {

	// add trainer and pokemon
	_, _ = AddTrainer(trainerMockup)
	pokemon, _ := AddPokemonToTrainer(trainerMockup.Username, pokemons.Pokemon{})
	trainer, _ := GetTrainerByUsername(trainerMockup.Username)

	assert.Contains(t, trainer.Pokemons, pokemon.Id.Hex())

	// remove pokemon from trainer
	_ = RemovePokemonFromTrainer(trainerMockup.Username, pokemon.Id)
	trainer, _ = GetTrainerByUsername(trainerMockup.Username)
	assert.NotContains(t, trainer.Pokemons, pokemon.Id.Hex())

	_ = DeleteTrainer(trainerMockup.Username)

}

func TestAppendAndRemoveItem(t *testing.T) {

	userName, err := AddTrainer(trainerMockup)

	toAppend := items.Item{
		Name: "Soup",
	}

	toAppend2 := items.Item{
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

func TestAppendAndRemovePokemon(t *testing.T) {

	userName, err := AddTrainer(trainerMockup)

	toAppend := pokemons.Pokemon{
	}

	toAppend2 := pokemons.Pokemon{
	}

	// add item, verify that it is in trainer
	pokemon, err := AddPokemonToTrainer(userName, toAppend)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	trainer, err := GetTrainerByUsername(userName)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.True(t, len(trainer.Pokemons) == 1)
	assert.Contains(t, trainer.Pokemons, pokemon.Id.Hex())

	// add another pokemon, verify that trainer has both pokemons
	pokemon2, err := AddPokemonToTrainer(userName, toAppend2)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	trainer, err = GetTrainerByUsername(userName)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.True(t, len(trainer.Pokemons) == 2)
	assert.Contains(t, trainer.Pokemons, pokemon.Id.Hex())
	assert.Contains(t, trainer.Pokemons, pokemon2.Id.Hex())

	// delete one pokemon, verify that trainer has  the remaining Pokemon
	err = RemovePokemonFromTrainer(userName, pokemon.Id)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	trainer, err = GetTrainerByUsername(userName)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.True(t, len(trainer.Pokemons) == 1)
	assert.NotContains(t, trainer.Pokemons, pokemon.Id.Hex())
	assert.Contains(t, trainer.Pokemons, pokemon2.Id.Hex())

	// Remove remaining pokemon, assure there are no pokemons remaining
	err = RemovePokemonFromTrainer(userName, pokemon2.Id)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	trainer, err = GetTrainerByUsername(userName)

	if err != nil {
		t.Error(err)
		t.Fail()
	}

	assert.True(t, len(trainer.Pokemons) == 0)
	_ = DeleteTrainer(userName)

}
