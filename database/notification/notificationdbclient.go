package notification

import (
	"context"
	"errors"
	"github.com/NOVAPokemon/utils"
	databaseUtils "github.com/NOVAPokemon/utils/database"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

const defaultMongoDBUrl = "mongodb://localhost:27017"
const databaseName = "NOVAPokemonDB"
const collectionName = "Notifications"

var dbClient databaseUtils.DBClient

func GetNotificationsByUsername(username string) (*utils.Notification, error) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	var result utils.Notification

	filter := bson.M{"username": username}
	err := collection.FindOne(*ctx, filter).Decode(&result)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &result, nil
}

func AddNotification(notification utils.Notification) error {
	ctx := dbClient.Ctx
	collection := dbClient.Collection

	notificationId := primitive.NewObjectID()
	notification.Id = notificationId

	_, err := collection.InsertOne(*ctx, notification)

	if err != nil {
		log.Error(err)
		return err
	}

	log.Infof("Added notification %s to user: %s", notification.Id, notification.Username)
	return err
}

func RemoveNotification(id primitive.ObjectID) error {
	ctx := dbClient.Ctx
	collection := dbClient.Collection

	res, err := collection.DeleteOne(*ctx, id)

	if err != nil {
		log.Error(err)
		return err
	}

	if res.DeletedCount < 1 {
		log.Error("tried to remove notification that didnt exist")
		return errors.New("")
	}

	log.Infof("Removed notification %s", id)
	return nil
}

func init() {

	url, exists := os.LookupEnv("MONGODB_URL")

	if !exists {
		url = defaultMongoDBUrl
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