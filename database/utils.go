package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

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
