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

func GetGymsForServer(serverName string) (*utils.GymsForServer, error) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	var filter = bson.D{{"servername", serverName}}
	res := collection.FindOne(*ctx, filter)
	if res.Err() != nil {
		return nil, wrapGetServerConfig(res.Err(), serverName)
	}
	var serverConfig = &utils.GymsForServer{}
	if err := res.Decode(serverConfig); err != nil {
		return nil, wrapGetServerConfig(res.Err(), serverName)
	}
	return serverConfig, nil
}

func UpsertGymsForServer(serverName string, config utils.GymsForServer) error {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	var filter = bson.D{{"ServerName:", serverName}}
	upsert := true
	updateOptions := options.ReplaceOptions{
		Upsert: &upsert,
	}
	config.ServerName = serverName
	_, err := collection.ReplaceOne(*ctx, filter, config, &updateOptions)
	if err != nil {
		return wrapUpdateServerConfig(err, serverName)
	}

	return nil
}
