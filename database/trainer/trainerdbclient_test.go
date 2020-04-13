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
	Pokemons: map[string]utils.Pokemon{},
	Items:    map[string]utils.Item{},
	Stats: utils.TrainerStats{
		Level: 0,
		Coins: 0,
	},
}

var mockupLocation = utils.Location{
	Latitude:  59.6,
	Longitude: 61.6,
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

	pokemon := utils.Pokemon{}
	_, _ = AddTrainer(trainerMockup)

	newPokemons, _ := AddPokemonToTrainer(trainerMockup.Username, pokemon)

	if len(newPokemons) == 0 {
		t.FailNow()
	}
	_ = DeleteTrainer(trainerMockup.Username)

}

func TestRemovePokemonFromTrainer(t *testing.T) {

	// add trainer and pokemon
	_, _ = AddTrainer(trainerMockup)

	pokemons, err := AddPokemonToTrainer(trainerMockup.Username, utils.Pokemon{})

	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	if len(pokemons) == 0 {
		t.FailNow()
		return
	}
	addedPokemon := utils.Pokemon{}
	for _, v := range pokemons {
		addedPokemon = v
		break
	}

	// remove pokemon from trainer
	pokemons, _ = RemovePokemonFromTrainer(trainerMockup.Username, addedPokemon.Id)
	if len(pokemons) != 0 {
		t.FailNow()
	}
	_ = DeleteTrainer(trainerMockup.Username)

}

func TestAppendAndRemoveItem(t *testing.T) {

	userName, err := AddTrainer(trainerMockup)

	toAppend := utils.Item{
		Name: "Soup",
	}

	toAppend2 := utils.Item{
		Name: "Soup",
	}

	// add item, verify that it is in trainer
	items, err := AddItemToTrainer(userName, toAppend)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.True(t, len(items) == 1)

	// add another item, verify that trainer has both items
	items, err = AddItemToTrainer(userName, toAppend2)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.True(t, len(items) == 2)

	toRemove := primitive.ObjectID{}
	for _, v := range items {
		toRemove = v.Id
	}

	// delete one item, verify that trainer has  the remaining Item
	items, err = RemoveItemFromTrainer(userName, toRemove)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.True(t, len(items) == 1)

	toRemove = primitive.ObjectID{}
	for _, v := range items {
		toRemove = v.Id
	}

	// Remove remaining item, assure there are no items remaining
	items, err = RemoveItemFromTrainer(userName, toRemove)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.True(t, len(items) == 0)
	_ = DeleteTrainer(userName)

}

func TestAppendAndRemovePokemon(t *testing.T) {

	userName, err := AddTrainer(trainerMockup)

	toAppend := utils.Pokemon{
	}

	toAppend2 := utils.Pokemon{
	}

	// add item, verify that it is in trainer
	pokemons, err := AddPokemonToTrainer(userName, toAppend)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.True(t, len(pokemons) == 1)

	// add another pokemon, verify that trainer has both pokemons
	pokemons, err = AddPokemonToTrainer(userName, toAppend2)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	toRemove := primitive.ObjectID{}
	for _, v := range pokemons {
		toRemove = v.Id
	}

	// delete one pokemon, verify that trainer has  the remaining Pokemon
	pokemons, err = RemovePokemonFromTrainer(userName, toRemove)

	assert.True(t, len(pokemons) == 1)

	toRemove = primitive.ObjectID{}
	for _, v := range pokemons {
		toRemove = v.Id
	}

	pokemons, err = RemovePokemonFromTrainer(userName, toRemove)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.True(t, len(pokemons) == 0)
	_ = DeleteTrainer(userName)

}
