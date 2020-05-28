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
const usersLocationCollectionName = "UsersLocation"
const gymsLocationCollectionName = "GymsLocation"
const globalConfigCollectionName = "RegionConfigs"
const wildPokemonCollectionName = "WildPokemons"

var dbClient databaseUtils.DBClientMultipleCollections

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

	usersLocationCollection := client.Database(databaseName).Collection(usersLocationCollectionName)
	gymsLocationCollection := client.Database(databaseName).Collection(gymsLocationCollectionName)
	wildPokemonsCollection := client.Database(databaseName).Collection(wildPokemonCollectionName)
	globalConfigCollection := client.Database(databaseName).Collection(globalConfigCollectionName)

	op := options.Index()
	op.SetUnique(true)
	index := mongo.IndexModel{
		Keys:    bson.M{"name": 1},
		Options: op,
	}
	_, err = gymsLocationCollection.Indexes().CreateOne(ctx, index)
	if err != nil {
		log.Error(err)
		return
	}

	collections := map[string]*mongo.Collection{
		usersLocationCollectionName: usersLocationCollection,
		gymsLocationCollectionName:  gymsLocationCollection,
		wildPokemonCollectionName:   wildPokemonsCollection,
		globalConfigCollectionName:  globalConfigCollection,
	}

	dbClient = databaseUtils.DBClientMultipleCollections{Client: client, Ctx: &ctx, Collections: collections}
}

func AddGym(gymWithSrv utils.GymWithServer) error {
	gym := gymWithSrv.Gym
	ctx := dbClient.Ctx
	collection := dbClient.Collections[gymsLocationCollectionName]
	_, err := collection.InsertOne(*ctx, gymWithSrv)
	if err != nil {
		return wrapAddGymError(err)
	}

	log.Infof("Added gym %s at %f %f", gym.Name, gym.Location.Latitude, gym.Location.Longitude)

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

func UpdateIfAbsentAddUserLocation(userLocation utils.UserLocation) (*utils.UserLocation, error) {
	ctx := dbClient.Ctx
	collection := dbClient.Collections[usersLocationCollectionName]

	filter := bson.M{"username": userLocation.Username}
	changes := bson.M{"$set": bson.M{"location": userLocation.Location}}

	upsert := true
	updateOptions := options.UpdateOptions{
		Upsert: &upsert,
	}

	res, err := collection.UpdateOne(*ctx, filter, changes, &updateOptions)
	if err != nil {
		return nil, wrapUpdateLocation(err, userLocation.Username)
	}

	if res.MatchedCount > 0 {
		log.Infof("Updated Trainer %s location", userLocation.Username)
	} else {
		log.Info("Started tracking ", userLocation.Username)
	}

	return &userLocation, nil
}

func GetUserLocation(username string) (*utils.UserLocation, error) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collections[usersLocationCollectionName]
	var result utils.UserLocation

	filter := bson.M{"username": username}
	err := collection.FindOne(*ctx, filter).Decode(&result)
	if err != nil {
		return nil, wrapGetLocation(err, username)
	}

	return &result, nil
}

func DeleteUserLocation(username string) error {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collections[usersLocationCollectionName]
	filter := bson.M{"username": username}

	_, err := collection.DeleteOne(*ctx, filter)

	return wrapDeleteLocation(err, username)
}

func UpdateServerConfig(serverName string, config utils.LocationServerBoundary) error {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collections[globalConfigCollectionName]
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

func GetServerConfig(serverName string) (*utils.LocationServerBoundary, error) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collections[globalConfigCollectionName]
	var filter = bson.D{{"servername", serverName}}

	res := collection.FindOne(*ctx, filter)

	if res.Err() != nil {
		return nil, wrapGetServerConfig(res.Err(), serverName)
	}

	var regionConfig = &utils.LocationServerBoundary{}
	if err := res.Decode(regionConfig); err != nil {
		return nil, wrapGetServerConfig(res.Err(), serverName)
	}
	return regionConfig, nil
}

func GetAllServerConfigs() (map[string]utils.LocationServerBoundary, error) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collections[globalConfigCollectionName]
	var filter = bson.M{}

	cursor, err := collection.Find(*ctx, filter)

	if err != nil {
		return nil, wrapGetGlobalServerConfigs(err)
	}

	var out = make(map[string]utils.LocationServerBoundary, 0)
	for cursor.Next(*ctx) {
		var regionCfg utils.LocationServerBoundary
		if err = cursor.Decode(&regionCfg); err != nil {
			log.Fatal(err)
		}
		out[regionCfg.ServerName] = regionCfg
	}
	return out, nil
}
