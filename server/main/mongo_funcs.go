package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRecord struct {
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
	LastLogin  time.Time `bson:"lastLogin,omitempty"`
	LastLogout time.Time `bson:"lastLogout,omitempty"`
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

type EventRecord struct {
	Owner     string    `bson:"owner"`
	Secondary string    `bson:"secondary"`
	Type      string    `bson:"eventtype"`
	Created   time.Time `bson:"created"`
	StageName string    `bson:"stagename,omitempty"`
	X         int       `bson:"x,omitempty"`
	Y         int       `bson:"y,omitempty"`
	Details   string    `bson:"details,omitempty"` // Could be interface, no purpose
}

type SessionDataRecord struct {
	ServerName             string              `bson:"serverName"`
	Timestamp              time.Time           `bson:"timestamp"`
	SessionStartTime       time.Time           `bson:"sessionStartTime"`
	PeakSessionPlayerCount int                 `bson:"peakSessionPlayerCount"`
	PeakSessionKillSteak   SessionStreakRecord `bson:"peakSessionKillSteak"`
	TotalSessionLogins     int                 `bson:"totalSessionLogins"`
	TotalSessionLogouts    int                 `bson:"totalSessionLogouts"`
	CurrentTeamPlayerCount map[string]int      `bson:"currentTeamPlayerCount"`
	Scoreboard             map[string]int      `bson:"scoreboard"`
}

type SessionStreakRecord struct {
	Streak     int    `bson:"streak"`
	PlayerName string `bson:"playerName"`
}

///////////////////////////////////////////////////////////
// User Record

func (db *DB) getAuthorizedUserById(identifier string) *UserRecord {
	var result UserRecord
	collection := db.users
	err := collection.FindOne(context.TODO(), bson.M{"identifier": bson.M{"$eq": identifier}}).Decode(&result)
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

func (db *DB) insertAuthorizedUser(user UserRecord) error {
	_, err := db.users.InsertOne(context.TODO(), user)
	return err
}

func (db *DB) updateUsernameForUserWithId(identifier, username string) bool {
	filter := bson.M{"identifier": bson.M{"$eq": identifier}, "username": ""}
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

/////////////////////////////////////////////////////////////
//  Player Record

func (db *DB) InsertPlayerRecord(player PlayerRecord) error {
	_, err := db.playerRecords.InsertOne(context.TODO(), player)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) getPlayerRecord(username string) (PlayerRecord, error) {
	collection := db.playerRecords
	var result PlayerRecord
	err := collection.FindOne(context.TODO(), bson.M{"username": bson.M{"$eq": username}}).Decode(&result)
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

func (db *DB) updateRecordForPlayer(p *Player, pTile *Tile) error {
	_, err := db.playerRecords.UpdateOne(
		context.TODO(),
		bson.M{"username": bson.M{"$eq": p.username}},
		bson.M{
			"$set": createPlayerSnapShot(p, pTile),
		},
	)
	return err
}

func (db *DB) updateLoginForPlayer(p *Player) error {
	_, err := db.playerRecords.UpdateOne(
		context.TODO(),
		bson.M{"username": bson.M{"$eq": p.username}},
		bson.M{
			"$set": bson.M{
				"lastLogin": time.Now(),
			},
		},
	)
	return err
}

func (db *DB) updatePlayerRecordOnLogout(p *Player, pTile *Tile) error {
	snapshot := createPlayerSnapShot(p, pTile)
	snapshot["lastLogout"] = time.Now()
	_, err := db.playerRecords.UpdateOne(
		context.TODO(),
		bson.M{"username": bson.M{"$eq": p.username}},
		bson.M{
			"$set": snapshot,
		},
	)
	return err
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

func createPlayerSnapShot(p *Player, pTile *Tile) bson.M {
	return bson.M{
		"x":               pTile.x,
		"y":               pTile.y,
		"health":          p.getHealthSync(),
		"stagename":       pTile.stage.name,
		"money":           p.getMoneySync(),
		"killCount":       p.getKillCountSync(),
		"deathCount":      p.getDeathCountSync(),
		"goalsScored":     p.getGoalsScored(),
		"hatList.current": p.hatList.indexSync(),
	}
}

//////////////////////////////////////////////////////////////////////
// Event Records

func (db *DB) saveKillEvent(tile *Tile, initiator *Player, defeated *Player) error {
	eventCollection := db.events
	event := EventRecord{
		Owner:     initiator.username,
		Secondary: defeated.username,
		Type:      "Kill",
		Created:   time.Now(),
		StageName: tile.stage.name,
		X:         tile.x,
		Y:         tile.y,
	}
	_, err := eventCollection.InsertOne(context.TODO(), event)
	if err != nil {
		log.Fatal("Event Insert Failed")
	}

	return nil
}

func (db *DB) saveScoreEvent(tile *Tile, initiator *Player, message string) error {
	eventCollection := db.events
	event := EventRecord{
		Owner:     initiator.username,
		Secondary: "",
		Type:      "Score",
		Created:   time.Now(),
		StageName: tile.stage.name,
		X:         tile.x,
		Y:         tile.y,
		Details:   message,
	}
	_, err := eventCollection.InsertOne(context.TODO(), event)
	if err != nil {
		log.Fatal("Event Insert Failed")
	}

	return nil
}

//////////////////////////////////////////////////////////////////////
// Highscores

func (db *DB) getTopNPlayersByField(field string, n int) ([]PlayerRecord, error) {
	// Should add indexes where needed
	findOptions := options.Find().
		SetSort(bson.D{{Key: field, Value: -1}}).
		SetLimit(int64(n))

	// Impact of adding a team filter ?
	cursor, err := db.playerRecords.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var results []PlayerRecord
	if err := cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}

///////////////////////////////////////////////////////////////////////
// Game Status Funcs

func getMostRecentSessionData(ctx context.Context, collection *mongo.Collection, serverName string) (*SessionDataRecord, error) {
	filter := bson.M{"serverName": serverName}
	// Sort by timestamp in descending order to get the most recent document.
	findOpts := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	var result SessionDataRecord
	err := collection.FindOne(ctx, filter, findOpts).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func saveGameStatus(ctx context.Context, collection *mongo.Collection, status SessionDataRecord) error {
	_, err := collection.InsertOne(ctx, status)
	return err
}
