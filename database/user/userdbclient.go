package user

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
	http "github.com/bruno-anjos/archimedesHTTPClient"
	cedUtils "github.com/bruno-anjos/cloud-edge-deployment/pkg/utils"
	"github.com/golang/geo/s2"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	databaseName   = "NOVAPokemonDB"
	collectionName = "Users"
)

var dbClient databaseUtils.DBClient

func InitUsersDBClient(archimedesEnabled bool) {
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
			log.Panic("no location in environment")
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
				Timeout:   clients.RequestTimeout,
				Transport: clients.NewTransport(),
			},
		}
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

	log.Infof("created client to %s", mongoUrl)

	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("client connected to %s", mongoUrl)

	collection := client.Database(databaseName).Collection(collectionName)

	log.Infof("retrieved collection %s from database %s", collectionName, databaseName)

	op := options.Index()
	op.SetUnique(true)
	index := mongo.IndexModel{
		Keys:    bson.M{"username": 1},
		Options: op,
	}

	_, _ = collection.Indexes().CreateOne(ctx, index)
	dbClient = databaseUtils.DBClient{Client: client, Ctx: &ctx, Collection: collection}

	log.Info("finished database setup")
}

func GetAllUsers() ([]utils.User, error) {
	ctx := dbClient.Ctx
	collection := dbClient.Collection
	var results []utils.User

	cur, err := collection.Find(*ctx, bson.M{})
	if err != nil {
		return nil, wrapGetAllUsersError(err)
	}

	defer databaseUtils.CloseCursor(cur, ctx)
	for cur.Next(*ctx) {
		var result utils.User
		err = cur.Decode(&result)
		if err != nil {
			return nil, wrapGetAllUsersError(err)
		} else {
			results = append(results, result)
		}
	}

	if err = cur.Err(); err != nil {
		return nil, wrapGetAllUsersError(err)
	}
	return results, nil
}

func AddUser(user *utils.User) (string, error) {
	ctx := dbClient.Ctx
	collection := dbClient.Collection
	_, err := collection.InsertOne(*ctx, *user)
	if err != nil {
		return "", wrapAddUserError(err, user.Username)
	}

	return user.Username, nil
}

func CheckIfUserExists(username string) (bool, error) {
	ctx := dbClient.Ctx
	collection := dbClient.Collection
	var limit int64 = 1

	filter := bson.M{"username": username}
	opts := options.CountOptions{Limit: &limit}

	count, err := collection.CountDocuments(*ctx, filter, &opts)
	if err != nil {
		return false, wrapUserExistsError(err, username)
	}

	return count > 0, nil
}

func GetUserByUsername(username string) (*utils.User, error) {
	ctx := dbClient.Ctx
	collection := dbClient.Collection

	result := &utils.User{}
	filter := bson.M{"username": username}

	err := collection.FindOne(*ctx, filter).Decode(result)
	if err != nil {
		return nil, wrapGetUserError(err, username)
	}

	return result, nil
}

func UpdateUser(username string, user *utils.User) (*utils.User, error) {
	ctx := dbClient.Ctx
	collection := dbClient.Collection
	filter := bson.M{"username": username}
	user.Username = username

	res, err := collection.ReplaceOne(*ctx, filter, *user)
	if err != nil {
		return nil, wrapUpdateUserError(err, username)
	}

	if res.MatchedCount > 0 {
		log.Infof("Updated User %+v", username)
	} else {
		return nil, wrapUpdateUserError(errors.New("no user found"), username)
	}

	return user, nil
}

func DeleteUser(username string) error {
	ctx := dbClient.Ctx
	collection := dbClient.Collection
	filter := bson.M{"username": username}

	_, err := collection.DeleteOne(*ctx, filter)

	return wrapDeleteUserError(err, username)
}

func removeAll() error {
	ctx := dbClient.Ctx
	collection := dbClient.Collection
	filter := bson.M{}

	_, err := collection.DeleteMany(*ctx, filter)

	return wrapDeleteAllUsersError(err)
}
