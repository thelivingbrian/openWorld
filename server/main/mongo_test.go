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

// Does this have a performance impace in prod?

var testClient *mongo.Client
var testDB *DB

func testdb() *DB {
	if testClient == nil {
		configuration := Configuration{
			envName:   "UNITTEST",
			usesTLS:   false,
			mongoHost: "localhost",
			mongoPort: ":27017",
			mongoUser: "",
			mongoPass: "",
		}
		testClient = mongoClient(&configuration) // Make test config
		testDB = &DB{
			users:         testClient.Database("bloop-TESTdb").Collection("testusers"),
			playerRecords: testClient.Database("bloop-TESTdb").Collection("testplayers"),
			events:        testClient.Database("bloop-TESTdb").Collection("testevents"),
		}
	}
	return testDB
}

const NUMBER_OF_TEST_ACOUNTS = 1000

func setupUsers() []User {
	emailIndex := mongo.IndexModel{
		Keys: bson.M{
			"email": 1,
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := testdb().users.Indexes().CreateOne(context.Background(), emailIndex)
	if err != nil {
		log.Fatal(err)
	}

	usernameIndex := mongo.IndexModel{
		Keys: bson.M{
			"username": 1,
		},
		Options: options.Index().SetUnique(true),
	}
	_, err = testdb().users.Indexes().CreateOne(context.Background(), usernameIndex)
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
		testdb().newAccount(testUsers[i])
	}

	return testUsers
}

func cleanUp() {
	// A filter that matches all documents
	//filter := bson.D{{}}
	//res, err := collection.DeleteMany(context.Background(), filter)

	testdb().users.Drop(context.Background())
	testdb().playerRecords.Drop(context.Background())
	// should drop event records here too but curious to see if and when/how this becomes a problem
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
			testdb().insertUser(testUsers[x])
		}
	}
	b.StopTimer()
}

// Rewrite to use real func
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
		testdb().updatePlayerRecord(testUsers[randomNumber].Username, testUpdate)

	}
	b.StopTimer()
}

// This is only being used by a test
func (db *DB) updatePlayerRecord(username string, updates map[string]any) (*PlayerRecord, error) {
	collection := db.playerRecords

	filter := bson.M{"username": bson.M{"$eq": username}}
	updateBson := bson.M{}
	for key, value := range updates {
		updateBson[key] = value
	}
	setBson := bson.M{
		"$set": updateBson,
	}
	ctx := context.Background()
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After) // Testing shows very small differnce 12ms vs 12.5ms differnce returning vs not

	var updatedRecord PlayerRecord
	err := collection.FindOneAndUpdate(ctx, filter, setBson, opts).Decode(&updatedRecord)
	if err != nil {
		return nil, err
	}

	return &updatedRecord, nil
}
