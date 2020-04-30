package trainer

import (
	"context"
	"github.com/NOVAPokemon/utils"
	databaseUtils "github.com/NOVAPokemon/utils/database"
	"github.com/NOVAPokemon/utils/experience"
	"github.com/NOVAPokemon/utils/items"
	"github.com/NOVAPokemon/utils/pokemons"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

const databaseName = "NOVAPokemonDB"
const collectionName = "Trainers"

var dbClient databaseUtils.DBClient

func init() {
	url, exists := os.LookupEnv(utils.MongoEnvVar)
	if !exists {
		url = databaseUtils.DefaultMongoDBUrl
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
	dbClient = databaseUtils.DBClient{Client: client, Ctx: &ctx, Collection: collection}
}

func AddTrainer(trainer utils.Trainer) (string, error) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection

	_, err := collection.InsertOne(*ctx, trainer)
	if err != nil {
		return "", wrapAddTrainerError(err, trainer.Username)
	} else {
		log.Infof("Added new trainer: %s", trainer.Username)
		return trainer.Username, nil
	}

}

func GetAllTrainers() ([]utils.Trainer, error) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	var results = make([]utils.Trainer, 0)

	cur, err := collection.Find(*ctx, bson.M{})
	if err != nil {
		return nil, wrapGetAllTrainersError(err)
	}

	defer databaseUtils.CloseCursor(cur, ctx)
	for cur.Next(*ctx) {
		var result utils.Trainer
		err := cur.Decode(&result)
		if err != nil {
			return nil, wrapGetAllTrainersError(err)
		} else {
			results = append(results, result)
		}
	}

	return results, wrapGetAllTrainersError(cur.Err())
}

func GetTrainerByUsername(username string) (*utils.Trainer, error) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	var result utils.Trainer

	filter := bson.M{"username": username}
	err := collection.FindOne(*ctx, filter).Decode(&result)

	if err != nil {
		return nil, wrapGetTrainerError(err, username)
	}

	return &result, nil
}

func UpdateTrainerStats(username string, stats utils.TrainerStats) (*utils.TrainerStats, error) {
	ctx := dbClient.Ctx
	collection := dbClient.Collection

	if stats.Level < 0 {
		return nil, wrapUpdateTrainerStatsError(ErrorInvalidLevel, username)
	}

	if stats.Coins < 0 {
		return nil, wrapUpdateTrainerStatsError(ErrorInvalidCoins, username)
	}

	filter := bson.M{"username": username}
	stats.Level = experience.CalculateLevel(stats.XP)
	changes := bson.M{"$set": bson.M{"stats.level": stats.Level, "stats.coins": stats.Coins, "stats.xp": stats.XP}}

	res, err := collection.UpdateOne(*ctx, filter, changes)
	if err != nil {
		return nil, wrapUpdateTrainerStatsError(err, username)
	}

	if res.MatchedCount > 0 {
		log.Infof("Updated Trainer %s", username)
	} else {
		return nil, wrapUpdateTrainerStatsError(ErrorTrainerNotFound, username)
	}

	return &stats, nil
}

func DeleteTrainer(username string) error {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection

	filter := bson.M{"username": username}
	_, err := collection.DeleteOne(*ctx, filter)

	return wrapDeleteTrainerError(err, username)
}

func removeAll() error {
	ctx := dbClient.Ctx
	collection := dbClient.Collection
	_, err := collection.DeleteMany(*ctx, bson.M{})

	return wrapDeleteAllTrainersError(err)
}

// BAG OPERATIONS

func AddItemToTrainer(username string, item items.Item) (map[string]items.Item, error) {
	ctx := dbClient.Ctx
	collection := dbClient.Collection

	itemId := primitive.NewObjectID()
	item.Id = itemId

	filter := bson.M{"username": username}
	change := bson.M{"$set": bson.M{"items." + itemId.Hex(): item}}
	opts := &options.FindOneAndUpdateOptions{}
	opts.SetReturnDocument(options.After)

	res := collection.FindOneAndUpdate(*ctx, filter, change, opts)
	if res.Err() != nil {
		return nil, wrapAddItemToTrainerError(ErrorTrainerNotFound, username)
	}

	trainer := utils.Trainer{}
	err := res.Decode(&trainer)
	return trainer.Items, wrapAddItemToTrainerError(err, username)
}

func AddItemsToTrainer(username string, itemsToAdd []items.Item) (map[string]items.Item, error) {
	ctx := dbClient.Ctx
	collection := dbClient.Collection

	itemsObjects := make(map[string]items.Item, len(itemsToAdd))

	for _, item := range itemsToAdd {
		itemsObjects["items."+item.Id.Hex()] = item
	}

	filter := bson.M{"username": username}
	change := bson.M{"$set": itemsObjects}
	opts := &options.FindOneAndUpdateOptions{}
	opts.SetReturnDocument(options.After)

	res := collection.FindOneAndUpdate(*ctx, filter, change, opts)
	if res.Err() != nil {
		return nil, wrapAddItemsToTrainerError(ErrorTrainerNotFound, username)
	}

	log.Infof("Added items to user %s:", username)
	for _, item := range itemsToAdd {
		log.Info(item.Id)
	}

	trainer := utils.Trainer{}
	err := res.Decode(&trainer)
	return trainer.Items, wrapAddItemsToTrainerError(err, username)
}

func RemoveItemFromTrainer(username string, itemId primitive.ObjectID) (map[string]items.Item, error) {
	ctx := dbClient.Ctx
	collection := dbClient.Collection
	filter := bson.M{"username": username}
	change := bson.M{"$unset": bson.M{"items." + itemId.Hex(): nil}}
	opts := &options.FindOneAndUpdateOptions{}
	opts.SetReturnDocument(options.After)

	res := collection.FindOneAndUpdate(*ctx, filter, change, opts)
	if res.Err() != nil {
		return nil, wrapRemoveItemToTrainerError(ErrorTrainerNotFound, username)
	}

	trainer := utils.Trainer{}
	err := res.Decode(&trainer)
	return trainer.Items, wrapRemoveItemToTrainerError(err, username)
}

func RemoveItemsFromTrainer(username string, itemIds []primitive.ObjectID) (map[string]items.Item, error) {
	ctx := dbClient.Ctx
	collection := dbClient.Collection
	filter := bson.M{"username": username}

	itemsObjects := make(map[string]*struct{}, len(itemIds))
	for _, id := range itemIds {
		itemsObjects["items."+id.Hex()] = nil
	}

	change := bson.M{"$unset": itemsObjects}
	opts := &options.FindOneAndUpdateOptions{}
	opts.SetReturnDocument(options.Before)

	res := collection.FindOneAndUpdate(*ctx, filter, change, opts)
	if res.Err() != nil {
		return nil, wrapRemoveItemsToTrainerError(ErrorTrainerNotFound, username)
	}

	trainer := utils.Trainer{}
	err := res.Decode(&trainer)
	return trainer.Items, wrapRemoveItemsToTrainerError(err, username)
}

// POKEMON OPERATIONS

func AddPokemonToTrainer(username string, pokemon pokemons.Pokemon) (map[string]pokemons.Pokemon, error) {
	ctx := dbClient.Ctx
	collection := dbClient.Collection

	filter := bson.M{"username": username}
	change := bson.M{"$set": bson.M{"pokemons." + pokemon.Id.Hex(): pokemon}}
	opts := &options.FindOneAndUpdateOptions{}
	opts.SetReturnDocument(options.After)

	res := collection.FindOneAndUpdate(*ctx, filter, change, opts)
	if res.Err() != nil {
		return nil, wrapAddPokemonToTrainerError(ErrorTrainerNotFound, username)
	}

	trainer := utils.Trainer{}
	err := res.Decode(&trainer)
	return trainer.Pokemons, wrapAddPokemonToTrainerError(err, username)
}

func UpdateTrainerPokemon(username string, pokemonId primitive.ObjectID, pokemon pokemons.Pokemon) (map[string]pokemons.Pokemon, error) {

	ctx := dbClient.Ctx
	collection := dbClient.Collection

	filter := bson.M{"username": username}
	change := bson.M{"$set": bson.M{"pokemons." + pokemonId.Hex(): pokemon}}
	opts := &options.FindOneAndUpdateOptions{}
	opts.SetReturnDocument(options.After)
	pokemon.Id = pokemonId

	res := collection.FindOneAndUpdate(*ctx, filter, change, opts)
	if res.Err() != nil {
		return nil, wrapUpdateTrainerPokemonError(ErrorTrainerNotFound, username)
	}

	trainer := utils.Trainer{}
	err := res.Decode(&trainer)
	return trainer.Pokemons, wrapUpdateTrainerPokemonError(err, username)
}

func RemovePokemonFromTrainer(username string, pokemonId primitive.ObjectID) (map[string]pokemons.Pokemon, error) {
	ctx := dbClient.Ctx
	collection := dbClient.Collection

	filter := bson.M{"username": username}
	change := bson.M{"$unset": bson.M{"pokemons." + pokemonId.Hex(): nil}}
	opts := &options.FindOneAndUpdateOptions{}
	opts.SetReturnDocument(options.Before)

	res := collection.FindOneAndUpdate(*ctx, filter, change, opts)
	if res.Err() != nil {
		return nil, wrapRemovePokemonFromTrainerError(ErrorTrainerNotFound, username)
	}

	trainer := utils.Trainer{}
	err := res.Decode(&trainer)
	return trainer.Pokemons, wrapRemovePokemonFromTrainerError(err, username)
}
