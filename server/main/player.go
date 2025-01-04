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
	stageName                string
	conn                     WebsocketConnection
	connLock                 sync.RWMutex
	tangible                 bool
	tangibilityLock          sync.Mutex
	// x, y are highly mutated and are unsafe to read/difficult to lock. Use tile instead ?
	x          int
	y          int
	actions    *Actions
	health     int
	healthLock sync.Mutex
	money      int
	moneyLock  sync.Mutex

	killCount       int
	killCountLock   sync.Mutex
	deathCount      int
	deathCountLock  sync.Mutex
	goalsScored     int
	goalsScoredLock sync.Mutex
	//experience int //?

	killstreak int
	streakLock sync.Mutex
	menues     map[string]Menu
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
	player.stageName = stage.name
}

func (player *Player) setStageName(name string) {
	player.stageLock.Lock()
	defer player.stageLock.Unlock()
	player.stageName = name
}

func (player *Player) getStageNameSync() string {
	player.stageLock.Lock()
	defer player.stageLock.Unlock()
	return player.stageName
}

func (player *Player) getStageSync() *Stage {
	player.stageLock.Lock()
	defer player.stageLock.Unlock()
	return player.stage
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

func getStageFromStageName(world *World, stageName string) *Stage {
	stage := world.getNamedStageOrDefault(stageName)
	if stage == nil {
		log.Fatal("Fatal: Default Stage Not Found.")
	}

	return stage
}

func placePlayerOnStageAt(p *Player, stage *Stage, y, x int) {
	if y >= len(stage.tiles) || x >= len(stage.tiles[y]) {
		log.Fatal("Fatal: Invalid coords to place on stage.")
	}

	p.setStage(stage)
	spawnItemsFor(p, stage)
	stage.addPlayer(p)
	stage.tiles[y][x].addPlayerAndNotifyOthers(p)
	p.setSpaceHighlights()
	updateScreenFromScratch(p)
}

func spawnItemsFor(p *Player, stage *Stage) {
	// Should be nil safe but test needed
	for i := range stage.spawn {
		stage.spawn[i].activateFor(p, stage)
	}
}

func handleDeath(player *Player) {
	player.tileLock.Lock()
	// NotifyAllExcept
	player.tile.addMoneyAndNotifyAll(max(halveMoneyOf(player), 10)) // Tile money needs mutex.
	player.tileLock.Unlock()
	player.removeFromTileAndStage()
	player.incrementDeathCount()
	player.setStageName(infirmaryStagenameForPlayer(player))
	// set stagename
	respawn(player)
}

func (player *Player) updateRecord() {
	go player.world.db.updateRecordForPlayer(player)
}

func (player *Player) removeFromTileAndStage() {
	/*
		if !player.tile.removePlayerAndNotifyOthers(player) {
			fmt.Println("Trying again") // Can prevent race with transfer but not perfect
			player.tile.removePlayerAndNotifyOthers(player)
		}
	*/
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

func respawn(player *Player) {
	player.tangibilityLock.Lock()
	defer player.tangibilityLock.Unlock()
	if !player.tangible {
		return
	}

	player.setHealth(150)
	player.setKillStreak(0)
	//player.setStageName("clinic") // Do in handle death as well in case player became intangible?
	player.x = 2
	player.y = 2
	player.actions = createDefaultActions()
	player.updateRecord()
	stage := getStageFromStageName(player.world, player.getStageNameSync())
	placePlayerOnStageAt(player, stage, 2, 2)

	// redo because place wipes buffer
	player.updateInformation()
}

func (p *Player) moveNorth() {
	if p.y == 0 {
		p.tryGoNeighbor(-1, 0)
		return
	}
	p.move(-1, 0)
}

func (p *Player) moveNorthBoost() {
	if p.actions.boostCounter > 0 {
		p.useBoost()
		p.pushUnder(-1, 0)
		if p.y <= 1 {
			p.tryGoNeighbor(-2, 0)
			return
		}
		p.move(-2, 0)
	} else {
		p.moveNorth()
	}
}

func (p *Player) moveSouth() {
	if p.y == len(p.stage.tiles)-1 {
		p.tryGoNeighbor(1, 0)
		return
	}
	p.move(1, 0)
}

func (p *Player) moveSouthBoost() {
	if p.actions.boostCounter > 0 {
		p.useBoost()
		p.pushUnder(1, 0)
		if p.y == len(p.stage.tiles)-1 || p.y == len(p.stage.tiles)-2 {
			p.tryGoNeighbor(2, 0)
			return
		}
		p.move(2, 0)
	} else {
		p.moveSouth()
	}
}

func (p *Player) moveEast() {
	if p.x == len(p.stage.tiles[p.y])-1 {
		p.tryGoNeighbor(0, 1)
		return
	}
	p.move(0, 1)
}

func (p *Player) moveEastBoost() {
	if p.actions.boostCounter > 0 {
		p.useBoost()
		p.pushUnder(0, 1)
		if p.x == len(p.stage.tiles[p.y])-1 || p.x == len(p.stage.tiles[p.y])-2 {
			p.tryGoNeighbor(0, 2)
			return
		}
		p.move(0, 2)
	} else {
		p.moveEast()
	}
}

func (p *Player) moveWest() {
	if p.x == 0 {
		p.tryGoNeighbor(0, -1)
		return
	}
	p.move(0, -1)
}

func (p *Player) moveWestBoost() {
	if p.actions.boostCounter > 0 {
		p.useBoost()
		p.pushUnder(0, -1)
		if p.x == 0 || p.x == 1 {
			p.tryGoNeighbor(0, -2)
			return
		}
		p.move(0, -2)
	} else {
		p.moveWest()
	}
}

func (p *Player) tryGoNeighbor(yOffset, xOffset int) {
	newTile := p.world.getRelativeTile(p.stage.tiles[p.y][p.x], yOffset, xOffset)
	p.push(newTile, nil, yOffset, xOffset)
	if walkable(newTile) {
		t := &Teleport{destStage: newTile.stage.name, destY: newTile.y, destX: newTile.x}

		p.stage.tiles[p.y][p.x].removePlayerAndNotifyOthers(p)
		p.applyTeleport(t)
	}
}

func (p *Player) move(yOffset int, xOffset int) {
	destY := p.y + yOffset
	destX := p.x + xOffset

	if validCoordinate(destY, destX, p.stage.tiles) {
		sourceTile := p.stage.tiles[p.y][p.x]
		destTile := p.stage.tiles[destY][destX]

		p.push(destTile, nil, yOffset, xOffset)
		if walkable(destTile) {
			// atomic map swap
			if transferPlayer(p, sourceTile, destTile) {
				impactedTiles := p.updateSpaceHighlights()
				updateOneAfterMovement(p, impactedTiles, sourceTile)
			}
		}
	}
}

func transferPlayer(p *Player, source, dest *Tile) bool {
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

	_, ok := source.playerMap[p.id]
	if ok {
		delete(source.playerMap, p.id)
		ok = dest.addLockedPlayertoLockedTile(p)
		go func() {
			source.stage.updateAllExcept(playerBox(source), p)
			dest.stage.updateAllExcept(playerBox(dest), p)
		}()
	}

	return ok
}

func (p *Player) pushUnder(yOffset int, xOffset int) {
	currentTile := p.stage.tiles[p.y][p.x]
	if currentTile != nil && currentTile.interactable != nil {
		p.push(p.stage.tiles[p.y][p.x], nil, yOffset, xOffset)
	}
}

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
			tile.interactable = interactable
			tile.stage.updateAll(interactableBox(tile))
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

func (player *Player) nextPower() {
	player.actions.spaceStack.pop() // Throw old power away
	player.setSpaceHighlights()
	updateOne(sliceOfTileToHighlightBoxes(highlightMapToSlice(player), spaceHighlighter()), player)
}

func (player *Player) setSpaceHighlights() {
	player.actions.spaceHighlightMutex.Lock()
	defer player.actions.spaceHighlightMutex.Unlock()
	player.actions.spaceHighlights = map[*Tile]bool{}
	absCoordinatePairs := findOffsetsGivenPowerUp(player.y, player.x, player.actions.spaceStack.peek())
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
	absCoordinatePairs := findOffsetsGivenPowerUp(player.y, player.x, player.actions.spaceStack.peek())
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

func highlightMapToSlice(player *Player) []*Tile {
	out := make([]*Tile, 0)
	player.actions.spaceHighlightMutex.Lock()
	defer player.actions.spaceHighlightMutex.Unlock()
	for tile := range player.actions.spaceHighlights {
		out = append(out, tile)
	}
	return out
}

func (player *Player) applyTeleport(teleport *Teleport) {
	if player.stageName != teleport.destStage {
		player.stage.removePlayerById(player.id)
	}
	stage := getStageFromStageName(player.world, teleport.destStage)
	placePlayerOnStageAt(player, stage, teleport.destY, teleport.destX)

	player.updateRecord() // no?
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

func (player *Player) updateBottomText(message string) {
	msg := fmt.Sprintf(`
			<div id="bottom_text">
				&nbsp;&nbsp;> %s
			</div>`, message)
	updateOne(msg, player)
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
}

type PowerUp struct {
	areaOfInfluence [][2]int
	//damageAtRadius  [4]int // unused
}

type StackOfPowerUp struct {
	powers     []*PowerUp
	powerMutex sync.Mutex
}

func (player *Player) addBoosts(n int) {
	// not thread-safe
	player.actions.boostCounter += n
	updateOne(divPlayerInformation(player), player)
}

func (player *Player) useBoost() {
	player.actions.boostCounter--
	updateOne(divPlayerInformation(player), player)
}

func (stack *StackOfPowerUp) hasPower() bool {
	stack.powerMutex.Lock()
	defer stack.powerMutex.Unlock()
	return len(stack.powers) > 0
}

// Don't even return anything? Delete?
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

func createDefaultActions() *Actions {
	return &Actions{
		spaceHighlights:     map[*Tile]bool{},
		spaceHighlightMutex: sync.Mutex{},
		spaceStack:          &StackOfPowerUp{},
		boostCounter:        0,
	}
}
