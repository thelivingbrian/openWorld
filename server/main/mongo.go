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

type UserRecord struct {
	Identifier    string     `bson:"identifier"`
	Username      string     `bson:"username"`
	CreationEmail string     `bson:"creationEmail"`
	Created       time.Time  `bson:"created,omitempty"`
	LastLogin     time.Time  `bson:"lastLogin,omitempty"`
	BanReason     string     `bson:"banReason,omitempty"`
	BannedBy      string     `bson:"bannedBy,omitempty"`
	BanExpiresAt  *time.Time `bson:"banExpiresAt,omitempty"`
}

type PlayerRecord struct {
	// ID
	Username string `bson:"username"`
	UserID   string `bson:"userId,omitempty"`
	WorldID  string `bson:"worldId,omitempty"`

	// Meta
	LastLogin  time.Time `bson:"lastLogin,omitempty"`
	LastLogout time.Time `bson:"lastLogout,omitempty"`
	// total logins / cumulative play time
	GuestCreateTime *time.Time `bson:"guestCreateTime,omitempty"`

	// World Location
	StageName string `bson:"stagename"`
	X         int    `bson:"x"`
	Y         int    `bson:"y"`

	// Details
	Team   string            `bson:"team"`
	Health int64             `bson:"health"`
	Money  int64             `bson:"money"`
	Stats  PlayerStatsRecord `bson:"stats"`

	// Unlocks
	Accomplishments map[string]Accomplishment `bson:"accomplishments,omitempty"`
}

type PlayerStatsRecord struct {
	// Stats
	KillCount      int64 `bson:"killCount,omitempty"`
	KillCountNpc   int64 `bson:"killCountNpc,omitempty"`
	PeakKillStreak int64 `bson:"peakKillStreak,omitempty"`
	DeathCount     int64 `bson:"deathCount,omitempty"`
	GoalsScored    int64 `bson:"goalsScored,omitempty"`
	PeakWealth     int64 `bson:"peakWealth,omitempty"`
}

type EventRecord struct {
	WorldID   string    `bson:"worldId,omitempty"`
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
	WorldID                string              `bson:"worldId,omitempty"`
	ServerName             string              `bson:"serverName"`
	Timestamp              time.Time           `bson:"timestamp"`
	SessionStartTime       time.Time           `bson:"sessionStartTime"`
	PeakSessionPlayerCount int                 `bson:"peakSessionPlayerCount"`
	PeakSessionKillStreak  SessionStreakRecord `bson:"peakSessionKillStreak"`
	TotalSessionLogins     int                 `bson:"totalSessionLogins"`
	TotalSessionLogouts    int                 `bson:"totalSessionLogouts"`
	CurrentTeamPlayerCount map[string]int      `bson:"currentTeamPlayerCount"`
	Scoreboard             map[string]int      `bson:"scoreboard"`
}

type SessionStreakRecord struct {
	Streak     int    `bson:"streak"`
	PlayerName string `bson:"playerName"`
}

type AdminActionRecord struct {
	ActionType        string    `bson:"actionType"`
	ActingAdmin       string    `bson:"actingAdmin"`
	TargetPlayer      string    `bson:"targetPlayer,omitempty"`
	TargetIdentifier  string    `bson:"targetIdentifier,omitempty"`
	Payload           bson.M    `bson:"payload,omitempty"`
	ResultStatus      string    `bson:"resultStatus"`
	ResultDescription string    `bson:"resultDescription,omitempty"`
	Created           time.Time `bson:"created"`
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

func (db *DB) getAuthorizedUserByUsername(username string) *UserRecord {
	var result UserRecord
	err := db.users.FindOne(context.TODO(), bson.M{"username": bson.M{"$eq": username}}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		log.Fatal(err)
	}
	return &result
}

func (db *DB) insertAuthorizedUser(user UserRecord) error {
	_, err := db.users.InsertOne(context.TODO(), user)
	return err
}

func (db *DB) updateLastLoginForUserWithId(identifier string) error {
	_, err := db.users.UpdateOne(
		context.TODO(),
		bson.M{"identifier": bson.M{"$eq": identifier}},
		bson.M{"$set": bson.M{"lastLogin": time.Now()}},
	)
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

func (db *DB) upsertBanForIdentifier(identifier, actingAdmin, reason string, banExpiresAt *time.Time) error {
	_, err := db.users.UpdateOne(
		context.TODO(),
		bson.M{"identifier": bson.M{"$eq": identifier}},
		bson.M{"$set": bson.M{
			"bannedBy":     actingAdmin,
			"banReason":    reason,
			"banExpiresAt": banExpiresAt,
		}},
	)
	return err
}

func (db *DB) clearBanForIdentifier(identifier string) error {
	_, err := db.users.UpdateOne(
		context.TODO(),
		bson.M{"identifier": bson.M{"$eq": identifier}},
		bson.M{"$unset": bson.M{
			"bannedBy":     "",
			"banReason":    "",
			"banExpiresAt": "",
		}},
	)
	return err
}

func (db *DB) saveAdminAction(record AdminActionRecord) error {
	if db.adminActions == nil {
		return nil
	}
	_, err := db.adminActions.InsertOne(context.TODO(), record)
	return err
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
	return db.getWorldPlayerRecord(legacyWorldID, username)
}

func (db *DB) getWorldPlayerRecord(worldID, username string) (PlayerRecord, error) {
	collection := db.playerRecords
	var result PlayerRecord
	filter := worldPlayerFilter(worldID, username)
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Error().Err(err).Msg("No document was found with the given username")
			return PlayerRecord{Username: "invalild"}, err
		} else {
			log.Fatal(err)
		}
	}
	return result, nil
}

func worldPlayerFilter(worldID, username string) bson.M {
	if worldID == legacyWorldID {
		return bson.M{"username": username, "$or": bson.A{bson.M{"worldId": legacyWorldID}, bson.M{"worldId": bson.M{"$exists": false}}}}
	}
	return bson.M{"username": username, "worldId": worldID}
}

func (db *DB) foundUsername(username string) bool {
	_, err := db.getPlayerRecord(username)
	return err == nil
}

func (db *DB) updateRecordForPlayer(p *Player, pTile *Tile) error {
	_, err := db.playerRecords.UpdateOne(
		context.TODO(),
		worldPlayerFilter(p.world.config.worldID, p.username),
		bson.M{
			"$set": createPlayerSnapShot(p, pTile),
		},
	)
	return err
}

func (db *DB) updateLoginForPlayer(p *Player) error {
	_, err := db.playerRecords.UpdateOne(
		context.TODO(),
		worldPlayerFilter(p.world.config.worldID, p.username),
		bson.M{
			"$set": bson.M{
				"lastLogin": time.Now(),
			}, // Increment a counter?
		},
	)
	return err
}

func (db *DB) updatePlayerRecordOnLogout(p *Player, pTile *Tile) error {
	snapshot := createPlayerSnapShot(p, pTile)
	snapshot["lastLogout"] = time.Now()
	_, err := db.playerRecords.UpdateOne(
		context.TODO(),
		worldPlayerFilter(p.world.config.worldID, p.username),
		bson.M{
			"$set": snapshot,
		},
	)
	return err
}

func (db *DB) adminUpdatePlayerRecord(worldID, username string, updates bson.M) error {
	_, err := db.playerRecords.UpdateOne(
		context.TODO(),
		worldPlayerFilter(worldID, username),
		bson.M{"$set": updates},
	)
	return err
}

func (db *DB) addAccomplishmentToPlayer(worldID, username string, key string, value Accomplishment) error {
	_, err := db.playerRecords.UpdateOne(
		context.TODO(),
		worldPlayerFilter(worldID, username),
		bson.M{
			"$set": bson.M{
				fmt.Sprintf("accomplishments.%s", key): value,
			},
		},
	)
	return err
}

func createPlayerSnapShot(p *Player, pTile *Tile) bson.M {
	return bson.M{
		"x":         pTile.x,
		"y":         pTile.y,
		"health":    p.health.Load(),
		"stagename": pTile.stage.name,
		"money":     p.money.Load(),
		"stats":     statsRecordFromPlayerStats(&p.PlayerStats),
	}
}

func statsRecordFromPlayerStats(stats *PlayerStats) PlayerStatsRecord {
	return PlayerStatsRecord{
		KillCount:      stats.killCount.Load(),
		KillCountNpc:   stats.killCountNpc.Load(),
		DeathCount:     stats.deathCount.Load(),
		GoalsScored:    stats.goalsScored.Load(),
		PeakKillStreak: stats.peakKillStreak.Load(),
		PeakWealth:     stats.peakWealth.Load(),
	}
}

//////////////////////////////////////////////////////////////////////
// Event Records

func (db *DB) saveKillEvent(tile *Tile, initiator Character, defeated *Player) error {
	eventCollection := db.events
	event := EventRecord{
		WorldID:   defeated.world.config.worldID,
		Owner:     initiator.getName(),
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
		WorldID:   initiator.world.config.worldID,
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
	return db.getTopNPlayersByFieldForWorld(legacyWorldID, field, n)
}

func (db *DB) getTopNPlayersByFieldForWorld(worldID, field string, n int) ([]PlayerRecord, error) {
	// Should add indexes where needed
	findOptions := options.Find().
		SetSort(bson.D{{Key: field, Value: -1}}).
		SetLimit(int64(n))

	// Impact of adding a team filter ?
	filter := bson.M{"worldId": worldID}
	if worldID == legacyWorldID {
		filter = bson.M{"$or": bson.A{bson.M{"worldId": legacyWorldID}, bson.M{"worldId": bson.M{"$exists": false}}}}
	}
	cursor, err := db.playerRecords.Find(context.TODO(), filter, findOptions)
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
