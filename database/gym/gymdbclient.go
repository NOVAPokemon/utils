package location

import (
	"context"
	"github.com/NOVAPokemon/utils"
	databaseUtils "github.com/NOVAPokemon/utils/database"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

const databaseName = "NOVAPokemonDB"
const gymsConfigCollectionName = "GymConfigs"

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
	gymsLocationCollection := client.Database(databaseName).Collection(gymsConfigCollectionName)
	dbClient = databaseUtils.DBClient{Client: client, Ctx: &ctx, Collection: gymsLocationCollection}
}

func GetGymsForServer(serverName string) ([]utils.Gym, error) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	var filter = bson.D{{"ServerName", serverName}}
	cursor, err := collection.Find(*ctx, filter)
	if err != nil {
		return nil, wrapGetServerConfig(err, serverName)
	}

	var gymsForServer []utils.Gym
	defer log.Error(cursor.Close(*ctx))
	for cursor.Next(*ctx) {
		var elem utils.Gym
		err := cursor.Decode(&elem)
		if err != nil {
			log.Fatal(wrapGetServerConfig(err, serverName))
		}
		gymsForServer = append(gymsForServer, elem)
	}
	return gymsForServer, nil
}

func UpsertGymWithServer(gym utils.GymWithServer) error {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	var filter = bson.D{{"ServerName", gym.ServerName}}
	upsert := true
	updateOptions := options.ReplaceOptions{
		Upsert: &upsert,
	}
	_, err := collection.ReplaceOne(*ctx, filter, gym, &updateOptions)
	if err != nil {
		return wrapUpdateServerConfig(err, gym.ServerName)
	}

	return nil
}
