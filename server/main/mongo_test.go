package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupUsers() ([]User, *mongo.Collection) {
	client := mongoClient()
	collection := client.Database("bloopdb").Collection("test-users")
	emailIndex := mongo.IndexModel{
		Keys: bson.M{
			"email": 1,
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(context.Background(), emailIndex)
	if err != nil {
		log.Fatal(err)
	}

	usernameIndex := mongo.IndexModel{
		Keys: bson.M{
			"username": 1,
		},
		Options: options.Index().SetUnique(true),
	}
	_, err = collection.Indexes().CreateOne(context.Background(), usernameIndex)
	if err != nil {
		log.Fatal(err)
	}
	testUsers := make([]User, 10000)
	for i := range testUsers {
		iStr := strconv.Itoa(i)
		testUsers[i] = User{
			Email:    iStr + "@example.com",
			Verified: true,
			Username: "testuser" + iStr,
			Hashword: "hashedpassword",
			Created:  time.Now(),
		}
		insertUser(collection, testUsers[i])
	}

	return testUsers, collection
}

func cleanUsers(col *mongo.Collection) {
	// A filter that matches all documents
	//filter := bson.D{{}}
	//res, err := collection.DeleteMany(context.Background(), filter)

	col.Drop(context.Background())
}

func BenchmarkMongoInsert(b *testing.B) {
	testUsers := make([]User, 10000)
	for i := range testUsers {
		iStr := strconv.Itoa(i)
		testUsers[i] = User{
			Email:    iStr + "@example.com",
			Verified: true,
			Username: "testuser" + iStr,
			Hashword: "hashedpassword",
		}
	}
	client := mongoClient()
	collection := client.Database("bloopdb").Collection("testusers")
	b.ResetTimer()
	// Test
	for i := 0; i < b.N; i++ {
		for x := 0; x < 1; x++ {
			insertUser(collection, testUsers[x])
		}
	}

	collection.Drop(context.Background())
}
func BenchmarkMongoUpdate(b *testing.B) {
	testUsers, collection := setupUsers()
	defer cleanUsers(collection) // Use defer to ensure cleanup happens after the benchmark.

	fmt.Println("Added Users.")
	b.ResetTimer()

	// Test
	for i := 0; i < b.N; i++ {
		randomNumber := rand.Intn(10000)
		setUserHealth(collection, testUsers[randomNumber].Email) // User shouldn't have health
	}

	b.StopTimer()
}
