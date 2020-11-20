package notification

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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const databaseName = "NOVAPokemonDB"
const collectionName = "Notifications"

var dbClient databaseUtils.DBClient

func InitNotificationDBClient(archimedesEnabled bool) {
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

	collection := client.Database(databaseName).Collection(collectionName)

	op := options.Index()
	op.SetUnique(true)
	index := mongo.IndexModel{
		Keys:    bson.M{"_id": 1},
		Options: op,
	}

	_, _ = collection.Indexes().CreateOne(ctx, index)
	dbClient = databaseUtils.DBClient{Client: client, Ctx: &ctx, Collection: collection}
}

func AddNotification(notification utils.Notification) error {
	ctx := dbClient.Ctx
	collection := dbClient.Collection

	notificationId := primitive.NewObjectID()
	notification.Id = notificationId.Hex()

	_, err := collection.InsertOne(*ctx, notification)

	if err != nil {
		return wrapAddNotificationError(err, notification.Username)
	}

	log.Infof("Added notification %s to user: %s", notification.Id, notification.Username)
	return err
}

func RemoveNotification(id primitive.ObjectID) error {
	ctx := dbClient.Ctx
	collection := dbClient.Collection

	res, err := collection.DeleteOne(*ctx, id)

	if err != nil {
		return wrapRemoveNotificationError(err, id.Hex())
	}

	if res.DeletedCount < 1 {
		return wrapRemoveNotificationError(errorNotificationNotFound, id.Hex())
	}

	log.Infof("Removed notification %s", id)
	return nil
}
