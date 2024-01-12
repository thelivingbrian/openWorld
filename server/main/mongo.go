package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User represents the structure for the MongoDB document
type User struct {
	Email       string    `bson:"email"`
	Verified    bool      `bson:"verified"`
	Username    string    `bson:"username"`
	Hashword    string    `bson:"hashword"`
	CSSClass    string    `bson:"cssClass,omitempty"`
	Created     time.Time `bson:"created,omitempty"`
	LastLogin   time.Time `bson:"lastLogin,omitempty"`
	LastLogout  time.Time `bson:"lastLogout,omitempty"`
	LastRespawn time.Time `bson:"lastRespawn,omitempty"`
	Health      int       `bson:"health,omitempty"`
	StageName   string    `bson:"stagename,omitempty"`
	X           int       `bson:"x,omitempty"`
	Y           int       `bson:"y,omitempty"`
	Kills       []int     `bson:"kills,omitempty"`  // []string for guid? or []primitive.ObjectId
	Deaths      []int     `bson:"deaths,omitempty"` // . . .
	Experience  int       `bson:"experience,omitempty"`
	Records     []int     `bson:"records,omitempty"` // . . .
	Money       int       `bson:"money,omitempty"`
	Inventory   []int     `bson:"inventory,omitempty"` // . .
	Bank        []int     `bson:"bank,omitempty"`
}

func mongoClient() *mongo.Client {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	return client
}

func addUser(collection *mongo.Collection, person User) {
	//collection := client.Database("bloopdb").Collection("users")

	// Insert a single document
	_, err := collection.InsertOne(context.TODO(), person)
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println("Inserted a single document: ", insertResult.InsertedID)
}

func createIndex(client *mongo.Client) { // Move to tools?
	collection := client.Database("bloopdb").Collection("users")

	// Define a unique index on the email field
	indexModel := mongo.IndexModel{
		Keys: bson.M{
			"email": 1, // index in ascending order
		},
		Options: options.Index().SetUnique(true),
	}

	// Create the index
	indexName, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Index %s created\n", indexName)
}

func incrementUserCoordinates(collection *mongo.Collection, userEmail string) (*User, error) {
	ctx := context.Background()

	// Filter to match the user by email
	filter := bson.M{"email": userEmail}

	// Update to increment X and Y
	update := bson.M{
		"$inc": bson.M{"x": 1, "y": 1},
	}

	// Return the updated document
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedUser User
	err := collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedUser)
	if err != nil {
		fmt.Println("oops")
		return nil, err
	}

	return &updatedUser, nil
}

func setUserHealth(collection *mongo.Collection, userEmail string) (*User, error) {
	ctx := context.Background()

	// Filter to match the user by email
	filter := bson.M{"email": userEmail}

	// Update to increment X and Y
	update := bson.M{
		"$set": bson.M{"health": 110},
	}

	// Return the updated document
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedUser User
	err := collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedUser)
	if err != nil {
		return nil, err
	}

	return &updatedUser, nil
}
