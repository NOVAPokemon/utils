package location

import (
	"context"
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

const (
	databaseName                = "NOVAPokemonDB"
	usersLocationCollectionName = "UsersLocation"
	gymsLocationCollectionName  = "GymsLocation"
	globalConfigCollectionName  = "RegionConfigs"
	wildPokemonCollectionName   = "WildPokemons"
)

var (
	dbClient databaseUtils.DBClientMultipleCollections
)

func InitLocationDBClient(archimedesEnabled bool) {
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
	var ctx = dbClient.Ctx
	var collection = dbClient.Collections[globalConfigCollectionName]
	var filter = bson.D{{"servername:", serverName}}
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
	var ctx = dbClient.Ctx
	var collection = dbClient.Collections[globalConfigCollectionName]
	var filter = bson.D{{"servername", serverName}}

	res := collection.FindOne(*ctx, filter)

	if res.Err() != nil {
		return nil, wrapGetServerConfig(res.Err(), serverName)
	}

	var regionConfig = &utils.LocationServerCells{}
	if err := res.Decode(regionConfig); err != nil {
		return nil, wrapGetServerConfig(res.Err(), serverName)
	}
	return regionConfig, nil
}

func GetAllServerConfigs() (map[string]utils.LocationServerCells, error) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collections[globalConfigCollectionName]
	var filter = bson.M{}

	cursor, err := collection.Find(*ctx, filter)

	if err != nil {
		return nil, wrapGetGlobalServerConfigs(err)
	}

	var out = make(map[string]utils.LocationServerCells, 0)
	for cursor.Next(*ctx) {
		var serverCells utils.LocationServerCells
		if err = cursor.Decode(&serverCells); err != nil {
			log.Fatal(err)
		}
		out[serverCells.ServerName] = serverCells
	}
	return out, nil
}
