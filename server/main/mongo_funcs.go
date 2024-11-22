package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"
)

type User struct {
	Email    string    `bson:"email"`
	Verified bool      `bson:"verified"`
	Username string    `bson:"username"`
	Hashword string    `bson:"hashword"`
	Created  time.Time `bson:"created,omitempty"`
}

type AuthorizedUser struct {
	Identifier string    `bson:"identifier"`
	Username   string    `bson:"username"`
	Created    time.Time `bson:"created,omitempty"`
	LastLogin  time.Time `bson:"lastLogin,omitempty"`
}

type PlayerRecord struct {
	Username    string    `bson:"username"`
	Team        string    `bson:"team"`
	Trim        string    `bson:"trim"`
	LastLogin   time.Time `bson:"lastLogin,omitempty"`
	LastLogout  time.Time `bson:"lastLogout,omitempty"`
	LastRespawn time.Time `bson:"lastRespawn,omitempty"`
	CSSClass    string    `bson:"cssClass,omitempty"`
	Health      int       `bson:"health,omitempty"`
	StageName   string    `bson:"stagename,omitempty"`
	X           int       `bson:"x,omitempty"`
	Y           int       `bson:"y,omitempty"`
	Kills       []string  `bson:"kills,omitempty"` // This might make loading a user expensive consider a ref table
	Deaths      []string  `bson:"deaths,omitempty"`
	Experience  int       `bson:"experience,omitempty"`
	Records     []string  `bson:"records,omitempty"`
	Money       int       `bson:"money,omitempty"`
	Inventory   []int     `bson:"inventory,omitempty"` // What is ID of an Item? string?
	Bank        []int     `bson:"bank,omitempty"`
}

type Event struct {
	ID        string    `bson:"eventid"`
	Owner     string    `bson:"owner"`
	Secondary string    `bson:"secondary"`
	Type      string    `bson:"eventtype"`
	Created   time.Time `bson:"created"`
	StageName string    `bson:"stagename,omitempty"`
	X         int       `bson:"x,omitempty"`
	Y         int       `bson:"y,omitempty"`
	Details   string    `bson:"details,omitempty"`
}

func (db *DB) newAccount(user User) error {
	player := PlayerRecord{
		Username:  user.Username,
		Health:    100,
		StageName: "tutorial:0-0",
		X:         4,
		Y:         4,
		Money:     80,
	}
	err := db.insertUser(user)
	if err != nil {
		return err // This is fine
	}
	err = db.InsertPlayerRecord(player)
	if err != nil {
		return err // This is not fun
	}
	return nil
}

func (db *DB) insertUser(user User) error {
	_, err := db.users.InsertOne(context.TODO(), user)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) InsertPlayerRecord(player PlayerRecord) error {
	_, err := db.playerRecords.InsertOne(context.TODO(), player)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) getUserByEmail(email string) (*User, error) {
	var result User
	collection := db.users
	err := collection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No document was found with the given email")
			return nil, err
		} else {
			log.Fatal(err)
		}
	}
	return &result, nil
}

/////////////////////////////////////////////////////////////
//  Authorized

func (db *DB) getAuthorizedUserById(identifier string) *AuthorizedUser {
	var result AuthorizedUser
	collection := db.users
	err := collection.FindOne(context.TODO(), bson.M{"identifier": identifier}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No document was found with the given identifier")
			return nil
		} else {
			log.Fatal(err)
		}
	}
	return &result
}

func (db *DB) insertAuthorizedUser(user AuthorizedUser) error {
	_, err := db.users.InsertOne(context.TODO(), user)
	return err
}

func (db *DB) updateUserName(identifier, username string) bool {
	filter := bson.M{"identifier": identifier, "username": ""}
	update := bson.M{"$set": bson.M{"username": username}}

	result, err := db.users.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Println("Error updating document:", err)
		return false
	}

	if result.MatchedCount == 0 {
		fmt.Println("No document matched the identifier with an empty username.")
		return false
	}

	if result.ModifiedCount == 0 {
		fmt.Println("Document was matched, but username was not empty.")
		return false
	}

	fmt.Println("Document updated successfully.")
	return true
}

// needs to return error?
func (db *DB) getPlayerRecord(username string) (*PlayerRecord, error) {
	collection := db.playerRecords
	var result PlayerRecord
	err := collection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No document was found with the given email")
			return nil, err
		} else {
			log.Fatal(err)
		}
	}
	return &result, nil
}

func (db *DB) usernameExists(username string) bool {
	record, _ := db.getPlayerRecord(username)
	return record != nil
}

// This is only being used by a test
func (db *DB) updatePlayerRecord(username string, updates map[string]any) (*PlayerRecord, error) {
	collection := db.playerRecords

	filter := bson.M{"username": username}
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

func (db *DB) saveKillEvent(tile *Tile, initiator *Player, defeated *Player) error {
	eventCollection := db.events
	event := Event{
		ID:        uuid.New().String(),
		Owner:     initiator.username,
		Secondary: defeated.username,
		Type:      "Kill",
		Created:   time.Now(),
		X:         tile.x,
		Y:         tile.y,
	}
	_, err := eventCollection.InsertOne(context.TODO(), event)
	if err != nil {
		log.Fatal("Event Insert Failed")
	}

	return nil
}

func (db *DB) updateRecordForPlayer(p *Player) error {
	_, err := db.playerRecords.UpdateOne(
		context.TODO(),
		bson.M{"username": p.username},
		bson.M{
			"$set": bson.M{
				"x":          p.x,
				"y":          p.y,
				"health":     p.health,
				"stagename":  p.stageName, // feels risky
				"money":      p.money,
				"killCount":  p.killCount,
				"deathCount": p.deathCount,
				"trim":       p.trim,
			},
		},
	)
	return err //Is nil or err
}
