package location

import (
	"context"
	"net/url"
	"os"
	"time"

	"github.com/NOVAPokemon/utils"
	databaseUtils "github.com/NOVAPokemon/utils/database"
	http "github.com/bruno-anjos/archimedesHTTPClient"
	cedUtils "github.com/bruno-anjos/cloud-edge-deployment/pkg/utils"
	"github.com/golang/geo/s2"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	originalHTTP "net/http"
	"github.com/NOVAPokemon/utils/clients"
)

const (
	databaseName                = "NOVAPokemonDB"
	usersLocationCollectionName = "UsersLocation"
	gymsLocationCollectionName  = "GymsLocation"
	globalConfigCollectionName  = "RegionConfigs"
	wildPokemonCollectionName   = "WildPokemons"
)

var dbClient databaseUtils.DBClientMultipleCollections

func InitLocationDBClient(archimedesEnabled bool) {
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

		client := &http.Client{Client: originalHTTP.Client{Timeout: clients.RequestTimeout}}
		client.InitArchimedesClient(node, http.DefaultArchimedesPort, s2.CellIDFromToken(location).LatLng())

		var (
			resolvedHostPort string
			found            bool
		)

		for {
			resolvedHostPort, found, err = client.ResolveServiceInArchimedes(urlParsed.Host)
			if err != nil {
				panic(err)
			}

			if found {
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

	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	usersLocationCollection := client.Database(databaseName).Collection(usersLocationCollectionName)
	gymsLocationCollection := client.Database(databaseName).Collection(gymsLocationCollectionName)
	wildPokemonsCollection := client.Database(databaseName).Collection(wildPokemonCollectionName)
	globalConfigCollection := client.Database(databaseName).Collection(globalConfigCollectionName)
	collections := map[string]*mongo.Collection{
		usersLocationCollectionName: usersLocationCollection,
		gymsLocationCollectionName:  gymsLocationCollection,
		wildPokemonCollectionName:   wildPokemonsCollection,
		globalConfigCollectionName:  globalConfigCollection,
	}
	dbClient = databaseUtils.DBClientMultipleCollections{Client: client, Ctx: &ctx, Collections: collections}
}

func UpdateIfAbsentAddGym(gymWithSrv utils.GymWithServer) error {
	gym := gymWithSrv.Gym
	ctx := dbClient.Ctx
	collection := dbClient.Collections[gymsLocationCollectionName]

	filter := bson.M{"gym.name": gymWithSrv.Gym.Name}
	upsert := true
	updateOptions := &options.ReplaceOptions{
		Upsert: &upsert,
	}

	_, err := collection.ReplaceOne(*ctx, filter, gymWithSrv, updateOptions)
	if err != nil {
		return wrapAddGymError(err)
	}

	log.Infof("Added gym %s at %f %f", gym.Name, gym.Location.Lat, gym.Location.Lng)

	return nil
}

func GetGyms() ([]utils.GymWithServer, error) {
	ctx := dbClient.Ctx
	collection := dbClient.Collections[gymsLocationCollectionName]

	cur, err := collection.Find(*ctx, bson.M{})
	if err != nil {
		return nil, wrapGetGymsError(err)
	}

	var gymsWithSrv []utils.GymWithServer
	err = cur.All(*ctx, &gymsWithSrv)
	if err != nil {
		return nil, wrapGetGymsError(err)
	}

	return gymsWithSrv, nil
}

func UpdateServerConfig(serverName string, config utils.LocationServerCells) error {
	ctx := dbClient.Ctx
	collection := dbClient.Collections[globalConfigCollectionName]
	filter := bson.D{{"servername:", serverName}}
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

func GetServerConfig(serverName string) (*utils.LocationServerCells, error) {
	var (
		ctx        = dbClient.Ctx
		collection = dbClient.Collections[globalConfigCollectionName]
		filter     = bson.D{{"servername", serverName}}
	)

	res := collection.FindOne(*ctx, filter)

	if res.Err() != nil {
		return nil, wrapGetServerConfig(res.Err(), serverName)
	}

	regionConfig := &utils.LocationServerCells{}
	if err := res.Decode(regionConfig); err != nil {
		return nil, wrapGetServerConfig(res.Err(), serverName)
	}
	return regionConfig, nil
}

func GetAllServerConfigs() (map[string]utils.LocationServerCells, error) {
	ctx := dbClient.Ctx
	collection := dbClient.Collections[globalConfigCollectionName]
	filter := bson.M{}

	cursor, err := collection.Find(*ctx, filter)
	if err != nil {
		return nil, wrapGetGlobalServerConfigs(err)
	}

	out := make(map[string]utils.LocationServerCells, 0)
	for cursor.Next(*ctx) {
		var serverCells utils.LocationServerCells
		if err = cursor.Decode(&serverCells); err != nil {
			log.Fatal(err)
		}
		out[serverCells.ServerName] = serverCells
	}
	return out, nil
}
