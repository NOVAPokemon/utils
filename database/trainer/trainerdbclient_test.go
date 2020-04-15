package trainer

import (
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/pokemons"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	newPokemons, _ := AddPokemonToTrainer(trainerMockup.Username, pokemon)

	if len(newPokemons) == 0 {
		t.FailNow()
	}
	_ = DeleteTrainer(trainerMockup.Username)

}

func TestAppendRemovePokemon(t *testing.T) {

	// add trainer and pokemon
	_, _ = AddTrainer(trainerMockup)

	toAdd := pokemons.Pokemon{
		Id: primitive.NewObjectID(),
	}

	trainerPokemons, err := AddPokemonToTrainer(trainerMockup.Username, toAdd)

	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	_, ok := trainerPokemons[toAdd.Id.Hex()]

	if !ok {
		t.Error("Pokemon is not in added pokemons")
		t.FailNow()
	}

	// remove pokemon from trainer
	trainerPokemons, err = RemovePokemonFromTrainer(trainerMockup.Username, toAdd.Id)

	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	user, err := GetTrainerByUsername(trainerMockup.Username)
	_, ok = user.Pokemons[toAdd.Id.Hex()]

	if ok {
		t.Error("Pokemon was not removed")
		t.FailNow()
	}

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
	itemsAdded, err := AddItemToTrainer(userName, toAppend)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.True(t, len(itemsAdded) == 1)

	// add another item, verify that trainer has both itemsAdded
	itemsAdded, err = AddItemToTrainer(userName, toAppend2)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.True(t, len(itemsAdded) == 2)

	toRemove := primitive.ObjectID{}
	for _, v := range itemsAdded {
		toRemove = v.Id
	}

	// delete one item, verify that trainer has  the remaining Item
	itemsAdded, err = RemoveItemFromTrainer(userName, toRemove)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.True(t, len(itemsAdded) == 1)

	toRemove = primitive.ObjectID{}
	for _, v := range itemsAdded {
		toRemove = v.Id
	}

	// Remove remaining item, assure there are no itemsAdded remaining
	itemsAdded, err = RemoveItemFromTrainer(userName, toRemove)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.True(t, len(itemsAdded) == 0)
	_ = DeleteTrainer(userName)

}
