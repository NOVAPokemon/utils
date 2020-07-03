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

func GetAllGyms() ([]utils.GymWithServer, error){
	var (
		ctx        = dbClient.Ctx
		collection = dbClient.Collection
	)

	cursor, err := collection.Find(*ctx, nil)
	if err != nil {
		return nil, wrapGetConfig(err)
	}

	var gyms []utils.GymWithServer
	defer func() {
		if err := cursor.Close(*ctx); err != nil {
			log.Error(err)
		}
	}()

	if err := cursor.All(*ctx, &gyms); err != nil {
		return nil, wrapGetConfig(err)
	}

	if len(gyms) == 0 {
		return nil, wrapGetConfig(errors.New("no gyms found"))
	}

	return gyms, nil
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
