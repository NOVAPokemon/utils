package user

import (
	"context"
	"github.com/NOVAPokemon/utils"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

const defaultMongoDBUrl = "mongodb://localhost:27017"
const databaseName = "NOVAPokemonDB"
const collectionName = "Users"

type DBClient struct {
	client     *mongo.Client
	collection *mongo.Collection
	ctx        *context.Context
}

var dbClient DBClient

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

func GetUserById(username string) (error, *utils.User) {

	var ctx = dbClient.ctx
	var collection = dbClient.collection
	result := &utils.User{}

	filter := bson.M{"username": username}
	err := collection.FindOne(*ctx, filter).Decode(result)

	if err != nil {
		log.Error(err)
		return err, nil
	}

	return nil, result
}

func AddUser(user *utils.User) (error, string) {

	var ctx = dbClient.ctx
	var collection = dbClient.collection
	res, err := collection.InsertOne(*ctx, *user)

	if err != nil {
		log.Error(err)
		return err, ""
	}

	return nil, res.InsertedID.(string)
}

func GetUserByUsername(username string) (error, *utils.User) {
	var ctx = dbClient.ctx
	var collection = dbClient.collection
	result := &utils.User{}

	filter := bson.M{"username": username}
	err := collection.FindOne(*ctx, filter).Decode(result)

	if err != nil {
		log.Error(err)
		return err, nil
	}

	return nil, result
}

func UpdateUser(username string, user *utils.User) (error, *utils.User) {

	ctx := dbClient.ctx
	collection := dbClient.collection
	filter := bson.M{"username": username}
	user.Username = username

	res, err := collection.ReplaceOne(*ctx, filter, *user)

	if err != nil {
		log.Error(err)
		return err, nil
	}

	if res.MatchedCount > 0 {
		log.Infof("Updated User %+v", username)
	} else {
		log.Errorf("Update User failed because no user matched %+v", username)
	}

	return nil, user

}

func DeleteUser(username string) error {
	var ctx = dbClient.ctx
	var collection = dbClient.collection
	filter := bson.M{"username": username}

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

	collection.Indexes().CreateOne(ctx, index)

	dbClient = DBClient{client: client, ctx: &ctx, collection: collection}
}
