package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type Player struct {
	id       string
	username string
	team     string
	//trim                     string
	icon                     string
	viewLock                 sync.Mutex
	world                    *World
	stage                    *Stage // Shouldn't exist except for tile?
	stageLock                sync.Mutex
	tile                     *Tile
	tileLock                 sync.Mutex
	updates                  chan []byte
	clearUpdateBuffer        chan struct{}
	sessionTimeOutViolations atomic.Int32
	conn                     WebsocketConnection
	connLock                 sync.RWMutex
	tangible                 bool
	tangibilityLock          sync.Mutex
	health                   int
	healthLock               sync.Mutex
	money                    int
	moneyLock                sync.Mutex
	killCount                int
	killCountLock            sync.Mutex
	deathCount               int
	deathCountLock           sync.Mutex
	goalsScored              int
	goalsScoredLock          sync.Mutex
	killstreak               int
	streakLock               sync.Mutex
	actions                  *Actions
	menues                   map[string]Menu
	playerStages             map[string]*Stage
	pStageMutex              sync.Mutex
	hatList                  SyncHatList
}

type SyncHatList struct {
	sync.Mutex
	HatList
}

type HatList struct {
	Hats    []Hat `bson:"hats"`
	Current *int  `bson:"current"`
}

type Hat struct {
	Name           string    `bson:"name"`
	Trim           string    `bson:"trim"`
	ToggleDisabled bool      `bson:"toggleDisabled"`
	UnlockedAt     time.Time `bson:"unlockedAt"`
}

var EVERY_HAT_TO_TRIM map[string]string = map[string]string{
	"score-1-goal":    "black-b med",
	"score-1000-goal": "black-b thick",
	"most-dangerous":  "red-b med",
	"richest":         "green-b med",
	"puzzle-solve-1":  "white-b med",
	"contributor":     "gold-b thick",
}

func (hatList *SyncHatList) addByName(hatName string) *Hat {
	hatList.Lock()
	defer hatList.Unlock()
	for i := range hatList.Hats {
		if hatList.Hats[i].Name == hatName {
			return nil
		}
	}
	trim, ok := EVERY_HAT_TO_TRIM[hatName]
	if !ok {
		fmt.Println("INVALID HATNAME: ", hatName)
		return nil
	}
	newHat := Hat{Name: hatName, Trim: trim, ToggleDisabled: false, UnlockedAt: time.Now()}
	hatList.Hats = append(hatList.Hats, newHat)
	hatCount := len(hatList.Hats) - 1
	hatList.Current = &hatCount
	return &hatList.Hats[hatCount]
}

func (hatList *SyncHatList) peek() *Hat {
	hatList.Lock()
	defer hatList.Unlock()
	if hatList.Current == nil {
		return nil
	}
	return &hatList.Hats[*hatList.Current]
}

func (hatList *SyncHatList) next() *Hat {
	hatList.Lock()
	defer hatList.Unlock()
	hatCount := len(hatList.Hats)
	if hatCount == 0 {
		return nil
	}
	if hatList.Current == nil {
		current := 0
		hatList.Current = &current
		return &hatList.Hats[0]
	}
	if *hatList.Current == hatCount-1 {
		hatList.Current = nil
		return nil
	}
	*hatList.Current++
	return &hatList.Hats[*hatList.Current]
}

func (hatList *SyncHatList) nextValid() *Hat {
	for {
		hat := hatList.next()
		if hat == nil {
			return nil
		}
		if !hat.ToggleDisabled {
			return hat
		}
	}
}

func (hatList *SyncHatList) indexSync() *int {
	hatList.Lock()
	defer hatList.Unlock()
	return hatList.Current
}

func (hatList *SyncHatList) currentTrim() string {
	hatList.Lock()
	defer hatList.Unlock()
	if hatList.Current == nil {
		return ""
	}
	return hatList.Hats[*hatList.Current].Trim
}

// func (hatList *HatList) last() *Hat {
// 	hatList.Lock()
// 	defer hatList.Unlock()
// 	hatCount := len(hatList.hats)
// 	if hatCount == 0 {
// 		return nil
// 	}
// 	hatCount--
// 	hatList.current = &hatCount
// 	return &hatList.hats[hatCount]
// }

func (player *Player) addHatByName(hatName string) {
	hat := player.hatList.addByName(hatName)
	if hat == nil {
		return
	}
	player.world.db.addHatToPlayer(player.username, *hat)
	player.updateInformation()
	return
}

func (player *Player) cycleHats() {
	player.hatList.nextValid()
	player.updateInformation()
	tile := player.getTileSync()
	tile.stage.updateAllExcept(playerBox(tile), player)
	return
}

// Save Hat and event seperately for potential to reconsile in event of error

type WebsocketConnection interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
	SetWriteDeadline(t time.Time) error
}

////////////////////////////////////////////////////////////
//   Movement

func (player *Player) moveNorth() {
	player.move(-1, 0)
}

func (player *Player) moveNorthBoost() {
	player.moveBoost(-1, 0)
}

func (player *Player) moveSouth() {
	player.move(1, 0)
}

func (player *Player) moveSouthBoost() {
	player.moveBoost(1, 0)
}

func (player *Player) moveEast() {
	player.move(0, 1)
}

func (player *Player) moveEastBoost() {
	player.moveBoost(0, 1)
}

func (player *Player) moveWest() {
	player.move(0, -1)
}

func (player *Player) moveWestBoost() {
	player.moveBoost(0, -1)
}

func (player *Player) move(yOffset int, xOffset int) {
	sourceTile := player.getTileSync()
	destTile := getRelativeTile(sourceTile, yOffset, xOffset, player)
	player.push(destTile, nil, yOffset, xOffset)
	if walkable(destTile) {
		transferPlayer(player, sourceTile, destTile)
	}
}

func (player *Player) moveBoost(yOffset int, xOffset int) {
	if player.useBoost() {
		player.pushUnder(2*yOffset, 2*xOffset)
		player.move(2*yOffset, 2*xOffset)
	} else {
		// always push under ?
		player.move(yOffset, xOffset)
	}
}

func (player *Player) applyTeleport(teleport *Teleport) {
	stage := getStageFromStageName(player, teleport.destStage)
	if !validCoordinate(teleport.destY, teleport.destX, stage.tiles) {
		log.Fatal("Fatal: Invalid coords from teleport: ", teleport.destStage, teleport.destY, teleport.destX)
	}
	// Is using getTileSync a risk with the menu teleport authorizer?
	transferPlayer(player, player.getTileSync(), stage.tiles[teleport.destY][teleport.destX])
}

// Atomic Transfers
func transferPlayer(p *Player, source, dest *Tile) {
	if source.stage == dest.stage {
		if transferPlayerWithinStage(p, source, dest) {
			updateOthersAfterMovement(p, dest, source)
			updatePlayerAfterMovement(p, dest, source)
		}
	} else {
		if transferPlayerAcrossStages(p, source, dest) {
			spawnItemsFor(p, dest.stage)
			updateOthersAfterMovement(p, dest, source)
			updatePlayerAfterStageChange(p)
		}
	}
}

func transferPlayerWithinStage(p *Player, source, dest *Tile) bool {
	p.tileLock.Lock()
	defer p.tileLock.Unlock()

	if !tryRemovePlayer(source, p) {
		return false
	}

	dest.addLockedPlayertoTile(p)
	return true
}

func transferPlayerAcrossStages(p *Player, source, dest *Tile) bool {
	p.stageLock.Lock()
	defer p.stageLock.Unlock()
	p.tileLock.Lock()
	defer p.tileLock.Unlock()

	if !tryRemovePlayer(source, p) {
		return false
	}

	p.stage.removeLockedPlayerById(p.id)
	p.stage = dest.stage

	dest.stage.addLockedPlayer(p)
	dest.addLockedPlayertoTile(p)
	return true
}

////////////////////////////////////////////////////////////
//   Pushing

func (p *Player) push(tile *Tile, incoming *Interactable, yOff, xOff int) bool { // Returns if given interacable successfully pushed
	// Do not nil check incoming interactable here.
	// incoming = nil is valid and will continue a push chain
	// e.g. by taking this tiles interactable and pushing it forward
	if tile == nil {
		return false
	}

	if hasTeleport(tile) {
		return p.pushTeleport(tile, incoming, yOff, xOff)
	}

	ownLock := tile.interactableMutex.TryLock()
	if !ownLock {
		return false // Tile is already locked by another operation
	}
	defer tile.interactableMutex.Unlock()

	if tile.interactable == nil {
		return replaceNilInteractable(tile, incoming)
	}

	if tile.interactable.React(incoming, p, tile, yOff, xOff) {
		return true
	}

	if tile.interactable.pushable {
		nextTile := getRelativeTile(tile, yOff, xOff, p)
		if nextTile != nil {
			if p.push(nextTile, tile.interactable, yOff, xOff) {
				swapInteractableAndUpdate(tile, incoming)
				return true
			}
		}
	}
	return false
}

func (p *Player) pushUnder(yOffset int, xOffset int) {
	currentTile := p.getTileSync()
	if currentTile != nil && currentTile.interactable != nil {
		p.push(currentTile, nil, yOffset, xOffset)
	}
}

func (p *Player) pushTeleport(tile *Tile, incoming *Interactable, yOff, xOff int) bool {
	if tile.teleport.rejectInteractable {
		return false
	}
	if canBeTeleported(incoming) {
		stage := getStageFromStageName(p, tile.teleport.destStage)
		if !validCoordinate(tile.teleport.destY+yOff, tile.teleport.destX+xOff, stage.tiles) {
			return false
		}
		return p.push(stage.tiles[tile.teleport.destY+yOff][tile.teleport.destX+xOff], incoming, yOff, xOff)
	}
	return false
}

func replaceNilInteractable(tile *Tile, incoming *Interactable) bool {
	if tile.interactable != nil {
		return false
	}
	if !tile.material.Walkable { // Prevents lock contention from using Walkable()
		return false
	}
	swapInteractableAndUpdate(tile, incoming)

	return true
}

func swapInteractableAndUpdate(tile *Tile, incoming *Interactable) {
	experiencedChange := tile.interactable != incoming
	tile.interactable = incoming
	if experiencedChange {
		tile.stage.updateAll(interactableBox(tile))
	}
}

func hasTeleport(tile *Tile) bool {
	if tile == nil || tile.teleport == nil {
		return false
	}
	return true
}

func canBeTeleported(interactable *Interactable) bool {
	if interactable == nil {
		return false
	}
	return !interactable.rejectTeleport
}

///////////////////////////////////////////////////////////////////////
// Death

func handleDeath(player *Player) {
	player.getTileSync().addMoneyAndNotifyAllExcept(max(halveMoneyOf(player), 10), player)
	removeFromTileAndStage(player) // After this should be impossible for any transfer to succeed
	player.incrementDeathCount()
	player.setHealth(150)
	player.setKillStreak(0)
	player.actions = createDefaultActions() // problematic

	stage := player.fetchStageSync(infirmaryStagenameForPlayer(player))
	player.setStage(stage)
	player.updateRecordOnDeath(stage.tiles[2][2])
	respawnOnStage(player, stage)
}

func halveMoneyOf(player *Player) int {
	currentMoney := player.getMoneySync()
	newValue := currentMoney / 2
	player.setMoney(newValue)
	return newValue
}

func respawnOnStage(player *Player, stage *Stage) {
	player.tangibilityLock.Lock()
	defer player.tangibilityLock.Unlock()
	if !player.tangible {
		return
	}

	placePlayerOnStageAt(player, stage, 2, 2)
	player.updateInformation()
}

func removeFromTileAndStage(player *Player) {
	player.stageLock.Lock()
	defer player.stageLock.Unlock()
	player.tileLock.Lock()
	defer player.tileLock.Unlock()
	if player.stage == nil || player.tile == nil {
		return
	}
	player.tile.removePlayerAndNotifyOthers(player)
	player.stage.removeLockedPlayerById(player.id)
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
//   Updates

func (player *Player) sendUpdates() {
	var buffer bytes.Buffer
	const maxBufferSize = 256 * 1024

	shouldSendUpdates := true
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case update, ok := <-player.updates:
			if !ok {
				fmt.Println("Player:", player.username, "- update channel closed")
				return
			}
			if !shouldSendUpdates {
				continue
			}

			if buffer.Len()+len(update) < maxBufferSize {
				// Accumulate the update in the buffer.
				buffer.Write(update)
			} else {
				// Has not occurred - nice to check anyway ?
				fmt.Printf("Player: %s - buffer exceeded %d bytes, wiping buffer\n", player.username, maxBufferSize)
				buffer.Reset()
			}
		case <-player.clearUpdateBuffer:
			buffer.Reset()
		case <-ticker.C:
			if !shouldSendUpdates || buffer.Len() == 0 {
				continue
			}
			// Every 25ms, if there's anything in the buffer, send it.
			err := sendUpdate(player, buffer.Bytes())
			if err != nil {
				//fmt.Println("Error - Stopping furture sends: ", err)
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
		//   fmt.Println("WARN: Attempted to serve update to expired connection.")
		return errors.New("connection is expired")
	}

	err := player.conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
	if err != nil {
		fmt.Println("Failed to set write deadline:", err)
		return err
	}
	err = player.conn.WriteMessage(websocket.TextMessage, update)
	if err != nil {
		//fmt.Printf("WARN: WriteMessage failed for player %s: %v\n", player.username, err)
		// Close connection if writes consistently fail
		if player.sessionTimeOutViolations.Add(1) >= 1 {
			return err
		}
	}

	return nil
}

// Updates - Enqueue
func updateOthersAfterMovement(player *Player, current, previous *Tile) {
	previous.stage.updateAllExcept(playerBox(previous), player)
	current.stage.updateAllExcept(playerBox(current), player)
}

func updatePlayerAfterMovement(player *Player, current, previous *Tile) {
	impactedHighlights := player.updateSpaceHighlights()

	playerIcon := playerBoxSpecifc(current.y, current.x, player.getIconSync())

	previousBoxes := ""
	if previous != nil && previous.stage == player.getStageSync() {
		previousBoxes += playerBox(previous)
	}

	player.updates <- []byte(highlightBoxesForPlayer(player, impactedHighlights) + previousBoxes + playerIcon)
}

func updatePlayerAfterStageChange(p *Player) {
	p.setSpaceHighlights()
	updateScreenFromScratch(p)
	p.updateRecord() // too much?
}

func updateScreenFromScratch(player *Player) {
	// player.clearUpdateBuffer <- struct{}{}
	// clearChannel(player.updates)
	player.updates <- htmlFromPlayer(player)
}

func clearChannel(ch chan []byte) {
	for {
		select {
		case <-ch: // Read from the channel
			// Do nothing, just drain
		default: // Exit when the channel is empty
			return
		}
	}
}

var (
	// Regular expression for *[color]
	wordRegex = regexp.MustCompile(`\*\[(.+?)\]`)

	// Regular expression for @[phrase|color]
	phraseColorRegex = regexp.MustCompile(`@\[(.+?)\|(.+?)\]`)
	// Regular expression for @[phrase|---]
	teamColorWildRegex = regexp.MustCompile(`@\[(.*?)\|---\]`)
)

func processStringForColors(input string) string {
	//  Replace matches with <span class="color-t">color</span>
	input = wordRegex.ReplaceAllString(input, `<strong class="$1-t">$1</strong>`)

	//  Replace matches with <span class="color-t">phrase</span>
	input = phraseColorRegex.ReplaceAllString(input, `<strong class="$2-t">$1</strong>`)

	return input
}

func (player *Player) updateBottomText(message string) {
	msg := fmt.Sprintf(`
			<div id="bottom_text">
				&nbsp;&nbsp;> %s
			</div>`, processStringForColors(message))
	updateOne(msg, player)
}

func (player *Player) updateInformation() {
	icon := player.setIcon()
	tile := player.getTileSync()
	updateOne(divPlayerInformation(player)+playerBoxSpecifc(tile.y, tile.x, icon), player)
}

// chan Update

func updateOne(update string, player *Player) {
	player.updates <- []byte(update)
}

func (p *Player) trySend(msg []byte) {
	p.updates <- msg
}

// Database update

func (player *Player) updateRecord() {
	currentTile := player.getTileSync()
	go player.world.db.updateRecordForPlayer(player, currentTile)
}

func (player *Player) updateRecordOnDeath(respawnTile *Tile) {
	go player.world.db.updateRecordForPlayer(player, respawnTile)
}

/////////////////////////////////////////////////////////////
// Stages

func getStageFromStageName(player *Player, stagename string) *Stage {
	stage := player.fetchStageSync(stagename)
	if stage == nil {
		fmt.Println("WARNING: Fetching default stage instead of: " + stagename)
		stage = player.fetchStageSync("clinic")
		if stage == nil {
			panic("Default stage not found")
		}
	}

	return stage
}

func (player *Player) fetchStageSync(stagename string) *Stage {
	player.world.wStageMutex.Lock()
	defer player.world.wStageMutex.Unlock()
	stage, ok := player.world.worldStages[stagename]
	if ok && stage != nil {
		return stage
	}
	// stagename + team || stagename + rand

	player.pStageMutex.Lock()
	defer player.pStageMutex.Unlock()
	stage, ok = player.playerStages[stagename]
	if ok && stage != nil {
		return stage
	}

	area, success := areaFromName(stagename)
	if !success {
		//panic("ERROR! invalid stage with no area: " + stagename)
		return nil
	}

	stage = createStageFromArea(area) // can create empty stage
	if area.LoadStrategy == "" {
		player.world.worldStages[stagename] = stage
	}
	if area.LoadStrategy == "Personal" {
		player.playerStages[stagename] = stage
	}
	if area.LoadStrategy == "Individual" {
		// no-op : stage will load fresh each time
	}

	return stage
}

/////////////////////////////////////////////////////////////
// Observers

// Does not handle death
func (player *Player) setHealth(n int) {
	player.healthLock.Lock()
	defer player.healthLock.Unlock()
	player.health = n
}

func (player *Player) getHealthSync() int {
	player.healthLock.Lock()
	defer player.healthLock.Unlock()
	return player.health
}

// Icon Observer, note that health can not be locked
func (player *Player) setIcon() string {
	player.viewLock.Lock()
	defer player.viewLock.Unlock()
	player.healthLock.Lock()
	defer player.healthLock.Unlock()
	if player.health <= 50 {
		player.icon = "dim-" + player.team + " " + player.hatList.currentTrim() + " r0"
		return player.icon
	} else {
		player.icon = player.team + " " + player.hatList.currentTrim() + " r0"
		return player.icon
	}
}

func (player *Player) getIconSync() string {
	player.viewLock.Lock()
	defer player.viewLock.Unlock()
	return player.icon
}

func (player *Player) getTeamNameSync() string {
	player.viewLock.Lock()
	defer player.viewLock.Unlock()
	return player.team
}

// Stage observer, also sets name.
func (player *Player) setStage(stage *Stage) {
	player.stageLock.Lock()
	defer player.stageLock.Unlock()
	player.stage = stage
}

func (player *Player) getStageNameSync() string {
	player.stageLock.Lock()
	defer player.stageLock.Unlock()
	return player.stage.name
}

func (player *Player) getStageSync() *Stage {
	player.stageLock.Lock()
	defer player.stageLock.Unlock()
	return player.stage
}

func (player *Player) getTileSync() *Tile {
	player.tileLock.Lock()
	defer player.tileLock.Unlock()
	return player.tile
}

// Money observer, All Money changes should go through here
func (player *Player) setMoney(n int) {
	player.moneyLock.Lock()
	defer player.moneyLock.Unlock()
	player.money = n
	updateOne(divPlayerInformation(player), player)
}

func (player *Player) getMoneySync() int {
	player.moneyLock.Lock()
	defer player.moneyLock.Unlock()
	return player.money
}

// Streak observer, All streak changes should go through here
func (player *Player) setKillStreak(n int) {
	player.streakLock.Lock()
	defer player.streakLock.Unlock()
	player.killstreak = n
}

func (player *Player) setKillStreakAndUpdate(n int) {
	player.setKillStreak(n)
	player.world.leaderBoard.mostDangerous.Update(player)
	updateOne(divPlayerInformation(player), player)
}

func (player *Player) getKillStreakSync() int {
	player.streakLock.Lock()
	defer player.streakLock.Unlock()
	return player.killstreak
}

func (player *Player) incrementKillStreak() {
	newStreak := player.getKillStreakSync() + 1
	player.setKillStreakAndUpdate(newStreak)
}

func (player *Player) getKillCountSync() int {
	player.killCountLock.Lock()
	defer player.killCountLock.Unlock()
	return player.killCount
}

// killCount Observer - no direct set
func (player *Player) incrementKillCount() {
	player.killCountLock.Lock()
	defer player.killCountLock.Unlock()
	player.killCount++
}

func (player *Player) getDeathCountSync() int {
	player.deathCountLock.Lock()
	defer player.deathCountLock.Unlock()
	return player.deathCount
}

// deathCount Observer - no direct set
func (player *Player) incrementDeathCount() {
	player.deathCountLock.Lock()
	defer player.deathCountLock.Unlock()
	player.deathCount++
}

// goals observer no direct set
func (player *Player) incrementGoalsScored() int {
	player.goalsScoredLock.Lock()
	defer player.goalsScoredLock.Unlock()
	// add trim if first ? nah
	player.goalsScored++
	return player.goalsScored
}

func (player *Player) getGoalsScored() int {
	player.goalsScoredLock.Lock()
	defer player.goalsScoredLock.Unlock()
	return player.goalsScored
}

// generally will trigger a logout
func (player *Player) closeConnectionSync() error {
	player.connLock.Lock()
	defer player.connLock.Unlock()
	if player.conn == nil {
		return errors.New("Player connection nil before attempted close.")
	}
	return player.conn.Close()
}
