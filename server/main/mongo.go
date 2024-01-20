package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func mongoClient(config *Configuration) *mongo.Client {
	clientOptions := options.Client().ApplyURI(config.getMongoURI())

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func createIndexes(client *mongo.Client) { // Move to tools?
	collection := client.Database("bloopdb").Collection("users")

	emailIndex := mongo.IndexModel{
		Keys: bson.M{
			"email": 1, // index in ascending order
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(context.Background(), emailIndex)
	if err != nil {
		log.Fatal(err)
	}

	usernameIndex := mongo.IndexModel{
		Keys: bson.M{
			"username": 1, // index in ascending order
		},
		Options: options.Index().SetUnique(true),
	}
	_, err = collection.Indexes().CreateOne(context.Background(), usernameIndex)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("bloopdb").Collection("players")
	_, err = collection.Indexes().CreateOne(context.Background(), usernameIndex)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Indexes created\n")
}
