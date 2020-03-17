package trainer

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
const collectionName = "Trainers"
const timeoutSeconds = 10

type DBCLient struct {
	client     *mongo.Client
	collection *mongo.Collection
	ctx        *context.Context
}

var dbClient DBCLient

func GetAllTrainers() []utils.Trainer {

	var ctx = dbClient.ctx
	var collection = dbClient.collection
	var results []utils.Trainer

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

func GetTrainerById(id primitive.ObjectID) (error, utils.Trainer) {

	var ctx = dbClient.ctx
	var collection = dbClient.collection
	var result utils.Trainer

	filter := bson.M{"_id": id}
	err := collection.FindOne(*ctx, filter).Decode(&result)

	if err != nil {
		log.Error(err)
		return err, utils.Trainer{}
	}

	return err, result
}

func AddTrainer(trainer utils.Trainer) (error, primitive.ObjectID) {

	var ctx = dbClient.ctx
	var collection = dbClient.collection
	res, err := collection.InsertOne(*ctx, trainer)

	if err != nil {
		log.Println(err)
		return err, [12]byte{}
	}

	return err, res.InsertedID.(primitive.ObjectID)
}

func UpdateTrainer(id primitive.ObjectID, trainer utils.Trainer) (error, utils.Trainer) {

	ctx := dbClient.ctx
	collection := dbClient.collection
	filter := bson.M{"_id": id}
	trainer.Id = id

	res, err := collection.ReplaceOne(*ctx, filter, trainer)

	if err != nil {
		log.Error(err)
		return err, utils.Trainer{}
	}

	if res.MatchedCount > 0 {
		log.Infof("Updated Trainer %+v", id)
	} else {
		log.Errorf("Update Trainer failed because no trainer matched %+v", id)
	}
	return err, trainer
}

func AddPokemonToTrainer(trainerId primitive.ObjectID, pokemonId primitive.ObjectID) error {

	ctx := dbClient.ctx
	collection := dbClient.collection

	filter := bson.M{"_id": trainerId}
	change := bson.M{"$push": bson.M{"pokemons": pokemonId}}

	_, err := collection.UpdateOne(*ctx, filter, change)

	if err != nil {
		log.Error(err)
	}

	return err
}

func RemovePokemonFromTrainer(trainerId primitive.ObjectID, pokemonId primitive.ObjectID) error {

	ctx := dbClient.ctx
	collection := dbClient.collection

	filter := bson.M{"_id": trainerId}
	change := bson.M{"$pull": bson.M{"pokemons": pokemonId}}

	_, err := collection.UpdateOne(*ctx, filter, change)

	if err != nil {
		log.Error(err)
	}

	return err
}

func DeleteTrainer(id primitive.ObjectID) error {

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
