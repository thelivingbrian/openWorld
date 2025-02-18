package main

import (
	"context"
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
	Identifier    string    `bson:"identifier"`
	Username      string    `bson:"username"`
	CreationEmail string    `bson:"creationEmail"`
	Created       time.Time `bson:"created,omitempty"`
	LastLogin     time.Time `bson:"lastLogin,omitempty"`
}

type PlayerRecord struct {
	// ID
	Username string `bson:"username"`
	// Meta
	LastLogin   time.Time `bson:"lastLogin,omitempty"`
	LastLogout  time.Time `bson:"lastLogout,omitempty"`
	LastRespawn time.Time `bson:"lastRespawn,omitempty"`
	// World Location
	StageName string `bson:"stagename"`
	X         int    `bson:"x"`
	Y         int    `bson:"y"`
	// Stats
	Team        string `bson:"team"`
	Trim        string `bson:"trim,omitempty"`
	Health      int    `bson:"health"`
	Money       int    `bson:"money,omitempty"`
	KillCount   int    `bson:"killCount,omitempty"`
	DeathCount  int    `bson:"deathCount,omitempty"`
	GoalsScored int    `bson:"goalsScored,omitempty"`
	// Unlocks
	HatList HatList `bson:"hatList,omitempty"`
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
			logger.Error().Err(err).Msg("No document was found with the given email") // logger.Error().Err(err).Msg(
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
			logger.Error().Err(err).Msg("No document was found with the given identifier")
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

func (db *DB) updateUsernameForUserWithId(identifier, username string) bool {
	filter := bson.M{"identifier": identifier, "username": ""}
	update := bson.M{"$set": bson.M{"username": username}}

	result, err := db.users.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		logger.Error().Err(err).Msg("Error updating document:")
		return false
	}

	if result.MatchedCount == 0 {
		logger.Error().Msg("No document matched the identifier with an empty username.")
		return false
	}

	if result.ModifiedCount == 0 {
		logger.Error().Msg("Document was matched, but username was not empty.")
		return false
	}

	logger.Info().Msg("Document updated successfully.")
	return true
}

// needs to return error?
func (db *DB) getPlayerRecord(username string) (PlayerRecord, error) {
	collection := db.playerRecords
	var result PlayerRecord
	err := collection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Error().Err(err).Msg("No document was found with the given email")
			return PlayerRecord{Username: "invalild"}, err
		} else {
			log.Fatal(err)
		}
	}
	return result, nil
}

func (db *DB) foundUsername(username string) bool {
	_, err := db.getPlayerRecord(username)
	return err == nil
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

func (db *DB) updateRecordForPlayer(p *Player, pTile *Tile) error {
	_, err := db.playerRecords.UpdateOne(
		context.TODO(),
		bson.M{"username": p.username},
		bson.M{
			"$set": bson.M{
				"x":               pTile.x, // All of this feels dangerous tbh
				"y":               pTile.y,
				"health":          p.getHealthSync(),
				"stagename":       pTile.stage.name, //p.getStageNameSync(), // feels risky
				"money":           p.getMoneySync(),
				"killCount":       p.getKillCountSync(),
				"deathCount":      p.getDeathCountSync(),
				"goalsScored":     p.getGoalsScored(),
				"hatList.current": p.hatList.indexSync(),
			},
		},
	)
	return err //Is nil or err
}

func (db *DB) addHatToPlayer(username string, newHat Hat) error {
	_, err := db.playerRecords.UpdateOne(
		context.TODO(),
		bson.M{"username": username},
		bson.M{
			"$push": bson.M{
				"hatList.hats": newHat,
			},
		},
	)
	return err
}
