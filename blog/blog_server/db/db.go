package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
)

func New() (*mongo.Client, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Printf("Could not connect to mongo: %v", err)
	}
	err = client.Ping(ctx, &readpref.ReadPref{})
	if err != nil {
		log.Fatalf("Could not ping mongodb: %v", err)
	}

	closeMongo := func() {
		log.Println("Closing mongodb connection...")
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}
	return client, closeMongo
}
