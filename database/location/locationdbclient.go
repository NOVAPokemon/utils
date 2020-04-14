package generator

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

const defaultMongoDBUrl = "mongodb://localhost:27017"
const databaseName = "NOVAPokemonDB"
const usersLocationCollectionName = "UsersLocation"
const gymsLocationCollectionName = "GymsLocation"

var (
	dbClient           databaseUtils.DBClientMultipleCollections
	ErrTrainerNotFound = errors.New("trainer Not Found")
)

func UpdateUserLocation(username string, loc utils.Location) (*utils.Location, error) {
	ctx := dbClient.Ctx
	collection := dbClient.Collections[usersLocationCollectionName]

	filter := bson.M{"username": username}
	changes := bson.M{"$set": bson.M{"location": loc}}

	res, err := collection.UpdateOne(*ctx, filter, changes)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if res.MatchedCount > 0 {
		log.Infof("Updated Trainer %s location", username)
	} else {
		return nil, ErrTrainerNotFound
	}

	return &loc, nil
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

	usersLocationCollection := client.Database(databaseName).Collection(usersLocationCollectionName)
	gymsLocationCollection := client.Database(databaseName).Collection(gymsLocationCollectionName)

	collections := map[string]*mongo.Collection{
		usersLocationCollectionName: usersLocationCollection,
		gymsLocationCollectionName:  gymsLocationCollection,
	}

	dbClient = databaseUtils.DBClientMultipleCollections{Client: client, Ctx: &ctx, Collections: collections}
}
