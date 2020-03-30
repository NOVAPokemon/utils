package user

import (
	"context"
	"github.com/NOVAPokemon/utils"
	databaseUtils"github.com/NOVAPokemon/utils/database"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

const defaultMongoDBUrl = "mongodb://localhost:27017"
const databaseName = "NOVAPokemonDB"
const collectionName = "Users"



var dbClient databaseUtils.DBClient

func GetAllUsers() []utils.User {

	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
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

func AddUser(user *utils.User) (error, string) {

	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	_, err := collection.InsertOne(*ctx, *user)

	if err != nil {
		log.Error(err)
		return err, ""
	}

	return nil, user.Username
}

func GetUserByUsername(username string) (error, *utils.User) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
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

	ctx := dbClient.Ctx
	collection := dbClient.Collection
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
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	filter := bson.M{"username": username}

	_, err := collection.DeleteOne(*ctx, filter)

	if err != nil {
		log.Error(err)
	}

	return err
}

func removeAll() error {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	filter := bson.M{}

	_, err := collection.DeleteMany(*ctx, filter)

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

	dbClient = databaseUtils.DBClient{Client: client, Ctx: &ctx, Collection: collection}
}
