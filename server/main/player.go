package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
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
	trim                     string
	icon                     string
	viewLock                 sync.Mutex
	world                    *World
	stage                    *Stage
	stageLock                sync.Mutex
	tile                     *Tile
	tileLock                 sync.Mutex
	updates                  chan []byte
	clearUpdateBuffer        chan struct{}
	sessionTimeOutViolations atomic.Int32
	conn                     WebsocketConnection
	connLock                 sync.RWMutex
	tangible                 bool
	tangibilityLock          sync.Mutex // still has purpose?
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
}

type WebsocketConnection interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
	SetWriteDeadline(t time.Time) error
}

// Health observer, All Health changes should go through here
func (player *Player) setHealth(n int) {
	player.healthLock.Lock()
	player.health = n
	player.healthLock.Unlock()
	if n <= 0 {
		handleDeath(player)
		return
	}
	player.updateInformation()
}

func (player *Player) updateInformation() {
	player.setIcon()
	player.tileLock.Lock()
	tile := player.tile
	player.tileLock.Unlock()
	updateOne(divPlayerInformation(player)+playerBoxSpecifc(tile.y, tile.x, player.getIconSync()), player)
}

func (player *Player) getHealthSync() int {
	player.healthLock.Lock()
	defer player.healthLock.Unlock()
	return player.health
}

// Icon Observer, note that health can not be locked
func (player *Player) setIcon() {
	player.viewLock.Lock()
	defer player.viewLock.Unlock()
	player.healthLock.Lock()
	defer player.healthLock.Unlock()
	if player.health <= 50 {
		player.icon = "dim-" + player.team + " " + player.trim + " r0"
		return
	} else {
		player.icon = player.team + " " + player.trim + " r0"
		return
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
	// nil ?
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
	player.killstreak = n
	player.streakLock.Unlock()

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
	player.setKillStreak(newStreak)
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
func (player *Player) incrementGoalsScored() {
	player.goalsScoredLock.Lock()
	defer player.goalsScoredLock.Unlock()
	player.goalsScored++
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

func handleDeath(player *Player) {
	player.removeFromTileAndStage()
	player.incrementDeathCount()
	player.setHealth(150)
	player.setKillStreak(0)
	player.actions = createDefaultActions()

	stage := getStageFromStageName(player.world, infirmaryStagenameForPlayer(player))
	player.setStage(stage)
	player.updateRecord()

	placePlayerOnStageAt(player, stage, 2, 2)
	player.updateInformation()
}

func (player *Player) updateRecord() {
	currentTile := player.getTileSync()
	go player.world.db.updateRecordForPlayer(player, currentTile)
}

func (player *Player) removeFromTileAndStage() {
	/*
		if !player.tile.removePlayerAndNotifyOthers(player) {
			fmt.Println("Trying again") // Can prevent race with transfer but not perfect
			player.tile.removePlayerAndNotifyOthers(player)
		}
	*/
	player.stageLock.Lock()
	defer player.stageLock.Unlock()
	player.tileLock.Lock()
	defer player.tileLock.Unlock()
	// NotifyAllExcept
	player.tile.addMoneyAndNotifyAll(max(halveMoneyOf(player), 10)) // Tile money needs mutex.
	player.tile.removePlayerAndNotifyOthers(player)

	player.stage.removePlayerById(player.id)
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
//   Movement

func (p *Player) moveNorth() {
	p.move(-1, 0)
}

func (p *Player) moveNorthBoost() {
	p.moveBoost(-1, 0)
}

func (p *Player) moveSouth() {
	p.move(1, 0)
}

func (p *Player) moveSouthBoost() {
	p.moveBoost(1, 0)
}

func (p *Player) moveEast() {
	p.move(0, 1)
}

func (p *Player) moveEastBoost() {
	p.moveBoost(0, 1)
}

func (p *Player) moveWest() {
	p.move(0, -1)
}

func (p *Player) moveWestBoost() {
	p.moveBoost(0, -1)
}

func (p *Player) move(yOffset int, xOffset int) {
	sourceTile := p.getTileSync()
	destTile := p.world.getRelativeTile(sourceTile, yOffset, xOffset)
	p.push(destTile, nil, yOffset, xOffset)
	if walkable(destTile) {
		transferPlayer(p, sourceTile, destTile)
	}
}

func (p *Player) moveBoost(yOffset int, xOffset int) {
	if p.useBoost() {
		p.pushUnder(yOffset, xOffset)
		p.move(2*yOffset, 2*xOffset)
	} else {
		p.move(yOffset, xOffset)
	}
}

func (player *Player) applyTeleport(teleport *Teleport) {
	stage := getStageFromStageName(player.world, teleport.destStage)
	if !validCoordinate(teleport.destY, teleport.destX, stage.tiles) {
		log.Fatal("Fatal: Invalid coords from teleport: ", teleport.destStage, teleport.destY, teleport.destX)
	}
	transferPlayer(player, player.getTileSync(), stage.tiles[teleport.destY][teleport.destX])
}

// Atomic Transfers
func transferPlayer(p *Player, source, dest *Tile) {
	if source.stage == dest.stage {
		if transferPlayerWithinStage(p, source, dest) {
			updateOneAfterMovement(p, dest, source)
		}
	} else {
		if transferPlayerAcrossStages(p, source, dest) {
			spawnItemsFor(p, dest.stage)
			updateOneAfterStageChange(p)
		}
	}

}

func transferPlayerWithinStage(p *Player, source, dest *Tile) bool {
	p.tileLock.Lock()
	defer p.tileLock.Unlock()

	if !source.playerMutex.TryLock() {
		//fmt.Println("failed to get lock")
		return false
	}
	defer source.playerMutex.Unlock()

	if !dest.playerMutex.TryLock() {
		return false
	}
	defer dest.playerMutex.Unlock()

	_, ok := source.playerMap[p.id]
	if ok {
		delete(source.playerMap, p.id)
		dest.addLockedPlayertoLockedTile(p)
		go func() {
			source.stage.updateAllExcept(playerBox(source), p)
			dest.stage.updateAllExcept(playerBox(dest), p) // technically unneeded to getAnewPlayer
		}()
	}

	return ok
}

func transferPlayerAcrossStages(p *Player, source, dest *Tile) bool {
	p.stageLock.Lock()
	defer p.stageLock.Unlock()
	if p.stage == nil || p.stage == dest.stage {
		return false
	}

	if !p.stage.playerMutex.TryLock() {
		return false
	}
	defer p.stage.playerMutex.Unlock()

	if !dest.stage.playerMutex.TryLock() {
		return false
	}
	defer dest.stage.playerMutex.Unlock()

	p.tileLock.Lock()
	defer p.tileLock.Unlock()

	if !source.playerMutex.TryLock() {
		return false
	}
	defer source.playerMutex.Unlock()
	if !dest.playerMutex.TryLock() {
		return false
	}
	defer dest.playerMutex.Unlock()

	_, foundOnStage := p.stage.playerMap[p.id]

	_, foundOnTile := source.playerMap[p.id]

	success := foundOnStage && foundOnTile
	if success {
		delete(p.stage.playerMap, p.id)
		delete(source.playerMap, p.id)
		dest.stage.playerMap[p.id] = p
		p.stage = dest.stage
		dest.addLockedPlayertoLockedTile(p)

		go func() {
			source.stage.updateAllExcept(playerBox(source), p)
			dest.stage.updateAllExcept(playerBox(dest), p)
		}()
	}

	return success
}

////////////////////////////////////////////////////////////
//   Pushing

func (p *Player) push(tile *Tile, interactable *Interactable, yOff, xOff int) bool { // Returns if given interacable successfully pushed
	if tile == nil || tile.teleport != nil {
		return false
	}

	ownLock := tile.interactableMutex.TryLock()
	if !ownLock {
		return false // Tile is already locked by another operation
	}
	defer tile.interactableMutex.Unlock()

	if tile.interactable == nil {
		if tile.material.Walkable { // Prevents lock contention from using Walkable()
			// nil = nil ?
			if interactable != nil {
				tile.interactable = interactable
				tile.stage.updateAll(interactableBox(tile))
			}
			return true
		}
		return false
	}

	if tile.interactable.React(interactable, p, tile) {
		tile.stage.updateAll(interactableBox(tile)) // full tile?
		return true
	}

	if tile.interactable.pushable {
		nextTile := p.world.getRelativeTile(tile, yOff, xOff)
		if nextTile != nil {
			if p.push(nextTile, tile.interactable, yOff, xOff) {
				tile.interactable = interactable
				tile.stage.updateAll(interactableBox(tile))
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
				fmt.Println("Error - Stopping furture sends: ", err)
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
		// This spams the tests agressively because updatinghtmlbyPlayer calls this directly
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
		fmt.Printf("WARN: WriteMessage failed for player %s: %v\n", player.username, err)
		// Close connection if writes consistently fail ?
		if player.sessionTimeOutViolations.Add(1) >= 1 {
			return err
		}
	}

	return nil
}

// Updates - Enqueue

func updateOneAfterMovement(player *Player, current, previous *Tile) {
	impactedHighlights := player.updateSpaceHighlights()

	playerIcon := playerBoxSpecifc(current.y, current.x, player.getIconSync())

	previousBoxes := ""
	if previous != nil && previous.stage == player.getStageSync() {
		previousBoxes += playerBox(previous)
	}

	player.updates <- []byte(highlightBoxesForPlayer(player, impactedHighlights) + previousBoxes + playerIcon)
}

func updateOneAfterStageChange(p *Player) {
	p.setSpaceHighlights()
	updateScreenFromScratch(p)
	p.updateRecord() // too much?
}

func updateScreenFromScratch(player *Player) {
	player.clearUpdateBuffer <- struct{}{}
	clearChannel(player.updates)
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

func (player *Player) updateBottomText(message string) {
	msg := fmt.Sprintf(`
			<div id="bottom_text">
				&nbsp;&nbsp;> %s
			</div>`, message)
	updateOne(msg, player)
}

func updateOne(update string, player *Player) {
	player.updates <- []byte(update)
}

func (p *Player) trySend(msg []byte) {
	p.updates <- msg
}

/////////////////////////////////////////////////////////////
// Actions

type Actions struct {
	spaceHighlights     map[*Tile]bool
	spaceHighlightMutex sync.Mutex
	spaceStack          *StackOfPowerUp
	boostCounter        int
	boostMutex          sync.Mutex
}

type PowerUp struct {
	areaOfInfluence [][2]int
	//damageAtRadius  [4]int // unused
}

type StackOfPowerUp struct {
	powers     []*PowerUp
	powerMutex sync.Mutex
}

func createDefaultActions() *Actions {
	return &Actions{
		spaceHighlights:     map[*Tile]bool{},
		spaceHighlightMutex: sync.Mutex{},
		spaceStack:          &StackOfPowerUp{},
		boostCounter:        0,
		boostMutex:          sync.Mutex{},
	}
}

// Space Highlights

func (player *Player) setSpaceHighlights() {
	player.actions.spaceHighlightMutex.Lock()
	defer player.actions.spaceHighlightMutex.Unlock()
	player.actions.spaceHighlights = map[*Tile]bool{}
	currentTile := player.getTileSync()
	absCoordinatePairs := findOffsetsGivenPowerUp(currentTile.y, currentTile.x, player.actions.spaceStack.peek())
	for _, pair := range absCoordinatePairs {
		if validCoordinate(pair[0], pair[1], player.stage.tiles) {
			tile := player.stage.tiles[pair[0]][pair[1]]
			player.actions.spaceHighlights[tile] = true
		}
	}
}

func (player *Player) updateSpaceHighlights() []*Tile { // Returns removed highlights
	player.actions.spaceHighlightMutex.Lock()
	defer player.actions.spaceHighlightMutex.Unlock()
	previous := player.actions.spaceHighlights
	player.actions.spaceHighlights = map[*Tile]bool{}
	currentTile := player.getTileSync()
	absCoordinatePairs := findOffsetsGivenPowerUp(currentTile.y, currentTile.x, player.actions.spaceStack.peek())
	var impactedTiles []*Tile
	for _, pair := range absCoordinatePairs {
		if validCoordinate(pair[0], pair[1], player.stage.tiles) {
			tile := player.stage.tiles[pair[0]][pair[1]]
			player.actions.spaceHighlights[tile] = true
			if _, contains := previous[tile]; contains {
				delete(previous, tile)
			} else {
				impactedTiles = append(impactedTiles, tile)
			}
		}
	}
	return append(impactedTiles, mapOfTileToArray(previous)...)
}

func (player *Player) activatePower() {
	tilesToHighlight := make([]*Tile, 0, len(player.actions.spaceHighlights))

	playerHighlights := highlightMapToSlice(player)
	for _, tile := range playerHighlights {
		go tile.damageAll(50, player)
		destroyInteractable(tile, player)
		tile.eventsInFlight.Add(1)
		tilesToHighlight = append(tilesToHighlight, tile)

		go tile.tryToNotifyAfter(100)
	}
	damageBoxes := sliceOfTileToWeatherBoxes(tilesToHighlight, randomFieryColor())
	player.stage.updateAll(damageBoxes)
	updateOne(sliceOfTileToHighlightBoxes(tilesToHighlight, ""), player)

	player.actions.spaceHighlights = map[*Tile]bool{}
	if player.actions.spaceStack.hasPower() {
		player.nextPower()
	}
}

func (player *Player) nextPower() {
	player.actions.spaceStack.pop() // Throw old power away
	player.setSpaceHighlights()
	updateOne(sliceOfTileToHighlightBoxes(highlightMapToSlice(player), spaceHighlighter()), player)
}

func highlightMapToSlice(player *Player) []*Tile {
	out := make([]*Tile, 0)
	player.actions.spaceHighlightMutex.Lock()
	defer player.actions.spaceHighlightMutex.Unlock()
	for tile := range player.actions.spaceHighlights {
		out = append(out, tile)
	}
	return out
}

// Boosts

func (player *Player) addBoosts(n int) {
	// not thread-safe
	player.actions.boostCounter += n
	updateOne(divPlayerInformation(player), player)
}

func (player *Player) useBoost() bool {
	player.actions.boostMutex.Lock()
	defer player.actions.boostMutex.Unlock()
	if player.actions.boostCounter > 0 {
		player.actions.boostCounter--
	}
	updateOne(divPlayerInformation(player), player)
	return player.actions.boostCounter > 0
}

func (stack *StackOfPowerUp) hasPower() bool {
	stack.powerMutex.Lock()
	defer stack.powerMutex.Unlock()
	return len(stack.powers) > 0
}

// Power up stack

func (stack *StackOfPowerUp) pop() *PowerUp {
	stack.powerMutex.Lock()
	defer stack.powerMutex.Unlock()
	if len(stack.powers) > 0 {
		out := stack.powers[len(stack.powers)-1]
		stack.powers = stack.powers[:len(stack.powers)-1]
		return out
	}
	return nil // Should be impossible but return default power instead?
}

func (stack *StackOfPowerUp) peek() *PowerUp {
	stack.powerMutex.Lock()
	defer stack.powerMutex.Unlock()
	if len(stack.powers) > 0 {
		return stack.powers[len(stack.powers)-1]
	}
	return nil
}

func (stack *StackOfPowerUp) push(power *PowerUp) *StackOfPowerUp {
	stack.powerMutex.Lock()
	defer stack.powerMutex.Unlock()
	stack.powers = append(stack.powers, power)
	return stack
}
