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
		StageName: "big",
		X:         2,
		Y:         2,
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

// This is onnly being used by a test
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
	playerCollection := db.playerRecords
	eventCollection := db.events

	id := uuid.New().String()
	event := Event{
		ID:        id,
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

	// Update the initiator player's kills array
	_, err = playerCollection.UpdateOne(
		context.TODO(),
		bson.M{"username": initiator.username},
		bson.M{
			"$push": bson.M{"kills": id},
			"$set": bson.M{
				"x":         tile.x,
				"y":         tile.y,
				"health":    initiator.health,
				"stagename": initiator.stageName,
				"money":     initiator.money,
			},
			"$inc": bson.M{
				"experience": 100,
			},
		},
	)
	if err != nil {
		log.Fatal("Update Initiator Kills Failed")
		return err
	}

	// Update the defeated player's deaths array
	_, err = playerCollection.UpdateOne(
		context.TODO(),
		bson.M{"username": defeated.username},
		bson.M{"$push": bson.M{"deaths": id}},
	)
	if err != nil {
		log.Fatal("Update Defeated Deaths Failed")
		return err
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
				"stagename":  p.stageName,
				"money":      p.money,
				"experience": p.experience,
			},
		},
	)
	return err //Is nil or err
}
