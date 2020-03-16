package bag

import (
	"github.com/NOVAPokemon/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

var bagMockup = utils.Bag{
	Id:    primitive.NewObjectIDFromTimestamp(time.Now()),
	Items: []primitive.ObjectID{},
}

func TestAddBag(t *testing.T) {

	logrus.Println("Testing add bag")
	err, _ := AddBag(bagMockup)

	if err != nil {
		logrus.Println(err)
		t.Fail()
	}

}

func TestGetAll(t *testing.T) {

	logrus.Println("testing get all bags")
	res := GetAllBags()
	for i, item := range res {
		logrus.Println(i, item)
	}
}

func TestGetByID(t *testing.T) {

	logrus.Println("testing Get By ID")
	err, bag := GetBagById(bagMockup.Id)

	if err != nil {
		logrus.Println(err)
		t.Fail()
	}

	logrus.Println(bag)
}

func TestUpdate(t *testing.T) {

	logrus.Println("testing update")

	var items []primitive.ObjectID
	var itemId primitive.ObjectID

	itemId = primitive.NewObjectIDFromTimestamp(time.Now())
	items = append(items, itemId)

	toUpdate := utils.Bag{
		Items: items,
	}

	err := UpdateBag(bagMockup.Id, toUpdate)

	if err != nil {
		logrus.Error(err)
		t.Fail()
	}

	err, updatedBag := GetBagById(bagMockup.Id)
	assert.Contains(t, updatedBag.Items, toUpdate.Items[0])

}

func TestAppendAndRemove(t *testing.T) {

	logrus.Println("testing append and remove from bag")

	bagMockup.Id = primitive.NewObjectID()
	err, bagId := AddBag(bagMockup)

	if err != nil {
		logrus.Error(err)
		t.Fail()
	}

	itemId := primitive.NewObjectID()
	toAppend := utils.Item{
		Id: itemId,
	}

	itemId2 := primitive.NewObjectID()
	toAppend2 := utils.Item{
		Id: itemId2,
	}

	err = AppendItem(bagId, toAppend)

	if err != nil {
		logrus.Error(err)
		t.Fail()
	}

	err, updatedBag := GetBagById(bagId)

	if err != nil {
		logrus.Error(err)
		t.Fail()
	}

	assert.Contains(t, updatedBag.Items, itemId)

	err = AppendItem(bagId, toAppend2)

	if err != nil {
		logrus.Error(err)
		t.Fail()
	}

	err, updatedBag = GetBagById(bagId)

	if err != nil {
		logrus.Error(err)
		t.Fail()
	}

	assert.Contains(t, updatedBag.Items, itemId)
	assert.Contains(t, updatedBag.Items, itemId2)

	err = RemoveItem(bagId, itemId)

	if err != nil {
		logrus.Error(err)
		t.Fail()
	}

	err, updatedBag = GetBagById(bagId)

	if err != nil {
		logrus.Error(err)
		t.Fail()
	}
	assert.NotContains(t, updatedBag.Items, itemId)

	err = RemoveItem(bagId, itemId2)

	if err != nil {
		logrus.Error(err)
		t.Fail()
	}

	err, updatedBag = GetBagById(bagId)

	if err != nil {
		logrus.Error(err)
		t.Fail()
	}

	logrus.Println(updatedBag)
	assert.NotContains(t, updatedBag.Items, itemId)

}

func TestDelete(t *testing.T) {
	logrus.Println("testing delete")
	err := DeleteBag(bagMockup.Id)

	if err != nil {
		logrus.Error(err)
		t.Fail()
	}

	bags := GetAllBags()

	for _, bag := range bags {
		if bag.Id == bagMockup.Id {
			t.Fail()
		}

	}
}
