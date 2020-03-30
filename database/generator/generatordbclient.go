package generator

import (
	"context"
	"github.com/NOVAPokemon/utils"
	databaseUtils "github.com/NOVAPokemon/utils/database"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

const defaultMongoDBUrl = "mongodb://localhost:27017"
const databaseName = "NOVAPokemonDB"
const wildPokemonCollectionName = "WildPokemons"
const catchableItemsCollectionName = "CatchableItems"

var dbClient databaseUtils.DBClientMultipleCollections

func AddWildPokemon(pokemon utils.Pokemon) (error, primitive.ObjectID) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collections[wildPokemonCollectionName]
	res, err := collection.InsertOne(*ctx, pokemon)

	if err != nil {
		log.Error(err)
		return nil, [12]byte{}
	}

	log.Infof("Inserted new wild Pokemon %s", res.InsertedID)

	return err, res.InsertedID.(primitive.ObjectID)
}

func DeleteWildPokemons() error {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collections[wildPokemonCollectionName]
	filter := bson.M{}

	_, err := collection.DeleteMany(*ctx, filter)

	if err != nil {
		log.Error(err)
	}

	return err
}

func GetCatchableItems() []utils.Item {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collections[catchableItemsCollectionName]
	var results []utils.Item

	cur, err := collection.Find(*ctx, bson.M{})

	if err != nil {
		log.Println(err)
	}

	defer cur.Close(*ctx)
	for cur.Next(*ctx) {
		var result utils.Item
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

func DeleteCatchableItems() error {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collections[catchableItemsCollectionName]
	filter := bson.M{}

	_, err := collection.DeleteMany(*ctx, filter)

	if err != nil {
		log.Error(err)
	}

	return err
}

func AddCatchableItem(item utils.Item) (error, primitive.ObjectID) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collections[catchableItemsCollectionName]
	res, err := collection.InsertOne(*ctx, item)

	if err != nil {
		log.Error(err)
		return err, [12]byte{}
	}

	log.Infof("Inserted new Catchable Item %+v", item)

	return err, res.InsertedID.(primitive.ObjectID)
}

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

	catchableItemsCollection := client.Database(databaseName).Collection(catchableItemsCollectionName)
	wildPokemons := client.Database(databaseName).Collection(wildPokemonCollectionName)

	collections := map[string]*mongo.Collection{
		wildPokemonCollectionName:          wildPokemons,
		catchableItemsCollectionName: catchableItemsCollection,
	}
	dbClient = databaseUtils.DBClientMultipleCollections{Client: client, Ctx: &ctx, Collections: collections}
}

