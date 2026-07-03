package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func (s *WorldStore) seedWorld(ctx context.Context, world *WorldDocument, seedRoot, seedName string) error {
	if err := validateResourcePart(seedName); err != nil {
		return err
	}
	collection, err := loadSeedCollection(seedRoot, seedName)
	if err != nil {
		return err
	}
	palettePath := filepath.Join(filepath.Dir(seedRoot), "colors", "colors.json")
	paletteJSON, err := os.ReadFile(palettePath)
	if err != nil {
		return fmt.Errorf("read seed palette: %w", err)
	}
	entry := firstSeedStage(collection)
	manifest := WorldManifest{Name: world.Name, Entry: WorldLocation{Stage: entry, Y: 3, X: 3}, Teams: []WorldTeam{{ID: "sky-blue", Label: "Sky Blue", Color: "sky-blue", Spawn: WorldLocation{Stage: entry, Y: 3, X: 3}}, {ID: "fuchsia", Label: "Fuchsia", Color: "fuchsia", Spawn: WorldLocation{Stage: entry, Y: 3, X: 3}}}, DefaultTeam: "sky-blue", MaxPlayers: 400, Lifecycle: world.Lifecycle, Leaderboards: []LeaderboardDefinition{{ID: "richest", Label: "Richest", Metric: "peakWealth"}, {ID: "deadliest", Label: "Deadliest", Metric: "peakKillStreak"}, {ID: "mvp", Label: "MVP", Metric: "goalsScored"}}}
	manifestJSON, _ := json.Marshal(manifest)
	now := time.Now().UTC()
	resources := []any{}
	metaJSON, err := json.Marshal(map[string]string{"name": collection.Name})
	if err != nil {
		return err
	}
	resources = append(resources,
		WorldResource{WorldID: world.ID, Kind: "collection", Key: "meta", Revision: 1, Content: metaJSON, Size: len(metaJSON), UpdatedAt: now},
		WorldResource{WorldID: world.ID, Kind: "palette", Key: "colors", Revision: 1, Content: paletteJSON, Size: len(paletteJSON), UpdatedAt: now},
		WorldResource{WorldID: world.ID, Kind: "manifest", Key: "world", Revision: 1, Content: manifestJSON, Size: len(manifestJSON), UpdatedAt: now},
	)

	spaceNames := make([]string, 0, len(collection.Spaces))
	for name := range collection.Spaces {
		spaceNames = append(spaceNames, name)
	}
	sort.Strings(spaceNames)
	for _, name := range spaceNames {
		payload, marshalErr := json.Marshal(collection.Spaces[name])
		if marshalErr != nil {
			return marshalErr
		}
		resources = append(resources, WorldResource{WorldID: world.ID, Kind: "space", Key: name, Revision: 1, Content: payload, Size: len(payload), UpdatedAt: now})
	}

	prototypeSetNames := make([]string, 0, len(collection.PrototypeSets))
	for name := range collection.PrototypeSets {
		prototypeSetNames = append(prototypeSetNames, name)
	}
	sort.Strings(prototypeSetNames)
	for _, name := range prototypeSetNames {
		payload, marshalErr := json.Marshal(collection.PrototypeSets[name])
		if marshalErr != nil {
			return marshalErr
		}
		resources = append(resources, WorldResource{WorldID: world.ID, Kind: "prototype-set", Key: name, Revision: 1, Content: payload, Size: len(payload), UpdatedAt: now})
	}

	interactableSetNames := make([]string, 0, len(collection.InteractableSets))
	for name := range collection.InteractableSets {
		interactableSetNames = append(interactableSetNames, name)
	}
	sort.Strings(interactableSetNames)
	for _, name := range interactableSetNames {
		payload, marshalErr := json.Marshal(collection.InteractableSets[name])
		if marshalErr != nil {
			return marshalErr
		}
		resources = append(resources, WorldResource{WorldID: world.ID, Kind: "interactable-set", Key: name, Revision: 1, Content: payload, Size: len(payload), UpdatedAt: now})
	}

	fragmentSetNames := make([]string, 0, len(collection.Fragments))
	for name := range collection.Fragments {
		fragmentSetNames = append(fragmentSetNames, name)
	}
	sort.Strings(fragmentSetNames)
	for _, name := range fragmentSetNames {
		payload := append([]byte(nil), collection.Fragments[name]...)
		resources = append(resources, WorldResource{WorldID: world.ID, Kind: "fragment-set", Key: name, Revision: 1, Content: payload, Size: len(payload), UpdatedAt: now})
	}
	_, err = s.db.worldResources.InsertMany(ctx, resources)
	return err
}

func loadSeedCollection(root, name string) (SourceCollection, error) {
	base := filepath.Join(root, name)
	collection := SourceCollection{Name: name, Spaces: map[string]*SourceSpace{}, Fragments: map[string]json.RawMessage{}, PrototypeSets: map[string][]SourcePrototype{}, InteractableSets: map[string][]InteractableDescription{}}
	if err := loadSeedDirectory(filepath.Join(base, "spaces"), func(key string, data []byte) error {
		var value SourceSpace
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		collection.Spaces[key] = &value
		return nil
	}); err != nil {
		return collection, err
	}
	if err := loadSeedDirectory(filepath.Join(base, "prototypes"), func(key string, data []byte) error {
		var value []SourcePrototype
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		collection.PrototypeSets[key] = value
		return nil
	}); err != nil {
		return collection, err
	}
	if err := loadSeedDirectory(filepath.Join(base, "interactables"), func(key string, data []byte) error {
		var value []InteractableDescription
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		collection.InteractableSets[key] = value
		return nil
	}); err != nil {
		return collection, err
	}
	_ = loadSeedDirectory(filepath.Join(base, "fragments"), func(key string, data []byte) error {
		collection.Fragments[key] = append(json.RawMessage(nil), data...)
		return nil
	})
	return collection, nil
}

func loadSeedDirectory(directory string, accept func(string, []byte) error) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			return err
		}
		if err := accept(strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())), data); err != nil {
			return fmt.Errorf("%s: %w", entry.Name(), err)
		}
	}
	return nil
}

func firstSeedStage(collection SourceCollection) string {
	names := make([]string, 0, len(collection.Spaces))
	for name := range collection.Spaces {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		space := collection.Spaces[name]
		if space != nil && len(space.Areas) > 0 {
			return space.Areas[0].Name
		}
	}
	return ""
}
func seedRootFromEnvironment() string {
	if value := strings.TrimSpace(os.Getenv("WORLD_SEED_DIR")); value != "" {
		return value
	}
	if info, err := os.Stat(filepath.Join("seeds", "collections")); err == nil && info.IsDir() {
		return filepath.Join("seeds", "collections")
	}
	return filepath.Clean(filepath.Join("..", "..", "tools", "main", "data", "collections"))
}
func (s *WorldStore) removeFailedWorld(ctx context.Context, worldID string) {
	_, _ = s.db.worldResources.DeleteMany(ctx, bson.M{"worldId": worldID})
	_, _ = s.db.worlds.DeleteOne(ctx, bson.M{"_id": worldID})
}
