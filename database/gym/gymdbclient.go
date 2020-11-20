package location

import (
	"context"
	"errors"
	"net/url"
	"os"

	"github.com/NOVAPokemon/utils"
	databaseUtils "github.com/NOVAPokemon/utils/database"
	http "github.com/bruno-anjos/archimedesHTTPClient"
	"github.com/golang/geo/s2"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const databaseName = "NOVAPokemonDB"
const gymsConfigCollectionName = "GymConfigs"

var dbClient databaseUtils.DBClient

func InitGymDBClient(archimedesEnabled bool) {
	mongoUrl, exists := os.LookupEnv(utils.MongoEnvVar)
	if !exists {
		mongoUrl = databaseUtils.DefaultMongoDBUrl
	}

	if archimedesEnabled {
		urlParsed, err := url.Parse(mongoUrl)
		if err != nil {
			panic(err)
		}

		var location string
		location, exists = os.LookupEnv("LOCATION")
		if !exists {
			log.Fatalf("no location in environment")
		}

		client := &http.Client{}
		client.InitArchimedesClient("localhost", http.DefaultArchimedesPort, s2.CellIDFromToken(location).LatLng())
		resolvedHostPort, err := client.ResolveServiceInArchimedes(urlParsed.Host)
		if err != nil {
			panic(err)
		}

		mongoUrl = "mongodb://" + resolvedHostPort
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoUrl))
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
		if err = cursor.Close(*ctx); err != nil {
			log.Error(err)
		}
	}()
	if err = cursor.All(*ctx, &gymsForServer); err != nil {
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
