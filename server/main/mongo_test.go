package main

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func setupUsers() ([]User, *mongo.Collection) {
	client := mongoClient()
	collection := client.Database("bloopdb").Collection("testusers")
	testUsers := make([]User, 10000)
	for i := range testUsers {
		iStr := strconv.Itoa(i)
		testUsers[i] = User{
			Email:     iStr + "@example.com",
			Verified:  true,
			Username:  "testuser" + iStr,
			Hashword:  "hashedpassword",
			CSSClass:  "exampleClass",
			Created:   time.Now(),
			LastLogin: time.Now(),
			Health:    100,
			StageName: "big",
			X:         2,
			Y:         2,
		}
		addUser(collection, testUsers[i])
	}

	return testUsers, collection
}

func cleanUsers(col *mongo.Collection) {
	col.Drop(context.Background())
}

func BenchmarkMongoInsert(b *testing.B) { // This has weird counterintuitive results, 1 add slower than 10, inconsistent etc.
	testUsers := make([]User, 10000)
	for i := range testUsers {
		iStr := strconv.Itoa(i)
		testUsers[i] = User{
			Email:     iStr + "@example.com",
			Verified:  true,
			Username:  "testuser" + iStr,
			Hashword:  "hashedpassword",
			CSSClass:  "exampleClass",
			Created:   time.Now(),
			LastLogin: time.Now(),
			// Initialize other fields as required
			Health:    100,
			StageName: "big",
			X:         2,
			Y:         2,
		}
	}
	client := mongoClient()
	collection := client.Database("bloopdb").Collection("testusers")
	b.ResetTimer()
	// Test
	for i := 0; i < b.N; i++ {
		for x := 0; x < 1; x++ {
			addUser(collection, testUsers[x])
		}
	}

	// Cleanup
	// A filter that matches all documents
	/*filter := bson.D{{}}

	// Delete all documents
	res, err := collection.DeleteMany(context.Background(), filter)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Deleted %v documents\n", res.DeletedCount)
	*/
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
		setUserHealth(collection, testUsers[randomNumber].Email)
	}

	//cleanUsers(collection)
	b.StopTimer()
}
