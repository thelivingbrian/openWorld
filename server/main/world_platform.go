package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	legacyWorldID       = "legacy"
	maxOwnedWorlds      = 5
	maxWorldSourceBytes = 25 << 20
	maxResourceBytes    = 2 << 20
	maxPaletteEntries   = 128
	maxWorldTiles       = 250_000
	maxAreaSide         = 64
)

type WorldLifecycle string

const (
	LifecycleOwnerPresent WorldLifecycle = "owner-present"
	LifecycleUntilEmpty   WorldLifecycle = "until-empty"
	LifecyclePersistent   WorldLifecycle = "persistent"
)

type WorldDocument struct {
	ID                 string         `bson:"_id" json:"id"`
	OwnerID            string         `bson:"ownerId" json:"ownerId"`
	Name               string         `bson:"name" json:"name"`
	Slug               string         `bson:"slug,omitempty" json:"slug,omitempty"`
	Description        string         `bson:"description,omitempty" json:"description,omitempty"`
	Visibility         string         `bson:"visibility" json:"visibility"`
	ModerationState    string         `bson:"moderationState" json:"moderationState"`
	Lifecycle          WorldLifecycle `bson:"lifecycle" json:"lifecycle"`
	DraftGeneration    int64          `bson:"draftGeneration" json:"draftGeneration"`
	PublishedReleaseID string         `bson:"publishedReleaseId,omitempty" json:"publishedReleaseId,omitempty"`
	CreatedAt          time.Time      `bson:"createdAt" json:"createdAt"`
	UpdatedAt          time.Time      `bson:"updatedAt" json:"updatedAt"`
	DeletedAt          *time.Time     `bson:"deletedAt,omitempty" json:"-"`
}

type WorldResource struct {
	WorldID   string          `bson:"worldId" json:"worldId"`
	Kind      string          `bson:"kind" json:"kind"`
	Key       string          `bson:"key" json:"key"`
	Revision  int64           `bson:"revision" json:"revision"`
	Content   json.RawMessage `bson:"content" json:"content"`
	Size      int             `bson:"size" json:"size"`
	UpdatedAt time.Time       `bson:"updatedAt" json:"updatedAt"`
}

type WorldRelease struct {
	ID              string             `bson:"_id" json:"id"`
	WorldID         string             `bson:"worldId" json:"worldId"`
	Number          int64              `bson:"number" json:"number"`
	DraftGeneration int64              `bson:"draftGeneration" json:"draftGeneration"`
	SourceHash      string             `bson:"sourceHash" json:"sourceHash"`
	ArtifactHash    string             `bson:"artifactHash" json:"artifactHash"`
	CompilerVersion string             `bson:"compilerVersion" json:"compilerVersion"`
	SourceFileID    primitive.ObjectID `bson:"sourceFileId" json:"-"`
	ArtifactFileID  primitive.ObjectID `bson:"artifactFileId" json:"-"`
	CreatedBy       string             `bson:"createdBy" json:"createdBy"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
}

type WorldTeam struct {
	ID    string        `json:"id"`
	Label string        `json:"label"`
	Color string        `json:"color"`
	Spawn WorldLocation `json:"spawn"`
}

type WorldLocation struct {
	Stage string `json:"stage"`
	Y     int    `json:"y"`
	X     int    `json:"x"`
}

type LeaderboardDefinition struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Metric string `json:"metric"`
}

type WorldManifest struct {
	Name             string                  `json:"name"`
	Description      string                  `json:"description,omitempty"`
	Entry            WorldLocation           `json:"entry"`
	Teams            []WorldTeam             `json:"teams"`
	DefaultTeam      string                  `json:"defaultTeam"`
	MaxPlayers       int                     `json:"maxPlayers"`
	Lifecycle        WorldLifecycle          `json:"lifecycle"`
	Leaderboards     []LeaderboardDefinition `json:"leaderboards,omitempty"`
	OnboardingStages []string                `json:"onboardingStages,omitempty"`
	OnboardingExit   *WorldLocation          `json:"onboardingExit,omitempty"`
}

type WorldColor struct {
	Name string  `json:"name"`
	R    uint8   `json:"r"`
	G    uint8   `json:"g"`
	B    uint8   `json:"b"`
	A    float64 `json:"a"`
}

type WorldMapArea struct {
	MapID  string  `json:"mapId"`
	Top    float64 `json:"top"`
	Left   float64 `json:"left"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type CompiledRelease struct {
	Source       []byte
	Artifact     []byte
	SourceHash   string
	ArtifactHash string
}

var resourcePartPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$`)
var paletteNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,47}$`)
var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,47}[a-z0-9]$`)

func validateResourcePart(value string) error {
	if !resourcePartPattern.MatchString(value) || strings.Contains(value, "..") {
		return fmt.Errorf("invalid resource identifier %q", value)
	}
	return nil
}

func validateManifest(manifest WorldManifest, admin bool) error {
	if strings.TrimSpace(manifest.Name) == "" || len(manifest.Name) > 80 {
		return errors.New("manifest name is required and must be at most 80 characters")
	}
	if manifest.Entry.Stage == "" {
		return errors.New("manifest entry stage is required")
	}
	if manifest.MaxPlayers < 1 || manifest.MaxPlayers > 500 {
		return errors.New("maxPlayers must be between 1 and 500")
	}
	if manifest.Lifecycle != LifecycleOwnerPresent && manifest.Lifecycle != LifecycleUntilEmpty && manifest.Lifecycle != LifecyclePersistent {
		return errors.New("invalid lifecycle")
	}
	if manifest.Lifecycle == LifecyclePersistent && !admin {
		return errors.New("persistent lifecycle requires an administrator")
	}
	teamIDs := map[string]bool{}
	for _, team := range manifest.Teams {
		if err := validateResourcePart(team.ID); err != nil {
			return fmt.Errorf("team: %w", err)
		}
		if teamIDs[team.ID] {
			return fmt.Errorf("duplicate team %q", team.ID)
		}
		teamIDs[team.ID] = true
		if team.Spawn.Stage == "" {
			return fmt.Errorf("team %q has no spawn stage", team.ID)
		}
	}
	if len(manifest.Teams) > 0 && !teamIDs[manifest.DefaultTeam] {
		return errors.New("defaultTeam must reference a declared team")
	}
	allowedMetrics := map[string]bool{"peakWealth": true, "peakKillStreak": true, "goalsScored": true, "killCount": true, "deathCount": true}
	for _, board := range manifest.Leaderboards {
		if !allowedMetrics[board.Metric] {
			return fmt.Errorf("leaderboard metric %q is not allowed", board.Metric)
		}
	}
	return nil
}

func validatePalette(colors []WorldColor) error {
	if len(colors) > maxPaletteEntries {
		return fmt.Errorf("palette exceeds %d entries", maxPaletteEntries)
	}
	seen := map[string]bool{}
	for _, color := range colors {
		if !paletteNamePattern.MatchString(color.Name) {
			return fmt.Errorf("invalid palette name %q", color.Name)
		}
		if seen[color.Name] {
			return fmt.Errorf("duplicate palette name %q", color.Name)
		}
		if color.A < 0 || color.A > 1 {
			return fmt.Errorf("palette alpha for %q must be between 0 and 1", color.Name)
		}
		seen[color.Name] = true
	}
	return nil
}

func canonicalResources(resources []WorldResource) ([]byte, string, error) {
	sort.Slice(resources, func(i, j int) bool {
		if resources[i].Kind == resources[j].Kind {
			return resources[i].Key < resources[j].Key
		}
		return resources[i].Kind < resources[j].Kind
	})
	type entry struct {
		Kind    string          `json:"kind"`
		Key     string          `json:"key"`
		Content json.RawMessage `json:"content"`
	}
	out := make([]entry, 0, len(resources))
	for _, resource := range resources {
		if !json.Valid(resource.Content) {
			return nil, "", fmt.Errorf("%s/%s contains invalid JSON", resource.Kind, resource.Key)
		}
		var normalized any
		if err := json.Unmarshal(resource.Content, &normalized); err != nil {
			return nil, "", err
		}
		content, err := json.Marshal(normalized)
		if err != nil {
			return nil, "", err
		}
		out = append(out, entry{resource.Kind, resource.Key, content})
	}
	data, err := json.Marshal(out)
	if err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(data)
	return data, hex.EncodeToString(sum[:]), nil
}

type WorldStore struct {
	db     *DB
	bucket *gridfs.Bucket
}

func newWorldStore(db *DB) (*WorldStore, error) {
	if db == nil || db.database == nil {
		return nil, errors.New("world store requires MongoDB")
	}
	bucket, err := gridfs.NewBucket(db.database, options.GridFSBucket().SetName("worldReleaseFiles"))
	if err != nil {
		return nil, err
	}
	return &WorldStore{db: db, bucket: bucket}, nil
}

func (s *WorldStore) ensureIndexes(ctx context.Context) error {
	if err := s.migrateLegacyProfiles(ctx); err != nil {
		return err
	}
	models := []mongo.IndexModel{
		{Keys: bson.D{{Key: "slug", Value: 1}}, Options: options.Index().SetUnique(true).SetSparse(true)},
		{Keys: bson.D{{Key: "ownerId", Value: 1}, {Key: "updatedAt", Value: -1}}},
	}
	if _, err := s.db.worlds.Indexes().CreateMany(ctx, models); err != nil {
		return err
	}
	if _, err := s.db.worldResources.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "worldId", Value: 1}, {Key: "kind", Value: 1}, {Key: "key", Value: 1}}, Options: options.Index().SetUnique(true)}); err != nil {
		return err
	}
	if _, err := s.db.worldReleases.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "worldId", Value: 1}, {Key: "number", Value: -1}}, Options: options.Index().SetUnique(true)}); err != nil {
		return err
	}
	if _, err := s.db.playerRecords.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "worldId", Value: 1}, {Key: "userId", Value: 1}}, Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.M{"worldId": bson.M{"$exists": true}, "userId": bson.M{"$gt": ""}})}); err != nil {
		return err
	}
	_, err := s.db.runtimeInstances.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)})
	return err
}

func (s *WorldStore) migrateLegacyProfiles(ctx context.Context) error {
	if _, err := s.db.playerRecords.UpdateMany(ctx, bson.M{"worldId": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"worldId": legacyWorldID}}); err != nil {
		return err
	}
	if _, err := s.db.events.UpdateMany(ctx, bson.M{"worldId": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"worldId": legacyWorldID}}); err != nil {
		return err
	}
	if _, err := s.db.sessionData.UpdateMany(ctx, bson.M{"worldId": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"worldId": legacyWorldID}}); err != nil {
		return err
	}
	cursor, err := s.db.users.Find(ctx, bson.M{"username": bson.M{"$ne": ""}}, options.Find().SetProjection(bson.M{"identifier": 1, "username": 1}))
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var user UserRecord
		if err := cursor.Decode(&user); err != nil {
			return err
		}
		if _, err := s.db.playerRecords.UpdateMany(ctx, bson.M{"worldId": legacyWorldID, "username": user.Username, "$or": bson.A{bson.M{"userId": ""}, bson.M{"userId": bson.M{"$exists": false}}}}, bson.M{"$set": bson.M{"userId": user.Identifier}}); err != nil {
			return err
		}
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	guests, err := s.db.playerRecords.Find(ctx, bson.M{"worldId": legacyWorldID, "username": bson.M{"$regex": "^guest:"}, "$or": bson.A{bson.M{"userId": ""}, bson.M{"userId": bson.M{"$exists": false}}}}, options.Find().SetProjection(bson.M{"username": 1}))
	if err != nil {
		return err
	}
	defer guests.Close(ctx)
	for guests.Next(ctx) {
		var player PlayerRecord
		if err := guests.Decode(&player); err != nil {
			return err
		}
		if _, err := s.db.playerRecords.UpdateMany(ctx, bson.M{"worldId": legacyWorldID, "username": player.Username}, bson.M{"$set": bson.M{"userId": player.Username}}); err != nil {
			return err
		}
	}
	return guests.Err()
}

func (s *WorldStore) createWorld(ctx context.Context, ownerID, name string, bypassQuota ...bool) (*WorldDocument, error) {
	count, err := s.db.worlds.CountDocuments(ctx, bson.M{"ownerId": ownerID, "deletedAt": bson.M{"$exists": false}})
	if err != nil {
		return nil, err
	}
	if count >= maxOwnedWorlds && (len(bypassQuota) == 0 || !bypassQuota[0]) {
		return nil, fmt.Errorf("world quota of %d reached", maxOwnedWorlds)
	}
	now := time.Now().UTC()
	world := &WorldDocument{ID: uuid.NewString(), OwnerID: ownerID, Name: strings.TrimSpace(name), Visibility: "public", ModerationState: "active", Lifecycle: LifecycleOwnerPresent, DraftGeneration: 1, CreatedAt: now, UpdatedAt: now}
	if world.Name == "" {
		return nil, errors.New("world name is required")
	}
	_, err = s.db.worlds.InsertOne(ctx, world)
	return world, err
}

func (s *WorldStore) getWorld(ctx context.Context, route string) (*WorldDocument, error) {
	filter := bson.M{"deletedAt": bson.M{"$exists": false}, "$or": bson.A{bson.M{"_id": route}, bson.M{"slug": route}}}
	var world WorldDocument
	if err := s.db.worlds.FindOne(ctx, filter).Decode(&world); err != nil {
		return nil, err
	}
	return &world, nil
}

func (s *WorldStore) listWorlds(ctx context.Context, filter bson.M) ([]WorldDocument, error) {
	filter["deletedAt"] = bson.M{"$exists": false}
	cursor, err := s.db.worlds.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "updatedAt", Value: -1}}).SetLimit(100))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var worlds []WorldDocument
	return worlds, cursor.All(ctx, &worlds)
}

func (s *WorldStore) resources(ctx context.Context, worldID string) ([]WorldResource, error) {
	cursor, err := s.db.worldResources.Find(ctx, bson.M{"worldId": worldID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var resources []WorldResource
	return resources, cursor.All(ctx, &resources)
}

var ErrRevisionConflict = errors.New("resource revision conflict")

func (s *WorldStore) putResource(ctx context.Context, resource WorldResource, expected int64) (*WorldResource, error) {
	if err := validateResourcePart(resource.Kind); err != nil {
		return nil, err
	}
	if err := validateResourcePart(resource.Key); err != nil {
		return nil, err
	}
	if len(resource.Content) > maxResourceBytes || !json.Valid(resource.Content) {
		return nil, errors.New("resource is invalid or too large")
	}
	cursor, err := s.db.worldResources.Find(ctx, bson.M{"worldId": resource.WorldID, "$nor": bson.A{bson.M{"kind": resource.Kind, "key": resource.Key}}}, options.Find().SetProjection(bson.M{"size": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	total := len(resource.Content)
	for cursor.Next(ctx) {
		var existing struct {
			Size int `bson:"size"`
		}
		if err := cursor.Decode(&existing); err != nil {
			return nil, err
		}
		total += existing.Size
		if total > maxWorldSourceBytes {
			return nil, fmt.Errorf("world source exceeds %d bytes", maxWorldSourceBytes)
		}
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	filter := bson.M{"worldId": resource.WorldID, "kind": resource.Kind, "key": resource.Key, "revision": expected}
	update := bson.M{"$set": bson.M{"content": resource.Content, "size": len(resource.Content), "updatedAt": now}, "$inc": bson.M{"revision": 1}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	if expected == 0 {
		opts.SetUpsert(true)
		filter["$or"] = bson.A{bson.M{"revision": 0}, bson.M{"revision": bson.M{"$exists": false}}}
	}
	var saved WorldResource
	err = s.db.worldResources.FindOneAndUpdate(ctx, filter, update, opts).Decode(&saved)
	if mongo.IsDuplicateKeyError(err) || err == mongo.ErrNoDocuments {
		return nil, ErrRevisionConflict
	}
	if err != nil {
		return nil, err
	}
	_, err = s.db.worlds.UpdateByID(ctx, resource.WorldID, bson.M{"$inc": bson.M{"draftGeneration": 1}, "$set": bson.M{"updatedAt": now}})
	return &saved, err
}

func (s *WorldStore) publish(ctx context.Context, world *WorldDocument, release CompiledRelease, actor string) (*WorldRelease, error) {
	latest := WorldRelease{}
	err := s.db.worldReleases.FindOne(ctx, bson.M{"worldId": world.ID}, options.FindOne().SetSort(bson.D{{Key: "number", Value: -1}})).Decode(&latest)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}
	number := latest.Number + 1
	sourceID, err := s.bucket.UploadFromStream(fmt.Sprintf("%s-%d-source.json", world.ID, number), bytes.NewReader(release.Source))
	if err != nil {
		return nil, err
	}
	artifactID, err := s.bucket.UploadFromStream(fmt.Sprintf("%s-%d-release.zip", world.ID, number), bytes.NewReader(release.Artifact))
	if err != nil {
		_ = s.bucket.Delete(sourceID)
		return nil, err
	}
	record := &WorldRelease{ID: uuid.NewString(), WorldID: world.ID, Number: number, DraftGeneration: world.DraftGeneration, SourceHash: release.SourceHash, ArtifactHash: release.ArtifactHash, CompilerVersion: "1", SourceFileID: sourceID, ArtifactFileID: artifactID, CreatedBy: actor, CreatedAt: time.Now().UTC()}
	if _, err = s.db.worldReleases.InsertOne(ctx, record); err != nil {
		_ = s.bucket.Delete(sourceID)
		_ = s.bucket.Delete(artifactID)
		return nil, err
	}
	result, err := s.db.worlds.UpdateOne(ctx, bson.M{"_id": world.ID, "draftGeneration": world.DraftGeneration}, bson.M{"$set": bson.M{"publishedReleaseId": record.ID, "updatedAt": time.Now().UTC()}})
	if err != nil || result.ModifiedCount != 1 {
		_, _ = s.db.worldReleases.DeleteOne(ctx, bson.M{"_id": record.ID})
		_ = s.bucket.Delete(sourceID)
		_ = s.bucket.Delete(artifactID)
		if err != nil {
			return nil, err
		}
		return nil, ErrRevisionConflict
	}
	_ = s.pruneReleases(ctx, world.ID, 20)
	return record, nil
}

func (s *WorldStore) pruneReleases(ctx context.Context, worldID string, retain int64) error {
	cursor, err := s.db.worldReleases.Find(ctx, bson.M{"worldId": worldID}, options.Find().SetSort(bson.D{{Key: "number", Value: -1}}).SetSkip(retain))
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	var releases []WorldRelease
	if err := cursor.All(ctx, &releases); err != nil {
		return err
	}
	for _, release := range releases {
		_ = s.bucket.Delete(release.SourceFileID)
		_ = s.bucket.Delete(release.ArtifactFileID)
		if _, err := s.db.worldReleases.DeleteOne(ctx, bson.M{"_id": release.ID}); err != nil {
			return err
		}
	}
	return nil
}

func (s *WorldStore) release(ctx context.Context, worldID, releaseID string) (*WorldRelease, error) {
	var release WorldRelease
	err := s.db.worldReleases.FindOne(ctx, bson.M{"_id": releaseID, "worldId": worldID}).Decode(&release)
	return &release, err
}

func (s *WorldStore) downloadArtifact(ctx context.Context, release *WorldRelease) ([]byte, error) {
	var out bytes.Buffer
	_, err := s.bucket.DownloadToStream(release.ArtifactFileID, &out)
	return out.Bytes(), err
}

func (s *WorldStore) rollback(ctx context.Context, worldID, releaseID string) error {
	if _, err := s.release(ctx, worldID, releaseID); err != nil {
		return err
	}
	_, err := s.db.worlds.UpdateByID(ctx, worldID, bson.M{"$set": bson.M{"publishedReleaseId": releaseID, "updatedAt": time.Now().UTC()}})
	return err
}

func hashBytes(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
