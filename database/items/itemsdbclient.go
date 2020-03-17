package items

import (
	"context"
	"github.com/NOVAPokemon/utils"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

const defaultMongoDBUrl = "mongodb://localhost:27017"
const databaseName = "NOVAPokemonDB"
const itemsCollectionName = "Items"
const catchableItemsCollectionName = "CatchableItems"

type DBCLient struct {
	client      *mongo.Client
	collections map[string]*mongo.Collection
	ctx         *context.Context
}

var dbClient DBCLient

func GetAllItems() []utils.Item {
	var ctx = dbClient.ctx
	var collection = dbClient.collections[itemsCollectionName]
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

func GetCatchableItems() []utils.Item {

	var ctx = dbClient.ctx
	var collection = dbClient.collections[catchableItemsCollectionName]
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
	var ctx = dbClient.ctx
	var collection = dbClient.collections[catchableItemsCollectionName]
	filter := bson.M{}

	_, err := collection.DeleteMany(*ctx, filter)

	if err != nil {
		log.Error(err)
	}

	return err
}

func AddCatchableItem(item utils.Item) (error, primitive.ObjectID) {
	var ctx = dbClient.ctx
	var collection = dbClient.collections[catchableItemsCollectionName]
	res, err := collection.InsertOne(*ctx, item)

	if err != nil {
		log.Error(err)
		return err, [12]byte{}
	}

	log.Infof("Inserted new Catchable Item %+v", item)

	return err, res.InsertedID.(primitive.ObjectID)
}

func AddItems(item utils.Item) (error, primitive.ObjectID) {
	var ctx = dbClient.ctx
	var collection = dbClient.collections[itemsCollectionName]
	res, err := collection.InsertOne(*ctx, item)

	if err != nil {
		log.Error(err)
		return err, [12]byte{}
	}

	log.Infof("Inserted new Item %+v", item)

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

	itemsCollection := client.Database(databaseName).Collection(itemsCollectionName)
	catchableItemsCollection := client.Database(databaseName).Collection(catchableItemsCollectionName)

	collections := map[string]*mongo.Collection{
		itemsCollectionName:          itemsCollection,
		catchableItemsCollectionName: catchableItemsCollection,
	}

	dbClient = DBCLient{client: client, ctx: &ctx, collections: collections}
}
