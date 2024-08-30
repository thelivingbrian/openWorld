package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Player struct {
	id         string
	username   string
	world      *World
	stage      *Stage
	tile       *Tile
	stageName  string
	conn       *websocket.Conn
	connLock   sync.Mutex
	x          int
	y          int
	actions    *Actions
	health     int
	money      int
	moneyLock  sync.Mutex
	experience int
	killstreak int
	streakLock sync.Mutex
}

// Health observer, All Health changes should go through here
func (player *Player) setHealth(n int) {
	player.health = n
	if player.isDead() {
		player.handleDeath()
		return
	}
	updateOne(divPlayerInformation(player), player)
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

// Streak observer, All Money changes should go through here
func (player *Player) setKillStreak(n int) {
	player.streakLock.Lock()
	//defer player.streakLock.Unlock()
	player.killstreak = n
	player.streakLock.Unlock()

	player.world.leaderBoard.mostDangerous.Update(player)
	// New method. This is blocking defer
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

func (player *Player) isDead() bool {
	return player.health <= 0
}

func (player *Player) addToHealth(n int) bool {
	newHealth := player.health + n
	player.setHealth(newHealth)
	return newHealth > 0
}

// always called with placeOnStage?
func (p *Player) assignStageAndListen() {
	stage := p.world.getNamedStageOrDefault(p.stageName)
	fmt.Println("Have a stage")
	if stage == nil {
		log.Fatal("Fatal: Default Stage Not Found.")
	}
	p.stage = stage
}

func (p *Player) placeOnStage() {
	p.stage.addPlayer(p)

	// This is unsafe  (out of range)
	p.stage.tiles[p.y][p.x].addPlayerAndNotifyOthers(p)
	updateScreenFromScratch(p)

	// Could be enhanced
	p.stage.spawnItems()
}

func (player *Player) handleDeath() {
	player.removeFromStage()
	respawn(player)
}

func (player *Player) updateRecord() {
	go player.world.db.updateRecordForPlayer(player)
}

func (player *Player) removeFromStage() {
	player.tile.removePlayerAndNotifyOthers(player)
	player.stage.removePlayerById(player.id)
}

// Recv type
func respawn(player *Player) {
	player.setHealth(150)
	player.setKillStreak(0)
	player.stageName = "clinic"
	player.x = 2
	player.y = 2
	player.actions = createDefaultActions()
	player.updateRecord()
	player.assignStageAndListen()
	player.placeOnStage()
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
		if p.y == 0 {
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
	if p.world.initialPush(newTile, yOffset, xOffset) {
		t := &Teleport{destStage: newTile.stage.name, destY: newTile.y, destX: newTile.x}
		previousTile := p.stage.tiles[p.y][p.x]

		p.stage.tiles[p.y][p.x].removePlayerAndNotifyOthers(p)
		p.applyTeleport(t)
		impactedTiles := p.updateSpaceHighlights()
		updateOneAfterMovement(p, impactedTiles, previousTile)
	}
}

func (p *Player) move(yOffset int, xOffset int) {
	destY := p.y + yOffset
	destX := p.x + xOffset

	if validCoordinate(destY, destX, p.stage.tiles) && walkable(p.stage.tiles[destY][destX]) {
		sourceTile := p.stage.tiles[p.y][p.x]
		destTile := p.stage.tiles[destY][destX]

		if p.world.initialPush(destTile, yOffset, xOffset) {
			sourceTile.removePlayerAndNotifyOthers(p) // The routines coming in can race where the first successfully removes and both add
			destTile.addPlayerAndNotifyOthers(p)

			previousTile := sourceTile
			impactedTiles := p.updateSpaceHighlights()
			updateOneAfterMovement(p, impactedTiles, previousTile)
		}
	}
}

func (p *Player) pushUnder(yOffset int, xOffset int) {
	currentTile := p.stage.tiles[p.y][p.x]
	if currentTile != nil && currentTile.interactable != nil {
		p.world.initialPush(p.stage.tiles[p.y][p.x], yOffset, xOffset)
	}
}

func (w *World) initialPush(tile *Tile, yOff, xOff int) bool {
	if tile == nil {
		return false
	}
	tile.interactableMutex.Lock()
	defer tile.interactableMutex.Unlock()
	w.push(tile, yOff, xOff)
	return walkable(tile)
}

func (w *World) push(tile *Tile, yOff, xOff int) bool { // Returns availability of the tile for an interactible
	if tile == nil || tile.teleport != nil {
		return false
	}
	if tile.interactable == nil {
		return walkable(tile)
	}
	if tile.interactable.pushable {
		nextTile := w.getRelativeTile(tile, yOff, xOff)
		if nextTile != nil {
			ownLock := nextTile.interactableMutex.TryLock()
			if !ownLock {
				return false // Tile is already locked by another operation
			}
			defer nextTile.interactableMutex.Unlock()
			if w.push(nextTile, yOff, xOff) {
				nextTile.interactable = tile.interactable
				tile.interactable = nil // Take *Interactable and assign here to have option of reacting
				nextTile.stage.updateAll(playerBox(nextTile))
				return true
			}
		} else {
			return false
		}
	}
	return false
}

func (w *World) getRelativeTile(tile *Tile, yOff, xOff int) *Tile {
	destY := tile.y + yOff
	destX := tile.x + xOff
	if validCoordinate(destY, destX, tile.stage.tiles) {
		return tile.stage.tiles[destY][destX]
	} else {
		escapesVertically, escapesHorizontally := validityByAxis(destY, destX, tile.stage.tiles)
		fmt.Printf("escapes vert:%t hoz:%t\n", escapesVertically, escapesHorizontally)
		if escapesVertically && escapesHorizontally {
			// in bloop world cardinal direction travel may be non-communative
			// therefore north-east etc neighbor is not uniquely defined
			// order can probably be uniquely determined when tile.y != tile.x
			return nil
		}
		if escapesVertically {
			var newStage *Stage
			if yOff > 0 {
				newStage = w.getStageByName(tile.stage.south)
				if newStage == nil {
					newStage = w.loadStageByName(tile.stage.south)
				}
			}
			if yOff < 0 {
				newStage = w.getStageByName(tile.stage.north)
				if newStage == nil {
					newStage = w.loadStageByName(tile.stage.north)
				}
			}

			if newStage != nil {
				if validCoordinate(mod(destY, len(newStage.tiles)), destX, newStage.tiles) {
					return newStage.tiles[mod(destY, len(newStage.tiles))][destX]
				}
			}
			return nil
		}
		if escapesHorizontally {
			var newStage *Stage
			if xOff > 0 {
				newStage = w.getStageByName(tile.stage.east)
				if newStage == nil {
					fmt.Println("New stage was nil")
					newStage = w.loadStageByName(tile.stage.east)
				}
			}
			if xOff < 0 {
				newStage = w.getStageByName(tile.stage.west)
				if newStage == nil {
					newStage = w.loadStageByName(tile.stage.west)
				}
			}

			if newStage != nil {
				if validCoordinate(destY, mod(destX, len(newStage.tiles)), newStage.tiles) {
					return newStage.tiles[destY][mod(destX, len(newStage.tiles))]
				}
			}
			return nil
		}

		return nil
	}
}

func (player *Player) nextPower() {
	player.actions.spaceStack.pop() // Throw old power away
	player.setSpaceHighlights()
	oobUpdateWithHud(player, mapOfTileToArray(player.actions.spaceHighlights))
}

func (player *Player) setSpaceHighlights() {
	player.actions.spaceHighlights = map[*Tile]bool{}
	absCoordinatePairs := applyRelativeDistance(player.y, player.x, player.actions.spaceStack.peek().areaOfInfluence)
	for _, pair := range absCoordinatePairs {
		if validCoordinate(pair[0], pair[1], player.stage.tiles) {
			tile := player.stage.tiles[pair[0]][pair[1]]
			player.actions.spaceHighlights[tile] = true
		}
	}
}

func (player *Player) updateSpaceHighlights() []*Tile { // Returns removed highlights
	previous := player.actions.spaceHighlights
	player.actions.spaceHighlights = map[*Tile]bool{}
	absCoordinatePairs := applyRelativeDistance(player.y, player.x, player.actions.spaceStack.peek().areaOfInfluence)
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
	for tile := range player.actions.spaceHighlights {
		tile.damageAll(50, player)

		tileToHighlight := tile.incrementAndReturnIfFirst()
		if tileToHighlight != nil {
			tilesToHighlight = append(tilesToHighlight, tileToHighlight)
		}

		go tile.tryToNotifyAfter(100) // Flat for player if more powers?
	}
	highlightHtml := sliceOfTileToColoredOoB(tilesToHighlight, randomFieryColor())
	player.stage.updateAll(highlightHtml)

	player.actions.spaceHighlights = map[*Tile]bool{}
	if player.actions.spaceStack.hasPower() {
		player.nextPower()
	}
}

func (player *Player) applyTeleport(teleport *Teleport) {
	if player.stageName != teleport.destStage {
		player.removeFromStage()
	}
	player.stageName = teleport.destStage
	player.y = teleport.destY
	player.x = teleport.destX
	player.updateRecord()
	player.assignStageAndListen()
	player.placeOnStage()
}

func (player *Player) updateBottomText(message string) {
	msg := fmt.Sprintf(`
			<div id="bottom_text">
				&nbsp;&nbsp;> %s
			</div>`, message)
	updateOne(msg, player)
}

/*
// Show boost when boost counter first breaks 0 with message explaining shift
// Hide boost on next movement

func (player *Player) showBoost() {
	player.actions.shiftEngaged = true
	player.actions.shiftHighlights = map[*Tile]bool{}
	absCoordinatePairs := applyRelativeDistance(player.y, player.x, jumpCross())
	for _, pair := range absCoordinatePairs {
		if validCoordinate(pair[0], pair[1], player.stage.tiles) {
			tile := player.stage.tiles[pair[0]][pair[1]]
			player.actions.shiftHighlights[tile] = true
		}
	}
	oobUpdateWithHud(player, mapOfTileToArray(player.actions.shiftHighlights))
}
func (player *Player) hideBoost() {
	player.actions.shiftEngaged = false
	previous := player.actions.shiftHighlights
	player.actions.shiftHighlights = map[*Tile]bool{}
	oobUpdateWithHud(player, mapOfTileToArray(previous))
}
*/

/////////////////////////////////////////////////////////////
// Actions

type Actions struct {
	spaceReadyable  bool
	spaceHighlights map[*Tile]bool
	spaceStack      *StackOfPowerUp
	boostCounter    int
}

type PowerUp struct {
	areaOfInfluence [][2]int
	damageAtRadius  [4]int
}

type StackOfPowerUp struct {
	powers     []*PowerUp
	powerMutex sync.Mutex
}

func (player *Player) addBoosts(n int) {
	//first := player.actions.boostCounter == 0
	player.actions.boostCounter += n
	/*if first {
		player.showBoost()
	}*/
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
	if stack.hasPower() {
		stack.powerMutex.Lock()
		defer stack.powerMutex.Unlock()
		out := stack.powers[len(stack.powers)-1]
		stack.powers = stack.powers[:len(stack.powers)-1]
		return out
	}
	return nil // Should be impossible but return default power instead?
}

// Watch this lead to item dupe bugs
func (stack *StackOfPowerUp) peek() PowerUp {
	if stack.hasPower() {
		stack.powerMutex.Lock()
		defer stack.powerMutex.Unlock()
		return *stack.powers[len(stack.powers)-1]
	}
	return PowerUp{}
}

func (stack *StackOfPowerUp) push(power *PowerUp) *StackOfPowerUp {
	stack.powerMutex.Lock()
	defer stack.powerMutex.Unlock()
	stack.powers = append(stack.powers, power)
	return stack
}

func createDefaultActions() *Actions {
	return &Actions{false, map[*Tile]bool{}, &StackOfPowerUp{}, 0}
}
