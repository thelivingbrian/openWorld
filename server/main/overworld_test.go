package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestResponsiveViewportIncludesSceneryAndReceivesFarEdgeUpdates(t *testing.T) {
	stage := makeSpectatorTestStage(64, 64)
	updates := make(chan []byte, 32)
	player := &Player{tile: stage.tiles[30][30], actions: createDefaultActions(), updates: updates, camera: newCamera(updates)}
	player.camera.setView(30, 30, stage)
	<-updates
	player.resizeViewport(12, 28)
	snapshot := string(<-updates)
	if !strings.Contains(snapshot, `id="view" y="12" x="28" class="64,64"`) {
		t.Fatal("missing rectangular viewport dimensions")
	}
	// The visible window is (24,16)..(35,43); scenery anchored just below it
	// must be present so tall objects do not pop into view at camera boundaries.
	if !strings.Contains(snapshot, `y="39" x="30"`) {
		t.Fatal("missing scenery margin")
	}
	<-updates // highlights
	stage.tiles[30][43].updateAll("far-edge")
	select {
	case data := <-updates:
		if string(data) != "far-edge" {
			t.Fatalf("unexpected update: %s", data)
		}
	default:
		t.Fatal("wide view missed a live update across a camera zone")
	}
	for _, dimensions := range [][2]int{{0, 16}, {16, 1000}, {-1, 16}, {16, 7}} {
		player.resizeViewport(dimensions[0], dimensions[1])
	}
	if player.camera.height != 12 || player.camera.width != 28 {
		t.Fatal("invalid viewport accepted")
	}
	player.camera.drop()
}

func TestShortViewportCameraDoesNotOscillate(t *testing.T) {
	boundary := axisAdjust(20, 16, 8, 48, 5)
	for i := 0; i < 10; i++ {
		next := axisAdjust(20, boundary, 8, 48, 5)
		if next != boundary {
			t.Fatal("stationary camera oscillates with short viewport")
		}
		boundary = next
	}
	if axisAdjust(47, 30, 12, 48, 5) != 36 {
		t.Fatal("camera must stop at world edge")
	}
}

func TestWillowRuntimePreservesAuthoredDoorsAndManifest(t *testing.T) {
	collection, err := loadSeedCollection("../../tools/main/data/collections", "bloop")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../tools/main/data/colors/colors.json")
	if err != nil {
		t.Fatal(err)
	}
	palette, err := decodeWorldPalette(data)
	if err != nil {
		t.Fatal(err)
	}
	files, err := compileCollectionFiles(collection, palette)
	if err != nil {
		t.Fatal(err)
	}
	var compiled []Area
	if err = json.Unmarshal(files["areas.json"], &compiled); err != nil {
		t.Fatal(err)
	}
	areasByName := map[string]Area{}
	for _, area := range compiled {
		areasByName[area.Name] = area
	}
	data, err = os.ReadFile("../../tools/main/data/collections/bloop/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest WorldManifest
	if err = json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if err = validateManifest(manifest, true); err != nil {
		t.Fatal(err)
	}
	if err = validateWorldLocation(manifest.Entry, areasByName); err != nil {
		t.Fatal(err)
	}
	if err = validateWorldRules(manifest, areasByName); err != nil {
		t.Fatal(err)
	}
	stage := createStageFromArea(areasByName["willow-reach:0-0"])
	for _, door := range [][2]int{{19, 13}, {13, 23}, {35, 39}} {
		tile := stage.tiles[door[0]][door[1]]
		if tile.teleport == nil || !tile.material.Walkable {
			t.Fatal("cottage entrance is not usable")
		}
		if !strings.Contains(tile.quickSwapTemplate, "scenery-cottage") {
			t.Fatal("transport marker erased authored cottage")
		}
	}
	for _, area := range compiled {
		for _, door := range area.Transports {
			if strings.HasPrefix(area.Name, "willow-") && !strings.HasPrefix(door.DestStage, "willow-") {
				t.Fatal("new world connects to legacy")
			}
		}
	}
	if stage.safe || !areasByName["willow-refuge:0-0"].Safe {
		t.Fatal("frontier/refuge safety inverted")
	}

}

func TestHUDRefreshKeepsCompactControlsAndEscapesName(t *testing.T) {
	p := &Player{username: "<b>unsafe</b>", actions: createDefaultActions()}
	p.health.Store(150)
	hud := divPlayerInformation(p)
	for _, part := range []string{`class="play-hud"`, `id="fullscreen-toggle"`, `id="hud-menu"`, `id="hearts"`, `&lt;b&gt;unsafe&lt;/b&gt;`} {
		if !strings.Contains(hud, part) {
			t.Fatalf("HUD refresh omitted %s", part)
		}
	}
}
