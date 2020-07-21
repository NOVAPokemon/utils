package location

import (
	"context"
	"errors"
	"os"

	"github.com/NOVAPokemon/utils"
	databaseUtils "github.com/NOVAPokemon/utils/database"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func GetGymsForServer(serverName string) ([]utils.GymWithServer, error) {
	var (
		ctx        = dbClient.Ctx
		collection = dbClient.Collection
		filter     = bson.D{{"servername", serverName}}
	)

	cursor, err := collection.Find(*ctx, filter)

	if err != nil {
		return nil, wrapGetServerConfig(err, serverName)
	}

	var gymsForServer []utils.GymWithServer
	defer func() {
		if err := cursor.Close(*ctx); err != nil {
			log.Error(err)
		}
	}()
	if err := cursor.All(*ctx, &gymsForServer); err != nil {
		return nil, wrapGetServerConfig(err, serverName)

	}
	if len(gymsForServer) == 0 {
		return nil, wrapGetServerConfig(errors.New("no gyms found"), serverName)

	}
	return gymsForServer, nil

}

func AddGymWithServer(gym utils.GymWithServer) error {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	filter := bson.M{"gym.name": gym.Gym.Name}
	upsert := true
	updateOptions := &options.ReplaceOptions{
		Upsert: &upsert,
	}
	_, err := collection.ReplaceOne(*ctx, filter, gym, updateOptions)

	if err != nil {
		return wrapUpdateServerConfig(err, gym.ServerName)
	}
	return nil
}
