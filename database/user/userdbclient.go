package user

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
const collectionName = "Users"
const timeoutSeconds = 10

type DBCLient struct {
	client     *mongo.Client
	collection *mongo.Collection
	ctx        *context.Context
}

var dbClient DBCLient

func GetAllUsers() []utils.User {

	var ctx = dbClient.ctx
	var collection = dbClient.collection
	var results []utils.User

	cur, err := collection.Find(*ctx, bson.M{})

	if err != nil {
		log.Error(err)
	}

	defer cur.Close(*ctx)
	for cur.Next(*ctx) {
		var result utils.User
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

func GetUserById(id primitive.ObjectID) (error, *utils.User) {

	var ctx = dbClient.ctx
	var collection = dbClient.collection
	result := &utils.User{}

	filter := bson.M{"_id": id}
	err := collection.FindOne(*ctx, filter).Decode(result)

	if err != nil {
		log.Error(err)
		return err, nil
	}

	return nil, result
}

func AddUser(user utils.User) (error, primitive.ObjectID) {

	var ctx = dbClient.ctx
	var collection = dbClient.collection
	res, err := collection.InsertOne(*ctx, user)

	if err != nil {
		log.Println(err)
		return err, [12]byte{}
	}

	return err, res.InsertedID.(primitive.ObjectID)
}

func UpdateUser(id primitive.ObjectID, user *utils.User) (error, *utils.User) {

	ctx := dbClient.ctx
	collection := dbClient.collection
	filter := bson.M{"_id": id}
	user.Id = id

	res, err := collection.ReplaceOne(*ctx, filter, user)

	if err != nil {
		log.Error(err)
		return err, nil
	}

	if res.MatchedCount > 0 {
		log.Infof("Updated User %+v", id)
	} else {
		log.Errorf("Update User failed because no user matched %+v", id)
	}

	return nil, user

}

func DeleteUser(id primitive.ObjectID) error {

	var ctx = dbClient.ctx
	var collection = dbClient.collection
	filter := bson.M{"_id": id}

	_, err := collection.DeleteOne(*ctx, filter)

	if err != nil {
		log.Error(err)
		return err
	}

	return nil
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
