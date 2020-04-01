package main

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	PORT    = "8000"
	DB_NAME = "url_shortener"
)

func main() {

	port, found := os.LookupEnv("PORT")

	if !found {
		port = PORT
	}

	db := setupDatabase("localhost:27017")

	sv := NewServer(port, db)

	sv.Run()
}

func setupDatabase(address string) *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://"+address))
	if err != nil {
		panic(err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.Ping(ctx, readpref.Primary())

	if err != nil {
		panic(err)
	}

	return client.Database(DB_NAME)
}
