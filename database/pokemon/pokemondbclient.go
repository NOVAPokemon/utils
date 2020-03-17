package pokemon

import (
	"context"
	"github.com/NOVAPokemon/utils"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

const defaultMongoDBUrl = "mongodb://localhost:27017"
const databaseName = "NOVAPokemonDB"
const pokemonCollectionName = "Pokemons"
const wildPokemonCollectionName = "WildPokemons"

type DBCLient struct {
	client      *mongo.Client
	collections map[string]*mongo.Collection
	ctx         *context.Context
}

var dbClient DBCLient

func GetAllPokemons() []utils.Pokemon {
	var ctx = dbClient.ctx
	var collection = dbClient.collections[pokemonCollectionName]
	var results []utils.Pokemon

	cur, err := collection.Find(*ctx, bson.M{})

	if err != nil {
		log.Println(err)
	}

	defer cur.Close(*ctx)
	for cur.Next(*ctx) {
		var result utils.Pokemon
		err := cur.Decode(&result)
		if err != nil {
			log.Error(err)
		} else {
			results = append(results, result)
		}
	}

	if err := cur.Err(); err != nil {
		log.Error(err)
	}

	return results

}

func GetPokemonById(id primitive.ObjectID) (error, utils.Pokemon) {
	var ctx = dbClient.ctx
	var collection = dbClient.collections[pokemonCollectionName]
	var result utils.Pokemon

	filter := bson.M{"_id": id}
	err := collection.FindOne(*ctx, filter).Decode(&result)

	if err != nil {
		log.Error(err)
	}

	return err, result
}

func AddPokemonToUser(pokemon utils.Pokemon) (error, primitive.ObjectID) {
	var ctx = dbClient.ctx
	var collection = dbClient.collections[pokemonCollectionName]
	res, err := collection.InsertOne(*ctx, pokemon)

	if err != nil {
		log.Error(err)
		return nil, [12]byte{}
	}

	log.Infof("Inserted new Pokemon %s", res.InsertedID)

	return err, res.InsertedID.(primitive.ObjectID)
}

func AddWildPokemon(pokemon utils.Pokemon) (error, primitive.ObjectID) {
	var ctx = dbClient.ctx
	var collection = dbClient.collections[wildPokemonCollectionName]
	res, err := collection.InsertOne(*ctx, pokemon)

	if err != nil {
		log.Error(err)
		return nil, [12]byte{}
	}

	log.Infof("Inserted new wild Pokemon %s", res.InsertedID)

	return err, res.InsertedID.(primitive.ObjectID)
}

func DeleteWildPokemons() error {
	var ctx = dbClient.ctx
	var collection = dbClient.collections[wildPokemonCollectionName]
	filter := bson.M{}

	_, err := collection.DeleteOne(*ctx, filter)

	if err != nil {
		log.Error(err)
	}

	return err
}

func UpdatePokemon(id primitive.ObjectID, pokemon utils.Pokemon) (error, utils.Pokemon) {
	ctx := dbClient.ctx
	collection := dbClient.collections[pokemonCollectionName]
	filter := bson.M{"_id": id}
	pokemon.Id = id

	res, err := collection.ReplaceOne(*ctx, filter, pokemon)

	if err != nil {
		log.Error(err)
	}

	if res.MatchedCount > 0 {
		log.Infof("Updated Pokemon %s", id)
	} else {
		log.Errorf("Update Pokemon failed because no pokemon matched %s", id)
	}

	return err, pokemon

}

func DeletePokemon(id primitive.ObjectID) error {
	var ctx = dbClient.ctx
	var collection = dbClient.collections[pokemonCollectionName]
	filter := bson.M{"_id": id}

	_, err := collection.DeleteOne(*ctx, filter)

	if err != nil {
		log.Error(err)
	}

	return err
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

	userPokemons := client.Database(databaseName).Collection(pokemonCollectionName)
	wildPokemons := client.Database(databaseName).Collection(wildPokemonCollectionName)
	collections := map[string]*mongo.Collection{
		pokemonCollectionName:     userPokemons,
		wildPokemonCollectionName: wildPokemons,
	}
	dbClient = DBCLient{client: client, ctx: &ctx, collections: collections}
}
