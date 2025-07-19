package main

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type Player struct {
	id                       string
	username                 string
	team                     string
	icon                     string
	hat                      string
	viewLock                 sync.Mutex
	world                    *World
	tile                     *Tile
	tileLock                 sync.Mutex
	updates                  chan []byte
	sessionTimeOutViolations atomic.Int32
	textUpdatesInFlight      atomic.Int32
	conn                     WebsocketConnection
	connLock                 sync.RWMutex
	tangible                 bool
	tangibilityLock          sync.Mutex
	actions                  *Actions
	playerStages             map[string]*Stage
	pStageMutex              sync.Mutex
	//hatList                  SyncHatList
	accomplishments SyncAccomplishmentList
	health          atomic.Int64
	money           atomic.Int64
	killstreak      atomic.Int64
	PlayerStats
	SyncMenuList
	camera *Camera
}

type PlayerStats struct {
	killCount      atomic.Int64
	killCountNpc   atomic.Int64
	deathCount     atomic.Int64
	goalsScored    atomic.Int64
	peakKillStreak atomic.Int64
	peakWealth     atomic.Int64
}

type Camera struct {
	height, width, padding int
	positionLock           sync.Mutex
	topLeft                *Tile

	// Only for B
	ref *atomic.Pointer[Camera]

	// is == player.updates
	outgoing chan []byte
}

func (camera *Camera) setView(posY, posX int, stage *Stage) []*Tile {
	y, x := topLeft(len(stage.tiles), len(stage.tiles[0]), camera.height, camera.width, posY, posX)
	region := getRegion(stage.tiles, Rect{y, y + camera.height, x, x + camera.width})
	camera.outgoing <- []byte(fmt.Sprintf(`[~ id="set" y="%d" x="%d" class=""]`, y, x))
	newRef := &atomic.Pointer[Camera]{}
	newRef.Store(camera)
	for _, tile := range region {
		attachCameraPtr(tile, newRef)
	}
	camera.positionLock.Lock()
	defer camera.positionLock.Unlock()
	camera.topLeft = region[0]
	return region
}

func (player *Player) tryTrack() {
	player.camera.track(player)
}

func (camera *Camera) track(character Character) {
	focus := character.getTileSync()
	stageH, stageW := len(focus.stage.tiles), len(focus.stage.tiles[0])
	camera.positionLock.Lock()
	defer camera.positionLock.Unlock()
	//oldTopLeft := camera.topLeft

	//newY, newX := topLeft(stageH, stageW, camera.height, camera.width, focus.y, focus.x)
	newY := axisAdjust(focus.y, camera.topLeft.y, camera.height, stageH, camera.padding)
	newX := axisAdjust(focus.x, camera.topLeft.x, camera.width, stageW, camera.padding)
	dy := camera.topLeft.y - newY
	dx := camera.topLeft.x - newX
	if dx == 0 && dy == 0 {
		return
	}

	//camera.topLeft = focus.stage.tiles[newY][newX]
	camera.outgoing <- []byte(fmt.Sprintf(`[~ id="shift" y="%d" x="%d" class=""]`, dy, dx))

	updateTiles(camera, newY, newX)
}

func updateTiles(camera *Camera, newY, newX int) {
	updateTilesA(camera, newY, newX)
}

func updateTilesA(camera *Camera, newY, newX int) {
	oldTopLeft := camera.topLeft
	camera.topLeft = oldTopLeft.stage.tiles[newY][newX]
	stage := oldTopLeft.stage
	stageH, stageW := len(oldTopLeft.stage.tiles), len(oldTopLeft.stage.tiles[0])

	oldY0, oldY1 := oldTopLeft.y, oldTopLeft.y+camera.height-1
	oldX0, oldX1 := oldTopLeft.x, oldTopLeft.x+camera.width-1
	newY0, newY1 := newY, newY+camera.height-1
	newX0, newX1 := newX, newX+camera.width-1

	// Tiles that dropped *out* of view.
	for y := oldY0; y <= oldY1; y++ {
		if y < 0 || y >= stageH {
			continue
		}
		for x := oldX0; x <= oldX1; x++ {
			if x < 0 || x >= stageW {
				continue
			}
			if y < newY0 || y > newY1 || x < newX0 || x > newX1 {
				removeCamera(stage.tiles[y][x], camera)
			}
		}
	}

	// Tiles that came *into* view.
	for y := newY0; y <= newY1; y++ {
		if y < 0 || y >= stageH {
			continue
		}
		for x := newX0; x <= newX1; x++ {
			if x < 0 || x >= stageW {
				continue
			}
			if y < oldY0 || y > oldY1 || x < oldX0 || x > oldX1 {
				attachCamera(stage.tiles[y][x], camera)
			}
		}
	}
}

func updateTilesB(camera *Camera, newY, newX int) {
	camera.topLeft = camera.topLeft.stage.tiles[newY][newX]

	// "remove" all of the old camera
	camera.ref.Store(nil)

	newRef := &atomic.Pointer[Camera]{}
	newRef.Store(camera)

	for y := newY; y < newY+camera.height; y++ {
		for x := newX; x < newX+camera.width; x++ {
			attachCameraPtr(camera.topLeft.stage.tiles[y][x], newRef)
		}
	}
}

///////////////
// A

func attachCamera(tile *Tile, cam *Camera) {
	addCamera(tile, cam)
	cam.outgoing <- []byte(swapsForTileNoHighlight(tile))
}

// Honestly unsurprising that this add/remove strategy would leave zombie cameras
func addCamera(tile *Tile, cam *Camera) {
	tile.camerasLock.Lock()
	defer tile.camerasLock.Unlock()
	tile.cameras[cam] = struct{}{}
}

func removeCamera(tile *Tile, cam *Camera) {
	tile.camerasLock.Lock()
	defer tile.camerasLock.Unlock()
	delete(tile.cameras, cam)
}

func (camera *Camera) drop() {
	camera.positionLock.Lock()
	defer camera.positionLock.Unlock()

	region := getRegion(camera.topLeft.stage.tiles, Rect{camera.topLeft.y, camera.topLeft.y + camera.height, camera.topLeft.x, camera.topLeft.x + camera.width})
	for _, tile := range region {
		removeCamera(tile, camera)
	}
	camera.topLeft = nil
}

/////////////////
// B

func attachCameraPtr(tile *Tile, camRef *atomic.Pointer[Camera]) {
	addCameraPtr(tile, camRef)
	camera := camRef.Load()
	if camera == nil {
		return
	}
	// filter ?
	camera.outgoing <- []byte(swapsForTileNoHighlight(tile))
}

func addCameraPtr(tile *Tile, camRef *atomic.Pointer[Camera]) {
	tile.camerasLock.Lock()
	defer tile.camerasLock.Unlock()
	tile.cameraPtrs[camRef] = struct{}{}
}

func topLeft(gridHeight, gridWidth, viewHeight, viewWidth, y, x int) (row, col int) {
	// clamp the requested window
	if viewHeight > gridHeight {
		viewHeight = gridHeight
	}
	if viewWidth > gridWidth {
		viewWidth = gridWidth
	}

	// Ideal top‑left: put (r,c) at ⌊n/2⌋, ⌊m/2⌋ inside the window.
	row = y - viewHeight/2
	col = x - viewWidth/2

	// Clamp to the valid range 0 … rows‑n and 0 … cols‑m.
	if row < 0 {
		row = 0
	}
	if col < 0 {
		col = 0
	}
	if row > gridHeight-viewHeight {
		row = gridHeight - viewHeight
	}
	if col > gridWidth-viewWidth {
		col = gridWidth - viewWidth
	}
	return
}

func axisAdjust(pos, oldBoundary, viewLength, gridLength, padding int) int {
	// Nothing to do if the view already covers the whole axis.
	if viewLength >= gridLength {
		return 0
	}

	lo := oldBoundary + padding                  // nearest allowed position inside view
	hi := oldBoundary + viewLength - padding - 1 // farthest allowed position inside view

	newBoundary := oldBoundary
	switch {
	case pos < lo: // too close to the top/left edge
		newBoundary -= lo - pos // shift up/left just enough
	case pos > hi: // too close to the bottom/right edge
		newBoundary += pos - hi // shift down/right just enough
	}

	// Clamp to grid.
	if newBoundary < 0 {
		newBoundary = 0
	}
	max := gridLength - viewLength
	if newBoundary > max {
		newBoundary = max
	}
	return newBoundary
}

////////////////////////////////////////////////////////////
//  Special Movement

func (player *Player) moveNorthBoost() {
	player.moveBoost(-1, 0)
}

func (player *Player) moveSouthBoost() {
	player.moveBoost(1, 0)
}

func (player *Player) moveEastBoost() {
	player.moveBoost(0, 1)
}

func (player *Player) moveWestBoost() {
	player.moveBoost(0, -1)
}

func (player *Player) moveBoost(yOffset int, xOffset int) {
	if player.useBoost() {
		move(player, 2*yOffset, 2*xOffset)
	} else {
		move(player, yOffset, xOffset)
	}
}

func (player *Player) applyTeleport(teleport *Teleport) {
	applyTeleport(player, teleport)
	sendSoundToPlayer(player, "teleport")
}

///////////////////////////////////////////////////////////////////////
// Death

func handleDeath(player *Player) {
	popAndDropMoney(player)
	removeFromTileAndStage(player) // After this should be impossible for any transfer to succeed
	player.incrementDeathCount()
	player.resetHealth()
	player.zeroKillStreak()
	player.setHat("")
	player.setIcon()
	player.actions = createDefaultActions() // problematic, -> setDefaultActions(player)

	stage := player.fetchStageSync(infirmaryStagenameForPlayer(player))
	player.updateRecordOnDeath(stage.tiles[2][2])
	respawnOnStage(player, stage)
}

func popAndDropMoney(player *Player) {
	tile := player.getTileSync()

	playerLostMoney := halveMoneyOf(player)
	moneyToAdd := max(playerLostMoney, 10)
	tile.addMoneyAndNotifyAllExcept(moneyToAdd, player)

	pop := soundTriggerByName("pop-death")
	tile.updateAll(pop)
}

func halveMoneyOf(player *Player) int {
	lost := player.halveMoney()
	updateOne(spanMoney(player.money.Load()), player)
	return int(lost)
}

func respawnOnStage(player *Player, stage *Stage) {
	player.tangibilityLock.Lock()
	defer player.tangibilityLock.Unlock()
	if !player.tangible {
		return
	}

	placePlayerOnStageAt(player, stage, 2, 2)
	//sendSoundToPlayer(player, soundTriggerByName("pop-death"))
	player.updatePlayerHud()
	player.updateBottomText("You have died.")
}

func removeFromTileAndStage(player *Player) {
	player.tileLock.Lock()
	defer player.tileLock.Unlock()
	if player.tile == nil {
		return
	}
	player.tile.removePlayerAndNotifyOthers(player)
	player.tile.stage.removeLockedPlayerById(player.id)
}

func infirmaryStagenameForPlayer(player *Player) string {
	team := player.getTeamNameSync()
	if team != "sky-blue" && team != "fuchsia" {
		return "clinic"
	}
	longitude := strconv.Itoa(rand.IntN(4))
	latitude := ""
	if team == "fuchsia" {
		latitude = "0"
	}
	if team == "sky-blue" {
		latitude = "3"
	}
	return fmt.Sprintf("infirmary:%s-%s", latitude, longitude)
}

////////////////////////////////////////////////////////////
//	Updates

func (player *Player) sendUpdates() {
	var buffer bytes.Buffer
	const maxBufferSize = 10 * 256 * 1024

	shouldSendUpdates := true
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case update, ok := <-player.updates:
			if !ok {
				logger.Info().Msg("Player:" + player.username + "- update channel closed")
				return
			}
			if !shouldSendUpdates {
				continue
			}

			if buffer.Len()+len(update) < maxBufferSize {
				// Accumulate the update in the buffer.
				buffer.Write(update)
			} else {
				logger.Warn().Msg(fmt.Sprintf("Player: %s - buffer exceeded %d bytes, wiping buffer\n", player.username, maxBufferSize))
				buffer.Reset()
			}
		case <-ticker.C:
			if !shouldSendUpdates || buffer.Len() == 0 {
				continue
			}
			// Every 25ms, if there's anything in the buffer, send it.
			err := sendUpdate(player, buffer.Bytes())
			if err != nil {
				//logger.Warn().Err(err).Msg("Error - Stopping furture sends: ")
				shouldSendUpdates = false
				player.closeConnectionSync()
			}

			buffer.Reset()
		}
	}
}

func sendUpdate(player *Player, update []byte) error {
	player.connLock.Lock()
	defer player.connLock.Unlock()
	if player.conn == nil {
		return errors.New("Connection is expired for: " + player.username)
	}

	err := player.conn.SetWriteDeadline(time.Now().Add(2000 * time.Millisecond))
	if err != nil {
		logger.Error().Err(err).Msg("Failed to set write deadline:")
		return err
	}
	err = player.conn.WriteMessage(websocket.TextMessage, update)
	if err != nil {
		// Technically can be any write error
		logger.Debug().Msg("Incrementing websocket session timeout violations for: " + player.username)
		if player.sessionTimeOutViolations.Add(1) >= 2 {
			return err
		}
	}

	return nil
}

// Updates - Enqueue
func updateOthersAfterMovement(player *Player, current, previous *Tile) {
	previous.updateAll(characterBox(previous))
	current.updateAll(characterBox(current))
}

func updateAllAfterMovement(current, previous *Tile) {
	previous.updateAll(characterBox(previous))
	current.updateAll(characterBox(current))
}

func updatePlayerAfterMovement(player *Player, current, previous *Tile) {
	impactedHighlights := player.updateSpaceHighlights()

	playerIcon := playerBoxSpecifc(current.y, current.x, player.getIconSync())

	previousBoxes := ""
	if previous != nil && previous.stage == current.stage {
		previousBoxes += characterBox(previous)
	}

	player.updates <- []byte(highlightBoxesForPlayer(player, impactedHighlights) + previousBoxes + playerIcon)
}

func updatePlayerAfterStageChange(p *Player) {
	p.setSpaceHighlights()
	updateEntireExistingScreen(p)
}

func updateEntireExistingScreen(player *Player) {
	player.updates <- entireScreenAsSwaps(player)
}

const bottomTextTemplate = `
<div id="bottom_text">
	&nbsp;&nbsp;> %s
</div>`

const bottomTextEmpty = `
<div id="bottom_text">
</div>`

func (player *Player) updateBottomText(message string) {
	msg := fmt.Sprintf(bottomTextTemplate, processStringForColors(message))
	updateOne(msg, player) // Potential to send on closed ?
	player.textUpdatesInFlight.Add(1)
	go tryClearBottomTextAfter(player, 5000)
}

func tryClearBottomTextAfter(player *Player, delay int) {
	time.Sleep(time.Millisecond * time.Duration(delay))
	if player.textUpdatesInFlight.Add(-1) == 0 {
		ownLock := player.tangibilityLock.TryLock()
		if !ownLock {
			return
		}
		defer player.tangibilityLock.Unlock()
		if !player.tangible {
			return
		}

		updateOne(bottomTextEmpty, player)
	}
}

func (player *Player) updatePlayerHud() {
	player.updatePlayerBox()
	updateOne(divPlayerInformation(player), player)
}

func (player *Player) updatePlayerBox() {
	icon := player.setIcon()
	tile := player.getTileSync()
	updateOne(playerBoxSpecifc(tile.y, tile.x, icon), player)
}

func updateIconForAll(player *Player) {
	player.setIcon()
	tile := player.getTileSync()
	tile.updateAll(characterBox(tile))
}

// Still an issue re updates and tangibility?
func updateIconForAllIfTangible(player *Player) {
	player.setIcon()
	ownLock := player.tangibilityLock.TryLock()
	if !ownLock {
		return
	}
	defer player.tangibilityLock.Unlock()
	if !player.tangible {
		return
	}
	tile := player.getTileSync()
	tile.updateAll(characterBox(tile))
}

func sendSoundToPlayer(player *Player, soundName string) {
	updateOne(soundTriggerByName(soundName), player)
}

func soundTriggerByName(soundName string) string {
	return fmt.Sprintf(`<div id="sound">%s</div>`, soundName)

}

// chan Update

func updateOne(update string, player *Player) {
	player.updates <- []byte(update)
}

// Database update

func (player *Player) updateRecord() {
	currentTile := player.getTileSync()
	go player.world.db.updateRecordForPlayer(player, currentTile)
}

func (player *Player) updateRecordOnDeath(respawnTile *Tile) {
	// vs just incrementing death count?
	go player.world.db.updateRecordForPlayer(player, respawnTile)
}

func (player *Player) updateRecordOnLogin() {
	go player.world.db.updateLoginForPlayer(player)
}
func (player *Player) updateRecordOnLogout() {
	currentTile := player.getTileSync()
	go player.world.db.updatePlayerRecordOnLogout(player, currentTile)
}

/////////////////////////////////////////////////////////////
// Stages

func getStageByNameOrGetDefault(player *Player, stagename string) *Stage {
	// Never returns nil - Used on login as protection against old records with nonexistant locations
	stage := player.fetchStageSync(stagename)
	if stage == nil {
		logger.Warn().Msg("WARNING: Fetching default stage instead of: " + stagename)
		stage = player.fetchStageSync("clinic")
		if stage == nil {
			panic("Default stage not found")
		}
	}
	return stage
}

/////////////////////////////////////////////////////////////
//  Hats

func (player *Player) addHatByName(hatName string, _ bool) {
	hat, ok := HAT_NAME_TO_TRIM[hatName]
	if !ok {
		return
	}
	player.setHat(hat)
	/*
		hat := player.hatList.addByName(hatName)
		if hat == nil {
			return
		}
		if persist {
			player.world.db.addHatToPlayer(player.username, *hat)
		}
	*/
	updateIconForAllIfTangible(player) // May not originate from click hence check tangible
}

func (player *Player) setHat(hat string) {
	player.viewLock.Lock()
	defer player.viewLock.Unlock()
	player.hat = hat
}

/*
func (player *Player) cycleHats() {
	player.hatList.nextValid()
	updateIconForAll(player)
}
*/

/////////////////////////////////////////////////////////////
//  Accomplishments

func (player *Player) addAccomplishmentByName(accomplishmentName string) {
	acc := player.accomplishments.addByName(accomplishmentName)
	if acc == nil {
		return
	}
	player.world.db.addAccomplishmentToPlayer(player.username, acc.Name, *acc)
}

/////////////////////////////////////////////////////////////
// Observers

func (player *Player) resetHealth() {
	player.health.Store(150)
}

// Icon Observer, note that health can not be locked
func (player *Player) setIcon() string {
	player.viewLock.Lock()
	defer player.viewLock.Unlock()
	if player.health.Load() <= 50 {
		player.icon = "dim-" + player.team + " " + player.hat + " r0"
		return player.icon
	} else {
		player.icon = player.team + " " + player.hat + " r0"
		return player.icon
	}
}

func (player *Player) getTeamNameSync() string {
	player.viewLock.Lock()
	defer player.viewLock.Unlock()
	return player.team
}

func (player *Player) halveMoney() int64 {
	// Currently lost and remaining money are equal but may want to split into two returns
	// e.g. for factor other than 1/2
	for {
		old := player.money.Load()
		new := old / 2
		if player.money.CompareAndSwap(old, new) {
			return new
		}
	}
}

func (player *Player) addMoneyAndUpdate(n int) {
	totalMoney := player.money.Add(int64(n))
	if SetMaxAtomic64IfGreater(&player.peakWealth, totalMoney) {
		checkMoneyAccomplishments(player, int(totalMoney))
	}

	// Track richest?

	updateOne(spanMoney(totalMoney), player)
}

func (player *Player) zeroKillStreak() {
	player.killstreak.Store(0)
	player.world.leaderBoard.mostDangerous.incoming <- PlayerStreakRecord{id: player.id, username: player.username, killstreak: 0, team: player.getTeamNameSync()}
}

func (player *Player) incrementKillStreak() int64 {
	currentKs := player.killstreak.Add(1)
	if SetMaxAtomic64IfGreater(&player.peakKillStreak, currentKs) {
		checkStreakAccomplishments(player, int(currentKs))
	}

	// Vs - character.updateHud ?
	updateStreakIfTangible(player) // initiator may not have initiatied via click -> check tangible needed

	player.world.leaderBoard.mostDangerous.incoming <- PlayerStreakRecord{id: player.id, username: player.username, killstreak: currentKs, team: player.getTeamNameSync()}
	return currentKs
}

func (player *Player) incrementKillCount() int64 {
	player.addAccomplishmentByName(defeatPlayer)
	return player.killCount.Add(1)
}

func (player *Player) incrementKillCountNpc() int64 {
	return player.killCountNpc.Add(1)
}

func (player *Player) incrementDeathCount() int64 {
	return player.deathCount.Add(1)
}

func (player *Player) incrementGoalsScored() int64 {
	return player.goalsScored.Add(1)
}

// generally will trigger a logout
func (player *Player) closeConnectionSync() error {
	player.connLock.Lock()
	defer player.connLock.Unlock()
	if player.conn == nil {
		return errors.New("player connection nil before attempted close")
	}
	return player.conn.Close()
}
