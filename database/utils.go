package database

import (
	"context"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

const DefaultMongoDBUrl = "mongodb://localhost:27017"

type DBClient struct {
	Client     *mongo.Client
	Collection *mongo.Collection
	Ctx        *context.Context
}

type DBClientMultipleCollections struct {
	Client      *mongo.Client
	Collections map[string]*mongo.Collection
	Ctx         *context.Context
}

func CloseCursor(cursor *mongo.Cursor, ctx *context.Context) {
	if cursor == nil {
		return
	}
	if err := cursor.Close(*ctx); err != nil {
		log.Error(err)
	}
}
