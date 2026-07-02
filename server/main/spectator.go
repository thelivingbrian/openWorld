package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"math/rand/v2"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	spectatorQueueSize      = 1024
	stageCameraIdleDuration = 30 * time.Second
	spectatorPingInterval   = 20 * time.Second
	spectatorWriteTimeout   = 2 * time.Second
	spectatorMaxBatchSize   = 512 * 1024
)

var (
	quickSwapPattern      = regexp.MustCompile(`\[~[^\]]*\]`)
	quickSwapCoordPattern = regexp.MustCompile(`\[~\s+id="[^"]+"\s+y="(-?\d+)"\s+x="(-?\d+)"`)
)

type AdminWatchPageData struct {
	Mode       string
	Target     string
	SocketPath string
	ShowHUD    bool
	ShowMenu   bool
}

type playerWatcher struct {
	updates        chan []byte
	showHUD        bool
	showMenu       bool
	resyncRequired atomic.Bool
	offline        atomic.Bool
}

func (world *World) adminWatchPageHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := world.requireAdminIdentifier(w, r); !ok {
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	playerName := strings.TrimSpace(r.URL.Query().Get("player"))
	stageName := strings.TrimSpace(r.URL.Query().Get("stage"))
	mode, target := "", ""
	switch {
	case playerName != "" && stageName == "":
		if world.getPlayerByUsername(playerName) == nil {
			http.Error(w, "Player is no longer online", http.StatusNotFound)
			return
		}
		mode, target = "player", playerName
	case stageName != "" && playerName == "":
		if world.getAdminStageByName(stageName) == nil {
			http.Error(w, "Stage is no longer active", http.StatusNotFound)
			return
		}
		mode, target = "stage", stageName
	default:
		http.Error(w, "Choose exactly one player or stage to watch", http.StatusBadRequest)
		return
	}

	showHUD := watchOptionEnabled(r.URL.Query().Get("hud"))
	showMenu := watchOptionEnabled(r.URL.Query().Get("menu"))
	query := url.Values{mode: []string{target}}
	if showHUD {
		query.Set("hud", "1")
	}
	if showMenu {
		query.Set("menu", "1")
	}

	data := AdminWatchPageData{
		Mode:       mode,
		Target:     target,
		SocketPath: "/admin/watch/screen?" + query.Encode(),
		ShowHUD:    showHUD,
		ShowMenu:   showMenu,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "admin-watch", data); err != nil {
		logger.Error().Err(err).Msg("failed to render admin watch page")
	}
}

func (world *World) adminWatchSocketHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := world.requireAdminIdentifier(w, r); !ok {
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	playerName := strings.TrimSpace(r.URL.Query().Get("player"))
	stageName := strings.TrimSpace(r.URL.Query().Get("stage"))
	if (playerName == "") == (stageName == "") {
		http.Error(w, "Choose exactly one player or stage to watch", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error().Err(err).Msg("failed to upgrade admin watch socket")
		return
	}
	defer conn.Close()

	if playerName != "" {
		world.watchPlayerConnection(conn, playerName, watchOptionEnabled(r.URL.Query().Get("hud")), watchOptionEnabled(r.URL.Query().Get("menu")))
		return
	}
	world.watchStageConnection(conn, stageName)
}

func watchOptionEnabled(value string) bool {
	enabled, err := strconv.ParseBool(value)
	return err == nil && enabled
}

func (world *World) watchPlayerConnection(conn *websocket.Conn, username string, showHUD, showMenu bool) {
	player := world.getPlayerByUsername(username)
	if player == nil {
		watchOfflineConnection(conn, username)
		return
	}

	watcher := &playerWatcher{
		updates:  make(chan []byte, spectatorQueueSize),
		showHUD:  showHUD,
		showMenu: showMenu,
	}
	player.addWatcher(watcher)
	defer player.removeWatcher(watcher)

	if !writeSpectatorMessage(conn, player.watchSnapshot(showHUD, showMenu)) {
		return
	}

	disconnected := spectatorDisconnectChannel(conn)
	pingTicker := time.NewTicker(spectatorPingInterval)
	defer pingTicker.Stop()

	for {
		select {
		case first := <-watcher.updates:
			if watcher.resyncRequired.Swap(false) {
				drainSpectatorUpdates(watcher.updates)
				if !writeSpectatorMessage(conn, player.watchSnapshot(showHUD, showMenu)) {
					return
				}
				continue
			}
			payload := collectSpectatorUpdates(first, watcher.updates)
			if watcher.resyncRequired.Swap(false) {
				drainSpectatorUpdates(watcher.updates)
				payload = player.watchSnapshot(showHUD, showMenu)
			}
			if !writeSpectatorMessage(conn, payload) {
				return
			}
		case <-pingTicker.C:
			if watcher.resyncRequired.Swap(false) {
				drainSpectatorUpdates(watcher.updates)
				if !writeSpectatorMessage(conn, player.watchSnapshot(showHUD, showMenu)) {
					return
				}
				continue
			}
			if !writeSpectatorPing(conn) {
				return
			}
		case <-disconnected:
			return
		}
	}
}

func (world *World) watchStageConnection(conn *websocket.Conn, stageName string) {
	stage := world.getAdminStageByName(stageName)
	if stage == nil {
		writeSpectatorMessage(conn, watchUnavailableState("Stage is no longer active"))
		return
	}

	updates := make(chan []byte, spectatorQueueSize)
	var resyncRequired atomic.Bool
	camera := &Camera{
		height:              VIEW_HEIGHT,
		width:               VIEW_WIDTH,
		padding:             5,
		outgoing:            updates,
		dropUpdatesWhenBusy: true,
		resyncRequired:      &resyncRequired,
	}
	defer func() {
		if cameraHasView(camera) {
			camera.drop()
		}
	}()

	moveStageCamera(camera, stage)
	events, disconnected := spectatorInputChannels(conn)
	idleTimer := time.NewTimer(stageCameraIdleDuration)
	defer idleTimer.Stop()
	pingTicker := time.NewTicker(spectatorPingInterval)
	defer pingTicker.Stop()

	for {
		select {
		case first := <-updates:
			if resyncRequired.Swap(false) {
				drainSpectatorUpdates(updates)
				if !writeSpectatorMessage(conn, cameraSnapshot(camera)) {
					return
				}
				continue
			}
			payload, visibleActivity := collectStageUpdates(first, updates, camera)
			if resyncRequired.Swap(false) {
				drainSpectatorUpdates(updates)
				payload = cameraSnapshot(camera)
			}
			if visibleActivity {
				resetTimer(idleTimer, stageCameraIdleDuration)
			}
			if len(payload) > 0 && !writeSpectatorMessage(conn, payload) {
				return
			}
		case event, ok := <-events:
			if !ok {
				return
			}
			if event.EventName == "next-camera" {
				drainSpectatorUpdates(updates)
				moveStageCamera(camera, stage)
				resetTimer(idleTimer, stageCameraIdleDuration)
			}
		case <-idleTimer.C:
			drainSpectatorUpdates(updates)
			moveStageCamera(camera, stage)
			resetTimer(idleTimer, stageCameraIdleDuration)
		case <-pingTicker.C:
			if resyncRequired.Swap(false) {
				drainSpectatorUpdates(updates)
				if !writeSpectatorMessage(conn, cameraSnapshot(camera)) {
					return
				}
				continue
			}
			if !writeSpectatorPing(conn) {
				return
			}
		case <-disconnected:
			return
		}
	}
}

type spectatorEvent struct {
	EventName string `json:"eventname"`
}

func spectatorDisconnectChannel(conn *websocket.Conn) <-chan struct{} {
	disconnected := make(chan struct{})
	go func() {
		defer close(disconnected)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
	return disconnected
}

func spectatorInputChannels(conn *websocket.Conn) (<-chan spectatorEvent, <-chan struct{}) {
	events := make(chan spectatorEvent, 4)
	disconnected := make(chan struct{})
	go func() {
		defer close(events)
		defer close(disconnected)
		for {
			_, payload, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var event spectatorEvent
			if json.Unmarshal(payload, &event) == nil {
				select {
				case events <- event:
				default:
				}
			}
		}
	}()
	return events, disconnected
}

func watchOfflineConnection(conn *websocket.Conn, username string) {
	if !writeSpectatorMessage(conn, watchUnavailableState(fmt.Sprintf("%s has logged out", username))) {
		return
	}
	ticker := time.NewTicker(spectatorPingInterval)
	defer ticker.Stop()
	for range ticker.C {
		if !writeSpectatorPing(conn) {
			return
		}
	}
}

func (player *Player) addWatcher(watcher *playerWatcher) {
	player.watchersMutex.Lock()
	defer player.watchersMutex.Unlock()
	if player.watchers == nil {
		player.watchers = make(map[*playerWatcher]struct{})
	}
	player.watchers[watcher] = struct{}{}
}

func (player *Player) removeWatcher(watcher *playerWatcher) {
	player.watchersMutex.Lock()
	defer player.watchersMutex.Unlock()
	delete(player.watchers, watcher)
}

func (player *Player) publishWatchUpdate(update []byte) {
	player.rememberWatchMenu(update)
	player.watchersMutex.RLock()
	defer player.watchersMutex.RUnlock()
	for watcher := range player.watchers {
		if watcher.offline.Load() {
			continue
		}
		filtered := filterPlayerWatchUpdate(update, watcher.showHUD, watcher.showMenu)
		if len(filtered) == 0 {
			continue
		}
		select {
		case watcher.updates <- filtered:
		default:
			watcher.resyncRequired.Store(true)
		}
	}
}

func (player *Player) notifyWatchersPlayerLoggedOut() {
	message := watchUnavailableState(fmt.Sprintf("%s has logged out", player.username))
	player.watchersMutex.Lock()
	defer player.watchersMutex.Unlock()
	for watcher := range player.watchers {
		watcher.offline.Store(true)
		drainSpectatorUpdates(watcher.updates)
		select {
		case watcher.updates <- message:
		default:
			watcher.resyncRequired.Store(true)
		}
	}
}

func (player *Player) rememberWatchMenu(update []byte) {
	if !bytes.Contains(update, []byte(`id="modal_background"`)) {
		return
	}
	player.watchMenuMutex.Lock()
	defer player.watchMenuMutex.Unlock()
	player.watchMenuSnapshot = append(player.watchMenuSnapshot[:0], update...)
}

func (player *Player) watchSnapshot(showHUD, showMenu bool) []byte {
	if player.logoutInitiated.Load() || player.camera == nil || !cameraHasView(player.camera) {
		return watchUnavailableState(fmt.Sprintf("%s has logged out", player.username))
	}

	var out bytes.Buffer
	out.Write(cameraSnapshot(player.camera))
	tile := player.getTileSync()
	if tile != nil {
		topLeft, visibleTiles := cameraView(player.camera)
		if topLeft != nil {
			out.Write(highlightBoxesForPlayer(player, visibleTiles))
		}
	}
	if showHUD {
		out.WriteString(divPlayerInformation(player))
	}
	if showMenu {
		player.watchMenuMutex.RLock()
		out.Write(player.watchMenuSnapshot)
		player.watchMenuMutex.RUnlock()
	}
	return out.Bytes()
}

func filterPlayerWatchUpdate(update []byte, showHUD, showMenu bool) []byte {
	menuUpdate := containsAny(update, `id="modal_background"`, `id="modal_menu"`, `id="menulink_`)
	hudUpdate := containsAny(update, `id="info"`, `id="hearts"`, `id="streak"`, `id="boosts"`, `id="money"`, `id="power"`, `id="bottom_text"`)
	if (showMenu && menuUpdate) || (showHUD && hudUpdate) {
		return append([]byte(nil), update...)
	}
	return extractQuickSwaps(update)
}

func containsAny(update []byte, markers ...string) bool {
	for _, marker := range markers {
		if bytes.Contains(update, []byte(marker)) {
			return true
		}
	}
	return false
}

func extractQuickSwaps(update []byte) []byte {
	matches := quickSwapPattern.FindAll(update, -1)
	return bytes.Join(matches, nil)
}

func cameraSnapshot(camera *Camera) []byte {
	topLeft, tiles := cameraView(camera)
	if topLeft == nil {
		return nil
	}
	var out bytes.Buffer
	fmt.Fprintf(&out, `[~ id="set" y="%d" x="%d" class=""]`, topLeft.y, topLeft.x)
	for _, tile := range tiles {
		out.WriteString(swapsForTileNoHighlight(tile))
	}
	return out.Bytes()
}

func cameraView(camera *Camera) (*Tile, []*Tile) {
	camera.positionLock.Lock()
	defer camera.positionLock.Unlock()
	if camera.topLeft == nil || camera.topLeft.stage == nil {
		return nil, nil
	}
	topLeft := camera.topLeft
	stage := topLeft.stage
	tiles := getRegion(stage.tiles, Rect{
		topLeft.y,
		min(topLeft.y+camera.height-1, len(stage.tiles)-1),
		topLeft.x,
		min(topLeft.x+camera.width-1, len(stage.tiles[0])-1),
	})
	return topLeft, tiles
}

func cameraHasView(camera *Camera) bool {
	camera.positionLock.Lock()
	defer camera.positionLock.Unlock()
	return camera.topLeft != nil
}

func moveStageCamera(camera *Camera, stage *Stage) {
	previousY, previousX := -1, -1
	if topLeft, _ := cameraView(camera); topLeft != nil {
		previousY, previousX = topLeft.y, topLeft.x
		camera.drop()
	}

	y, x := randomStageCoordinate(stage)
	for range 8 {
		candidateY, candidateX := topLeft(len(stage.tiles), len(stage.tiles[0]), camera.height, camera.width, y, x)
		if candidateY != previousY || candidateX != previousX {
			break
		}
		y, x = randomStageCoordinate(stage)
	}
	camera.setView(y, x, stage)
}

func randomStageCoordinate(stage *Stage) (int, int) {
	return rand.IntN(len(stage.tiles)), rand.IntN(len(stage.tiles[0]))
}

func collectSpectatorUpdates(first []byte, updates <-chan []byte) []byte {
	var out bytes.Buffer
	out.Write(first)
	for out.Len() < spectatorMaxBatchSize {
		select {
		case update := <-updates:
			out.Write(update)
		default:
			return out.Bytes()
		}
	}
	return out.Bytes()
}

func collectStageUpdates(first []byte, updates <-chan []byte, camera *Camera) ([]byte, bool) {
	var out bytes.Buffer
	visible := stageUpdateIsVisible(first, camera)
	out.Write(extractQuickSwaps(first))
	for out.Len() < spectatorMaxBatchSize {
		select {
		case update := <-updates:
			visible = visible || stageUpdateIsVisible(update, camera)
			out.Write(extractQuickSwaps(update))
		default:
			return out.Bytes(), visible
		}
	}
	return out.Bytes(), visible
}

func stageUpdateIsVisible(update []byte, camera *Camera) bool {
	topLeft, _ := cameraView(camera)
	if topLeft == nil {
		return false
	}
	for _, match := range quickSwapCoordPattern.FindAllSubmatch(update, -1) {
		y, yErr := strconv.Atoi(string(match[1]))
		x, xErr := strconv.Atoi(string(match[2]))
		if yErr == nil && xErr == nil && y >= topLeft.y && y < topLeft.y+camera.height && x >= topLeft.x && x < topLeft.x+camera.width {
			return true
		}
	}
	return false
}

func drainSpectatorUpdates(updates <-chan []byte) {
	for {
		select {
		case <-updates:
		default:
			return
		}
	}
}

func resetTimer(timer *time.Timer, duration time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(duration)
}

func writeSpectatorMessage(conn *websocket.Conn, payload []byte) bool {
	if len(payload) == 0 {
		return true
	}
	if err := conn.SetWriteDeadline(time.Now().Add(spectatorWriteTimeout)); err != nil {
		return false
	}
	return conn.WriteMessage(websocket.TextMessage, payload) == nil
}

func writeSpectatorPing(conn *websocket.Conn) bool {
	if err := conn.SetWriteDeadline(time.Now().Add(spectatorWriteTimeout)); err != nil {
		return false
	}
	return conn.WriteMessage(websocket.PingMessage, nil) == nil
}

func watchUnavailableState(message string) []byte {
	return []byte(fmt.Sprintf(
		`<div id="watch-state" class="watch-state watch-state-offline" hx-swap-oob="true"><strong>%s</strong></div>`+
			`<div id="info" class="watch-hud" hx-swap-oob="true"></div>`+
			`<div id="bottom_text" class="watch-bottom-text" hx-swap-oob="true"></div>`+
			`<div id="modal_background" hx-swap-oob="true"></div>`,
		html.EscapeString(message),
	))
}

func (world *World) getAdminStageByName(name string) *Stage {
	world.wStageMutex.Lock()
	defer world.wStageMutex.Unlock()
	return world.worldStages[name]
}
