package trainer

import (
	"context"
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

const defaultMongoDBUrl = "mongodb://localhost:27017"
const databaseName = "NOVAPokemonDB"
const collectionName = "Trainers"

type DBCLient struct {
	client     *mongo.Client
	collection *mongo.Collection
	ctx        *context.Context
}

var dbClient DBCLient

func init() {

	url, exists := os.LookupEnv("MONGODB_URL")

	if !exists {
		url = defaultMongoDBUrl
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(url))

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	err = client.Connect(ctx)

	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database(databaseName).Collection(collectionName)

	op := options.Index()
	op.SetUnique(true)
	index := mongo.IndexModel{
		Keys:    bson.M{"username": 1},
		Options: op,
	}

	_, _ = collection.Indexes().CreateOne(ctx, index)
	dbClient = DBCLient{client: client, ctx: &ctx, collection: collection}
}

func AddTrainer(trainer utils.Trainer) (string, error) {

	var ctx = dbClient.ctx
	var collection = dbClient.collection
	_, err := collection.InsertOne(*ctx, trainer)

	if err != nil {
		log.Println(err)
		return "", err
	} else {
		log.Infof("Added new trainer: %+v", trainer)
		return trainer.Username, nil
	}

}

func GetAllTrainers() []utils.Trainer {

	var ctx = dbClient.ctx
	var collection = dbClient.collection
	var results = make([]utils.Trainer, 0)

	cur, err := collection.Find(*ctx, bson.M{})

	if err != nil {
		log.Error(err)
	}

	defer cur.Close(*ctx)
	for cur.Next(*ctx) {
		var result utils.Trainer
		err := cur.Decode(&result)
		if err != nil {
			log.Error(err)
		} else {
			results = append(results, result)
		}
	}

	if err := cur.Err(); err != nil {
		log.Error(err)
	}
	return results
}

func GetTrainerByUsername(username string) (*utils.Trainer, error) {

	var ctx = dbClient.ctx
	var collection = dbClient.collection
	var result utils.Trainer

	filter := bson.M{"username": username}
	err := collection.FindOne(*ctx, filter).Decode(&result)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &result, nil
}

func UpdateTrainer(username string, trainer utils.Trainer) (error, utils.Trainer) {

	ctx := dbClient.ctx
	collection := dbClient.collection
	filter := bson.M{"username": username}
	trainer.Username = username

	res, err := collection.ReplaceOne(*ctx, filter, trainer)

	if err != nil {
		log.Error(err)
		return err, utils.Trainer{}
	}

	if res.MatchedCount > 0 {
		log.Infof("Updated Trainer %+v", username)
	} else {
		log.Errorf("Update Trainer failed because no trainer matched %+v", username)
	}
	return err, trainer
}

func DeleteTrainer(username string) error {

	var ctx = dbClient.ctx
	var collection = dbClient.collection
	filter := bson.M{"username": username}

	_, err := collection.DeleteOne(*ctx, filter)

	if err != nil {
		log.Error(err)
	}

	return err
}

func removeAll() error {
	ctx := dbClient.ctx
	collection := dbClient.collection
	_, err := collection.DeleteMany(*ctx, bson.M{})

	if err != nil {
		log.Error(err)
	}

	return err
}

// BAG OPERATIONS

func AddItemToTrainer(username string, item utils.Item) (*utils.Item, error) {

	ctx := dbClient.ctx
	collection := dbClient.collection

	itemId := primitive.NewObjectID()
	item.Id = itemId

	filter := bson.M{"username": username}
	change := bson.M{"$set": bson.M{"items." + itemId.Hex(): item}}

	res, err := collection.UpdateOne(*ctx, filter, change)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	if res.MatchedCount > 0 {
		log.Infof("Added item %s to user: %s", item.Name, username)
	} else {
		return nil, errors.New(fmt.Sprintf("Update failed because no username matched: %s", username))
	}

	return &item, err
}

func AddItemsToTrainer(username string, items []*utils.Item) (*[]*utils.Item, error) {
	ctx := dbClient.ctx
	collection := dbClient.collection

	itemsObjects := make(map[string]*utils.Item, len(items))

	for _, item := range items {
		itemId := primitive.NewObjectID()
		item.Id = itemId
		itemsObjects["items." + item.Id.Hex()] = item
	}

	filter := bson.M{"username": username}
	change := bson.M{"$set": itemsObjects}

	_, err := collection.UpdateOne(*ctx, filter, change)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Update failed because no username matched: %s", username))
	} else {
		log.Infof("Added items to user %s:", username)
		for _, item := range items {
			log.Info(item.Id)
		}
	}

	return &items, err
}

func RemoveItemFromTrainer(username string, itemId primitive.ObjectID) (*utils.Item, error) {
	ctx := dbClient.ctx
	collection := dbClient.collection
	filter := bson.M{"username": username}
	change := bson.M{"$unset": bson.M{"items." + itemId.Hex(): nil}}

	oldTrainer := collection.FindOne(*ctx, filter)
	_, err := collection.UpdateOne(*ctx, filter, change)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Update failed because no username matched: %s", username))
	} else {
		log.Infof("Removed item %s from user: %s", itemId, username)
	}

	trainer := &utils.Trainer{}
	if err := oldTrainer.Decode(trainer); err != nil {
		log.Error(err)
		return nil, err
	}

	log.Info(trainer)

	return trainer.Items[itemId.Hex()], nil
}

func RemoveItemsFromTrainer(username string, itemIds []primitive.ObjectID) ([]*utils.Item, error) {
	ctx := dbClient.ctx
	collection := dbClient.collection
	filter := bson.M{"username": username}

	itemsObjects := make(map[string]*struct{}, len(itemIds))

	for _, id := range itemIds {
		itemsObjects["items." + id.Hex()] = nil
	}

	change := bson.M{"$unset": itemsObjects}

	oldTrainer := collection.FindOne(*ctx, filter)
	_, err := collection.UpdateOne(*ctx, filter, change)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Update failed because no username matched: %s", username))
	} else {
		log.Infof("Removed items from user %s:", username)
		for _, item := range itemIds {
			log.Info(item)
		}
	}

	trainer := &utils.Trainer{}
	if err := oldTrainer.Decode(trainer); err != nil {
		log.Error(err)
		return nil, err
	}

	returnItems := make([]*utils.Item, len(itemIds))
	for i, item := range itemIds {
		returnItems[i] = trainer.Items[item.Hex()]
	}

	return returnItems, nil
}

// POKEMON OPERATIONS

func AddPokemonToTrainer(username string, pokemon utils.Pokemon) (*utils.Pokemon, error) {

	ctx := dbClient.ctx
	collection := dbClient.collection

	pokemonId := primitive.NewObjectID()
	pokemon.Id = pokemonId

	filter := bson.M{"username": username}
	change := bson.M{"$set": bson.M{"pokemons." + pokemon.Id.Hex(): pokemon}}

	_, err := collection.UpdateOne(*ctx, filter, change)

	if err != nil {
		return nil, err
	}

	return &pokemon, err
}

func RemovePokemonFromTrainer(username string, pokemonId primitive.ObjectID) error {

	ctx := dbClient.ctx
	collection := dbClient.collection

	filter := bson.M{"username": username}
	change := bson.M{"$unset": bson.M{"pokemons." + pokemonId.Hex(): nil}}

	_, err := collection.UpdateOne(*ctx, filter, change)

	if err != nil {
		log.Error(err)
	}

	return err
}
