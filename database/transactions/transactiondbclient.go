package transactions

import (
	"context"
	"net/url"
	"os"

	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/database"
	http "github.com/bruno-anjos/archimedesHTTPClient"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const databaseName = "NOVAPokemonDB"
const collectionName = "Transactions"

var dbClient database.DBClient

func InitTransactionsDBClient(archimedesEnabled bool) {
	mongoUrl, exists := os.LookupEnv(utils.MongoEnvVar)
	if !exists {
		mongoUrl = database.DefaultMongoDBUrl
	}

	if archimedesEnabled {
		urlParsed, err := url.Parse(mongoUrl)
		if err != nil {
			panic(err)
		}

		client := http.Client{}
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
	dbClient = database.DBClient{Client: client, Ctx: &ctx, Collection: collection}
}

func GetTransactionsFromUser(username string) ([]utils.TransactionRecord, error) {
	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	var result []utils.TransactionRecord

	filter := bson.M{"user": username}

	cursor, err := collection.Find(*ctx, filter)
	if err != nil {
		return nil, wrapGetUserTransactionsError(err, username)
	}

	defer database.CloseCursor(cursor, ctx)
	err = cursor.All(*ctx, &result)
	if err != nil {
		return nil, wrapGetUserTransactionsError(err, username)
	}

	return result, nil
}

func AddTransaction(transaction utils.TransactionRecord) (*string, error) {
	transaction.Id = primitive.NewObjectID().Hex()

	var ctx = dbClient.Ctx
	var collection = dbClient.Collection
	_, err := collection.InsertOne(*ctx, transaction)

	if err != nil {
		return nil, wrapAddTransactionError(err, transaction.User)
	} else {
		log.Infof("Added new transaction: %s", transaction.Id)
		return &transaction.Id, nil
	}
}

func RemoveAllTransactions() error {
	ctx := dbClient.Ctx
	collection := dbClient.Collection
	_, err := collection.DeleteMany(*ctx, bson.M{})

	return wrapRemoveAllTransactionsError(err)
}
