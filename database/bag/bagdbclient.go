package bag

import (
	"context"
	"github.com/NOVAPokemon/utils"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"time"
)

const defaultMongoDBUrl = "mongodb://localhost:27017"
const databaseName = "NOVAPokemonDB"
const collectionName = "Bags"
const timeoutSeconds = 10

type DBCLient struct {
	client     *mongo.Client
	collection *mongo.Collection
	ctx        *context.Context
}

var dbClient DBCLient

func GetAllBags() []utils.Bag {

	var ctx = dbClient.ctx
	var collection = dbClient.collection
	var results []utils.Bag

	cur, err := collection.Find(*ctx, bson.M{})

	if err != nil {
		log.Println(err)
	}

	defer cur.Close(*ctx)
	for cur.Next(*ctx) {
		var result utils.Bag
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

func GetBagById(id primitive.ObjectID) (error, utils.Bag) {

	var ctx = dbClient.ctx
	var collection = dbClient.collection
	var result utils.Bag

	filter := bson.M{"_id": id}
	err := collection.FindOne(*ctx, filter).Decode(&result)

	if err != nil {
		log.Error(err)
		return err, utils.Bag{}
	}

	return err, result
}

func AddBag(bag utils.Bag) (error, primitive.ObjectID) {

	var ctx = dbClient.ctx
	var collection = dbClient.collection
	res, err := collection.InsertOne(*ctx, bag)

	if err != nil {
		log.Error(err)
		return err, [12]byte{}
	}

	log.Infof("Inserted new Bag %+v", bag)

	return err, res.InsertedID.(primitive.ObjectID)
}

func UpdateBag(id primitive.ObjectID, bag utils.Bag) error {

	ctx := dbClient.ctx
	collection := dbClient.collection
	filter := bson.M{"_id": id}
	bag.Id = id

	res, err := collection.ReplaceOne(*ctx, filter, bag)

	if err != nil {
		log.Error(err)
	}

	if res.MatchedCount > 0 {
		log.Infof("Updated Bag %s %+v", id, bag)
	} else {
		log.Errorf("Update Bag failed because no bag matched %s", id)
	}

	return err

}

func AppendItem(bagId primitive.ObjectID, item utils.Item) error {

	ctx := dbClient.ctx
	collection := dbClient.collection
	filter := bson.M{"_id": bagId}
	change := bson.M{"$push": bson.M{"items": item.Id}}

	res, err := collection.UpdateOne(*ctx, filter, change)

	if err != nil {
		log.Error(err)
		return err
	}

	if res.MatchedCount > 0 {
		log.Infof("Updated Bag %s", bagId)
	} else {
		log.Errorf("Update Bag failed because no bag matched %s", bagId)
	}

	return err
}

func RemoveItem(bagId primitive.ObjectID, itemId primitive.ObjectID) error {

	ctx := dbClient.ctx
	collection := dbClient.collection
	filter := bson.M{"_id": bagId}
	change := bson.M{"$pull": bson.M{"items": itemId}} // TODO

	res, err := collection.UpdateOne(*ctx, filter, change)

	if err != nil {
		log.Error(err)
	}

	if res.MatchedCount > 0 {
		log.Infof("Updated Bag %s", bagId)
	} else {
		log.Errorf("Update Bag failed because no bag matched %s", bagId)
	}

	return err
}

func DeleteBag(id primitive.ObjectID) error {

	var ctx = dbClient.ctx
	var collection = dbClient.collection
	filter := bson.M{"_id": id}

	_, err := collection.DeleteOne(*ctx, filter)

	if err != nil {
		log.Error(err)
	}

	return err
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

	ctx, _ := context.WithTimeout(context.Background(), timeoutSeconds*time.Second)
	err = client.Connect(ctx)

	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database(databaseName).Collection(collectionName)
	dbClient = DBCLient{client: client, ctx: &ctx, collection: collection}
}
