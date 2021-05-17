package location

import (
	"context"
	"errors"
	"net/url"
	"os"
	"time"

	originalHTTP "net/http"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/clients"
	databaseUtils "github.com/NOVAPokemon/utils/database"
	"github.com/NOVAPokemon/utils/websockets"
	http "github.com/bruno-anjos/archimedesHTTPClient"
	cedUtils "github.com/bruno-anjos/cloud-edge-deployment/pkg/utils"
	"github.com/golang/geo/s2"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	databaseName             = "NOVAPokemonDB"
	gymsConfigCollectionName = "GymConfigs"
)

var dbClient databaseUtils.DBClient

func InitGymDBClient(archimedesEnabled bool) {
	mongoUrl, exists := os.LookupEnv(utils.MongoEnvVar)
	if !exists {
		mongoUrl = databaseUtils.DefaultMongoDBUrl
	}

	if archimedesEnabled {
		log.Info("archimedes enabled")

		urlParsed, err := url.Parse(mongoUrl)
		if err != nil {
			panic(err)
		}

		var location string
		location, exists = os.LookupEnv("LOCATION")
		if !exists {
			log.Fatal("no location in environment")
		}

		var node string
		node, exists = os.LookupEnv(cedUtils.NodeIPEnvVarName)
		if !exists {
			log.Panicf("no NODE_IP env var")
		} else {
			log.Infof("Node IP: %s", node)
		}

		client := &http.Client{
			Client: originalHTTP.Client{
				Timeout:   websockets.Timeout,
				Transport: clients.NewTransport(),
			},
		}

		client.InitArchimedesClient(node, http.DefaultArchimedesPort, s2.CellIDFromToken(location).LatLng())

		log.Info("initialized archimedes client")

		var (
			resolvedHostPort string
			found            bool
		)

		for {
			log.Info("will try to resolve")
			resolvedHostPort, found, err = client.ResolveServiceInArchimedes(urlParsed.Host)
			if err != nil {
				panic(err)
			}

			if found {
				log.Info("found service")
				break
			}

			time.Sleep(2 * time.Second)
		}

		mongoUrl = "mongodb://" + resolvedHostPort

		log.Infof("resolved %s to %s", urlParsed, mongoUrl)
	} else {
		log.Info("archimedes disabled")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoUrl))
	if err != nil {
		log.Fatal(err)
	}

	log.Info("created mongo client")

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
	ctx := dbClient.Ctx
	collection := dbClient.Collection
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
