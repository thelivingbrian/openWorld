package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testReleaseResources(t *testing.T) []WorldResource {
	t.Helper()
	tiles := [][]SourceTile{{{PrototypeID: "floor"}, {PrototypeID: "floor"}}, {{PrototypeID: "floor"}, {PrototypeID: "floor"}}}
	area := SourceArea{Name: "rooms:0-0", Blueprint: &SourceBlueprint{Tiles: tiles, DefaultTileColor: "green", DefaultTileColor1: "brown"}}
	space := &SourceSpace{Name: "rooms", Topology: "plane", Latitude: 1, Longitude: 1, AreaHeight: 2, AreaWidth: 2, Areas: []SourceArea{area}}
	collection := SourceCollection{Name: "test", Spaces: map[string]*SourceSpace{"rooms": space}, PrototypeSets: map[string][]SourcePrototype{"default": {{ID: "floor", Walkable: true, MapColor: "green"}}}, InteractableSets: map[string][]InteractableDescription{}, Fragments: map[string]json.RawMessage{}}
	manifest := WorldManifest{Name: "Test", Entry: WorldLocation{Stage: "rooms:0-0"}, Teams: []WorldTeam{{ID: "blue", Label: "Blue", Color: "green", Spawn: WorldLocation{Stage: "rooms:0-0"}}}, DefaultTeam: "blue", MaxPlayers: 10, Lifecycle: LifecycleOwnerPresent, Leaderboards: []LeaderboardDefinition{{ID: "wealth", Label: "Wealth", Metric: "peakWealth"}}}
	palette := []WorldColor{{Name: "green", R: 1, G: 2, B: 3, A: 1}, {Name: "brown", R: 4, G: 5, B: 6, A: 1}}
	return []WorldResource{{Kind: "collection", Key: "source", Content: mustJSON(t, collection)}, {Kind: "manifest", Key: "world", Content: mustJSON(t, manifest)}, {Kind: "palette", Key: "colors", Content: mustJSON(t, palette)}}
}

func TestRenderSpaceMapUsesGroundAndPrototypeColors(t *testing.T) {
	space := &SourceSpace{
		Name: "rooms", Topology: "plane", Latitude: 1, Longitude: 1, AreaHeight: 1, AreaWidth: 3,
		Areas: []SourceArea{{Name: "rooms:0-0", Blueprint: &SourceBlueprint{
			Tiles:  [][]SourceTile{{{PrototypeID: "explicit"}, {PrototypeID: "layered"}, {PrototypeID: "plain"}}},
			Ground: [][]SourceCell{{{}, {}, {Status: 1}}}, DefaultTileColor: "ground", DefaultTileColor1: "alternate",
		}}},
	}
	colors := map[string]WorldColor{
		"ground": {Name: "ground", R: 1, G: 2, B: 3}, "alternate": {Name: "alternate", R: 4, G: 5, B: 6},
		"explicit-map": {Name: "explicit-map", R: 7, G: 8, B: 9}, "ceiling": {Name: "ceiling", R: 10, G: 11, B: 12},
	}
	prototypes := map[string]SourcePrototype{
		"explicit": {ID: "explicit", MapColor: "explicit-map", Ceiling2Css: "ceiling"},
		"layered":  {ID: "layered", CssColor: "ground", Floor1Css: "alternate", Ceiling2Css: "ceiling r0-tr"},
		"plain":    {ID: "plain"},
	}

	data, err := renderSpaceMap(space, prototypes, colors)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	wants := [][3]uint32{{7, 8, 9}, {10, 11, 12}, {4, 5, 6}}
	for x, want := range wants {
		r, g, b, a := img.At(x, 0).RGBA()
		got := [4]uint32{r >> 8, g >> 8, b >> 8, a >> 8}
		if got != [4]uint32{want[0], want[1], want[2], 255} {
			t.Fatalf("pixel %d = %v; want %v", x, got, want)
		}
	}
}

func mustJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestCompileWorldResourcesDeterministicAndSingleMap(t *testing.T) {
	first, err := compileWorldResources(testReleaseResources(t), false)
	if err != nil {
		t.Fatal(err)
	}
	second, err := compileWorldResources(testReleaseResources(t), false)
	if err != nil {
		t.Fatal(err)
	}
	if first.SourceHash != second.SourceHash || first.ArtifactHash != second.ArtifactHash || !bytes.Equal(first.Artifact, second.Artifact) {
		t.Fatal("identical sources produced different releases")
	}
	archive, err := zip.NewReader(bytes.NewReader(first.Artifact), int64(len(first.Artifact)))
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, entry := range archive.File {
		names[entry.Name] = true
		if filepath.Dir(entry.Name) == "images" {
			reader, openErr := entry.Open()
			if openErr != nil {
				t.Fatal(openErr)
			}
			data, readErr := io.ReadAll(reader)
			reader.Close()
			if readErr != nil {
				t.Fatal(readErr)
			}
			expected := "space-" + hashBytes(data)[:16] + ".png"
			if filepath.Base(entry.Name) != expected {
				t.Fatalf("map name %q is not content-addressed as %q", entry.Name, expected)
			}
		}
	}
	if !names["areas.json"] || !names["colors.css"] || !names["manifest.json"] {
		t.Fatalf("missing release files: %v", names)
	}
	maps := 0
	for name := range names {
		if filepath.Dir(name) == "images" {
			maps++
		}
	}
	if maps != 1 {
		t.Fatalf("got %d maps; want one base map per space", maps)
	}
}

func TestExtractReleaseArchiveRejectsTraversal(t *testing.T) {
	var encoded bytes.Buffer
	writer := zip.NewWriter(&encoded)
	entry, err := writer.Create("../escape")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = entry.Write([]byte("bad"))
	_ = writer.Close()
	if err := extractReleaseArchive(encoded.Bytes(), t.TempDir()); err == nil {
		t.Fatal("expected traversal archive to be rejected")
	}
}

func TestExtractReleaseArchiveWritesReadonlyFiles(t *testing.T) {
	release, err := compileWorldResources(testReleaseResources(t), false)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if err := extractReleaseArchive(release.Artifact, directory); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(directory, "areas.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0222 != 0 {
		t.Fatalf("release file is writable: %v", info.Mode())
	}
}

func TestPrefixRuntimeResponse(t *testing.T) {
	response := &http.Response{Header: http.Header{"Content-Type": []string{"text/html"}}, Body: io.NopCloser(bytes.NewBufferString(`<a href="/play" hx-post="/play"><script src="/assets/ws.js"></script></a>`))}
	if err := prefixRuntimeResponse("/w/world-1")(response); err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	text := string(body)
	for _, expected := range []string{`href="/w/world-1/play"`, `hx-post="/w/world-1/play"`, `src="/w/world-1/assets/ws.js"`} {
		if !strings.Contains(text, expected) {
			t.Fatalf("%q missing from %q", expected, text)
		}
	}
}

func TestBundledBloopSeedCompiles(t *testing.T) {
	root := seedRootFromEnvironment()
	collection, err := loadSeedCollection(root, "bloop")
	if err != nil {
		t.Fatal(err)
	}
	palette, err := os.ReadFile(filepath.Join(filepath.Dir(root), "colors", "colors.json"))
	if err != nil {
		t.Fatal(err)
	}
	manifest := WorldManifest{Name: "Bloop", Entry: WorldLocation{Stage: "tutorial1:0-0", Y: 3, X: 3}, Teams: []WorldTeam{{ID: "sky-blue", Label: "Sky Blue", Color: "sky-blue", Spawn: WorldLocation{Stage: "tutorial1:0-0", Y: 3, X: 3}}}, DefaultTeam: "sky-blue", MaxPlayers: 400, Lifecycle: LifecyclePersistent}
	resources := []WorldResource{{Kind: "collection", Key: "source", Content: mustJSON(t, collection)}, {Kind: "manifest", Key: "world", Content: mustJSON(t, manifest)}, {Kind: "palette", Key: "colors", Content: palette}}
	compiled, err := compileWorldResources(resources, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(compiled.Artifact) == 0 {
		t.Fatal("seed compilation produced no artifact")
	}
	archive, err := zip.NewReader(bytes.NewReader(compiled.Artifact), int64(len(compiled.Artifact)))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range archive.File {
		if filepath.Dir(entry.Name) != "images" {
			continue
		}
		reader, err := entry.Open()
		if err != nil {
			t.Fatal(err)
		}
		img, err := png.Decode(reader)
		reader.Close()
		if err != nil {
			t.Fatalf("decode %s: %v", entry.Name, err)
		}
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				_, _, _, alpha := img.At(x, y).RGBA()
				if alpha != 0xffff {
					t.Fatalf("%s contains transparent pixel at %d,%d", entry.Name, x, y)
				}
			}
		}
	}
}

type fakeRuntimeLauncher struct {
	starts int
	stops  int
}

func (f *fakeRuntimeLauncher) Start(_ context.Context, world *WorldDocument, release *WorldRelease, _ string) (*ManagedRuntime, error) {
	f.starts++
	return &ManagedRuntime{RuntimeInfo: RuntimeInfo{WorldID: world.ID, ReleaseID: release.ID, State: RuntimeRunning}, done: make(chan error, 1), cancel: func() {}}, nil
}
func (f *fakeRuntimeLauncher) Stop(_ context.Context, _ *ManagedRuntime) error { f.stops++; return nil }
