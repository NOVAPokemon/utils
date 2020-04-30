package user

import (
	"context"
	"errors"
	"github.com/NOVAPokemon/utils"
	databaseUtils "github.com/NOVAPokemon/utils/database"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

const databaseName = "NOVAPokemonDB"
const collectionName = "Users"

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

func GetAllUsers() ([]utils.User, error) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	var results []utils.User

	cur, err := collection.Find(*ctx, bson.M{})
	if err != nil {
		return nil, wrapGetAllUsersError(err)
	}

	defer databaseUtils.CloseCursor(cur, ctx)
	for cur.Next(*ctx) {
		var result utils.User
		err := cur.Decode(&result)
		if err != nil {
			return nil, wrapGetAllUsersError(err)
		} else {
			results = append(results, result)
		}
	}

	if err := cur.Err(); err != nil {
		return nil, wrapGetAllUsersError(err)
	}
	return results, nil
}

func AddUser(user *utils.User) (string, error) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	_, err := collection.InsertOne(*ctx, *user)

	if err != nil {
		return "", wrapAddUserError(err, user.Username)
	}

	return user.Username, nil
}

func CheckIfUserExists(username string) (bool, error) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	var limit int64 = 1

	filter := bson.M{"username": username}
	opts := options.CountOptions{Limit: &limit}

	count, err := collection.CountDocuments(*ctx, filter, &opts)
	if err != nil {
		return false, wrapUserExistsError(err, username)
	}

	return count > 0, nil
}

func GetUserByUsername(username string) (*utils.User, error) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection

	result := &utils.User{}
	filter := bson.M{"username": username}

	err := collection.FindOne(*ctx, filter).Decode(result)
	if err != nil {
		return nil, wrapGetUserError(err, username)
	}

	return result, nil
}

func UpdateUser(username string, user *utils.User) (*utils.User, error) {

	ctx := dbClient.Ctx
	collection := dbClient.Collection
	filter := bson.M{"username": username}
	user.Username = username

	res, err := collection.ReplaceOne(*ctx, filter, *user)

	if err != nil {
		return nil, wrapUpdateUserError(err, username)
	}

	if res.MatchedCount > 0 {
		log.Infof("Updated User %+v", username)
	} else {
		return nil, wrapUpdateUserError(errors.New("no user found"), username)
	}

	return user, nil

}

func DeleteUser(username string) error {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	filter := bson.M{"username": username}

	_, err := collection.DeleteOne(*ctx, filter)

	return wrapDeleteUserError(err, username)
}

func removeAll() error {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	filter := bson.M{}

	_, err := collection.DeleteMany(*ctx, filter)

	return wrapDeleteAllUsersError(err)
}
