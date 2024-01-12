package main

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	Email    string    `bson:"email"`
	Verified bool      `bson:"verified"`
	Username string    `bson:"username"`
	Hashword string    `bson:"hashword"`
	Created  time.Time `bson:"created,omitempty"`
}

type PlayerRecord struct {
	Username    string    `bson:"username"`
	LastLogin   time.Time `bson:"lastLogin,omitempty"`
	LastLogout  time.Time `bson:"lastLogout,omitempty"`
	LastRespawn time.Time `bson:"lastRespawn,omitempty"`
	CSSClass    string    `bson:"cssClass,omitempty"`
	Health      int       `bson:"health,omitempty"`
	StageName   string    `bson:"stagename,omitempty"`
	X           int       `bson:"x,omitempty"`
	Y           int       `bson:"y,omitempty"`
	Kills       []string  `bson:"kills,omitempty"`
	Deaths      []string  `bson:"deaths,omitempty"`
	Experience  int       `bson:"experience,omitempty"`
	Records     []string  `bson:"records,omitempty"`
	Money       int       `bson:"money,omitempty"`
	Inventory   []int     `bson:"inventory,omitempty"` // What is ID of an Item? string?
	Bank        []int     `bson:"bank,omitempty"`
}

func newAccount(db *mongo.Database, user User) error {
	player := PlayerRecord{
		Username:  user.Username,
		Health:    100,
		StageName: "big",
		X:         2,
		Y:         2,
		Money:     80,
	}
	err := insertUser(db.Collection("users"), user)
	if err != nil {
		return err // This is fine
	}
	err = InsertPlayerRecord(db.Collection("players"), player)
	if err != nil {
		return err // This is not fun
	}
	return nil
}

func insertUser(collection *mongo.Collection, user User) error {
	// Insert a single document
	_, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		return err
	}
	return nil
}

func InsertPlayerRecord(collection *mongo.Collection, player PlayerRecord) error {
	_, err := collection.InsertOne(context.TODO(), player)
	if err != nil {
		return err
	}
	return nil
}

func incrementUserCoordinates(collection *mongo.Collection, userEmail string) (*User, error) {
	ctx := context.Background()

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

	// Do Player instead
	filter := bson.M{"email": userEmail}
	update := bson.M{
		"$set": bson.M{"health": 110},
	}

	// Return the updated document // Investigate further
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedUser User
	err := collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedUser)
	if err != nil {
		return nil, err
	}

	return &updatedUser, nil
}
