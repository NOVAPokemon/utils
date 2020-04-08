package transactions

import (
	"context"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/database"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

const defaultMongoDBUrl = "mongodb://localhost:27017"
const databaseName = "NOVAPokemonDB"
const collectionName = "Transactions"

var dbClient database.DBClient

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
	dbClient = database.DBClient{Client: client, Ctx: &ctx, Collection: collection}
}

func GetTransactionsFromUser(username string) ([]utils.TransactionRecord, error) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	var result []utils.TransactionRecord

	filter := bson.M{"user": username}
	cursor, err := collection.Find(*ctx, filter)

	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer cursor.Close(*ctx)
	err = cursor.All(*ctx, &result)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	return result, nil
}

func AddTransaction(transaction utils.TransactionRecord) (*primitive.ObjectID, error) {

	transaction.Id = primitive.NewObjectID()

	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	_, err := collection.InsertOne(*ctx, transaction)

	if err != nil {
		return nil, err
	} else {
		log.Infof("Added new transaction: %s", transaction.Id.Hex())
		return &transaction.Id, nil
	}
}

func RemoveAllTransactions() error {
	ctx := dbClient.Ctx
	collection := dbClient.Collection
	_, err := collection.DeleteMany(*ctx, bson.M{})

	if err != nil {
		log.Error(err)
	}

	return err
}
