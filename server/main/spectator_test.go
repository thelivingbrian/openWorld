package main

import (
	"bytes"
	"strings"
	"sync/atomic"
	"testing"
)

func TestAdminWatchTemplateShowsModeSpecificControls(t *testing.T) {
	var playerPage bytes.Buffer
	if err := tmpl.ExecuteTemplate(&playerPage, "admin-watch", AdminWatchPageData{Mode: "player", Target: "Ada", SocketPath: "/socket"}); err != nil {
		t.Fatalf("render player watch page: %v", err)
	}
	if !strings.Contains(playerPage.String(), "Show HUD") || strings.Contains(playerPage.String(), "Next camera") {
		t.Fatalf("player watch controls were incorrect")
	}

	var stagePage bytes.Buffer
	if err := tmpl.ExecuteTemplate(&stagePage, "admin-watch", AdminWatchPageData{Mode: "stage", Target: "town", SocketPath: "/socket"}); err != nil {
		t.Fatalf("render stage watch page: %v", err)
	}
	if !strings.Contains(stagePage.String(), "Next camera") || strings.Contains(stagePage.String(), "Show HUD") {
		t.Fatalf("stage watch controls were incorrect")
	}
}

func TestFilterPlayerWatchUpdateVisibility(t *testing.T) {
	canvas := `[~ id="Lt1" y="2" x="3" class="salmon"]`
	hud := `<div id="info">secret hud</div>`
	menu := `<div id="modal_background">secret menu</div>`

	if got := string(filterPlayerWatchUpdate([]byte(canvas+hud), false, false)); got != canvas {
		t.Fatalf("canvas-only filter = %q, want %q", got, canvas)
	}
	if got := string(filterPlayerWatchUpdate([]byte(canvas+hud), true, false)); !strings.Contains(got, "secret hud") {
		t.Fatalf("HUD-enabled filter omitted HUD: %q", got)
	}
	if got := string(filterPlayerWatchUpdate([]byte(canvas+menu), false, false)); strings.Contains(got, "secret menu") {
		t.Fatalf("menu-disabled filter leaked menu: %q", got)
	}
	if got := string(filterPlayerWatchUpdate([]byte(canvas+menu), false, true)); !strings.Contains(got, "secret menu") {
		t.Fatalf("menu-enabled filter omitted menu: %q", got)
	}
}

func TestPlayerWatcherReceivesCanvasFromOutboundSocketPath(t *testing.T) {
	player := &Player{conn: &MockConn{}, watchers: make(map[*playerWatcher]struct{})}
	watcher := &playerWatcher{updates: make(chan []byte, 1)}
	player.addWatcher(watcher)

	update := []byte(`[~ id="Lp1" y="1" x="1" class="sky-blue"]<div id="info">private</div>`)
	if err := sendUpdate(player, update); err != nil {
		t.Fatalf("sendUpdate returned error: %v", err)
	}

	got := <-watcher.updates
	if !bytes.Contains(got, []byte(`id="Lp1"`)) || bytes.Contains(got, []byte("private")) {
		t.Fatalf("watcher update was not canvas-only: %q", got)
	}
}

func TestPlayerLogoutReplacesQueuedUpdatesWithDarkState(t *testing.T) {
	player := &Player{username: "Ada", watchers: make(map[*playerWatcher]struct{})}
	watcher := &playerWatcher{updates: make(chan []byte, 2)}
	watcher.updates <- []byte("stale")
	player.addWatcher(watcher)

	player.notifyWatchersPlayerLoggedOut()
	got := string(<-watcher.updates)
	if !strings.Contains(got, "watch-state-offline") || !strings.Contains(got, "Ada has logged out") {
		t.Fatalf("logout state = %q", got)
	}
}

func TestNonBlockingSpectatorCameraRequestsResyncWhenFull(t *testing.T) {
	updates := make(chan []byte, 1)
	updates <- []byte("occupied")
	var resync atomic.Bool
	camera := &Camera{outgoing: updates, dropUpdatesWhenBusy: true, resyncRequired: &resync}

	camera.send([]byte("new"))
	if !resync.Load() {
		t.Fatal("full spectator camera queue did not request a resync")
	}
}

func TestPlayerWatchSnapshotIncludesHighlights(t *testing.T) {
	stage := makeSpectatorTestStage(16, 16)
	updates := make(chan []byte, 300)
	player := &Player{
		username: "Ada",
		actions:  createDefaultActions(),
		updates:  updates,
		camera:   newCamera(updates),
	}
	player.tile = stage.tiles[8][8]
	player.actions.spaceHighlights[stage.tiles[7][8]] = true
	player.camera.setView(8, 8, stage)

	snapshot := string(player.watchSnapshot(false, false))
	if !strings.Contains(snapshot, `id="set"`) {
		t.Fatal("snapshot omitted camera origin")
	}
	if !strings.Contains(snapshot, `id="Lt1" y="7" x="8" class="trsp50 salmon"`) {
		t.Fatal("snapshot omitted active player highlight")
	}
}

func TestStageActivityOnlyCountsVisibleUpdates(t *testing.T) {
	stage := makeSpectatorTestStage(32, 32)
	updates := make(chan []byte, 300)
	camera := &Camera{height: 16, width: 16, outgoing: updates}
	camera.setView(0, 0, stage)

	if !stageUpdateIsVisible([]byte(`[~ id="Lp1" y="2" x="2" class="red"]`), camera) {
		t.Fatal("visible stage update was treated as idle")
	}
	if stageUpdateIsVisible([]byte(`[~ id="Lp1" y="30" x="30" class="red"]`), camera) {
		t.Fatal("off-camera stage update reset activity")
	}
}

func makeSpectatorTestStage(height, width int) *Stage {
	tiles := make([][]Material, height)
	for y := range tiles {
		tiles[y] = make([]Material, width)
	}
	return createStageFromArea(Area{Name: "spectator-test", Tiles: tiles})
}
