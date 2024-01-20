package main

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var testdb *DB = func() *DB {
	testClient := mongoClient(getConfiguration()) // Make test config
	return &DB{
		users:         testClient.Database("bloop-TESTdb").Collection("testusers"),
		playerRecords: testClient.Database("bloop-TESTdb").Collection("testplayers"),
	}
}()

const NUMBER_OF_TEST_ACOUNTS = 1000

func setupUsers() []User {
	emailIndex := mongo.IndexModel{
		Keys: bson.M{
			"email": 1,
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := testdb.users.Indexes().CreateOne(context.Background(), emailIndex)
	if err != nil {
		log.Fatal(err)
	}

	usernameIndex := mongo.IndexModel{
		Keys: bson.M{
			"username": 1,
		},
		Options: options.Index().SetUnique(true),
	}
	_, err = testdb.users.Indexes().CreateOne(context.Background(), usernameIndex)
	if err != nil {
		log.Fatal(err)
	}
	testUsers := make([]User, 1000)
	for i := range testUsers {
		iStr := strconv.Itoa(i)
		testUsers[i] = User{
			Email:    iStr + "@example.com",
			Verified: true,
			Username: "testuser" + iStr,
			Hashword: "hashedpassword",
			Created:  time.Now(),
		}
		testdb.newAccount(testUsers[i])
	}

	return testUsers
}

func cleanUp() {
	// A filter that matches all documents
	//filter := bson.D{{}}
	//res, err := collection.DeleteMany(context.Background(), filter)

	testdb.users.Drop(context.Background())
	testdb.playerRecords.Drop(context.Background())
}

func BenchmarkMongoInsert(b *testing.B) {
	defer cleanUp()
	numberToInsert := 1
	testUsers := make([]User, numberToInsert)
	for i := range testUsers {
		iStr := strconv.Itoa(i)
		testUsers[i] = User{
			Email:    iStr + "insertTest@example.com",
			Verified: true,
			Username: "testInsert" + iStr,
			Hashword: "hashedpassword",
		}
	}

	// Test
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for x := 0; x < numberToInsert; x++ {
			testdb.insertUser(testUsers[x])
		}
	}
	b.StopTimer()
}

func BenchmarkMongoUpdate(b *testing.B) {
	testUsers := setupUsers()
	defer cleanUp()
	testUpdate := make(map[string]any)
	testUpdate["X"] = 7
	testUpdate["Y"] = 11
	testUpdate["stageName"] = "testStageNameHere"

	// Test
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		randomNumber := rand.Intn(1000)
		testdb.updatePlayerRecord(testUsers[randomNumber].Username, testUpdate)

	}
	b.StopTimer()
}
