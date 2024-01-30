package main

import (
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
	experience int
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
	player.money = n
	updateOne(divPlayerInformation(player), player)
}

func (player *Player) isDead() bool {
	return player.health <= 0
}

func (player *Player) addToHealth(n int) bool {
	newHealth := player.health + n
	player.setHealth(newHealth)
	return newHealth > 0
}

func (p *Player) assignStageAndListen() {
	stage, new := p.world.getStageByName(p.stageName)
	if stage == nil {
		log.Fatal("Fatal: Stage Not Found.")
	}
	if new {
		go stage.sendUpdates()
	}
	p.stage = stage
}

func (p *Player) placeOnStage() {
	p.stage.playerMutex.Lock()
	p.stage.playerMap[p.id] = p
	p.stage.playerMutex.Unlock()

	p.stage.tiles[p.y][p.x].addPlayerAndNotifyOthers(p)
	updateScreenFromScratch(p) // This is using an old method for computing the highlights (Which weirdly works because space highlights have not yet been set)

	// Need spawn logic
	p.stage.tiles[4][4].addPowerUpAndNotifyAll(p, grid7x7) // This is completing before the space highlights are being set after a teleport at the end of move()
	p.stage.tiles[5][5].addPowerUpAndNotifyAll(p, grid3x3)
	p.stage.tiles[6][6].addPowerUpAndNotifyAll(p, x())
	p.stage.tiles[6][6].addBoostsAndNotifyAll(p)
	p.stage.tiles[5][5].money += 10
	p.stage.tiles[5][5].addBoostsAndNotifyAll(p)
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
	player.setHealth(100)
	player.stageName = "clinic"
	player.x = 2
	player.y = 2
	player.actions = createDefaultActions()
	player.updateRecord()
	player.assignStageAndListen()
	player.placeOnStage()
}

func (p *Player) moveNorth() {
	p.move(-1, 0)
}

func (p *Player) moveSouth() {
	p.move(1, 0)
}

func (p *Player) moveEast() {
	p.move(0, 1)
}

func (p *Player) moveWest() {
	p.move(0, -1)
}

func (p *Player) moveNorthBoost() {
	if p.actions.boostCounter > 0 {
		p.useBoost()
		p.move(-2, 0)
	} else {
		p.move(-1, 0)
	}
}

func (p *Player) moveSouthBoost() {
	if p.actions.boostCounter > 0 {
		p.useBoost()
		p.move(2, 0)
	} else {
		p.move(1, 0)
	}
}

func (p *Player) moveEastBoost() {
	if p.actions.boostCounter > 0 {
		p.useBoost()
		p.move(0, 2)
	} else {
		p.move(0, 1)
	}
}

func (p *Player) moveWestBoost() {
	if p.actions.boostCounter > 0 {
		p.useBoost()
		p.move(0, -2)
	} else {
		p.move(0, -1)
	}
}

func (p *Player) move(yOffset int, xOffset int) {
	destY := p.y + yOffset
	destX := p.x + xOffset
	if validCoordinate(destY, destX, p.stage.tiles) && walkable(p.stage.tiles[destY][destX]) {
		currentTile := p.stage.tiles[p.y][p.x]
		destTile := p.stage.tiles[destY][destX]

		currentStage := p.stage                    // Stage may change as result of teleport or etc
		currentTile.removePlayerAndNotifyOthers(p) // The routines coming in can race where the first successfully removes and both add
		destTile.addPlayerAndNotifyOthers(p)

		oobRemoveHighlights := mapOfTileToOoB(p.updateSpaceHighlights())
		//p.updateShiftHighlights()
		oobRemoveHighlights += mapOfTileToOoB(p.updateShiftHighlights())

		if currentStage == p.stage {
			updateOneWithHud(oobRemoveHighlights+htmlFromTile(currentTile), p)
		}
	}
}

func (player *Player) nextPower() {
	player.actions.spaceStack.pop() // Throw old power away
	previous := player.actions.spaceHighlights
	player.setSpaceHighlights()
	updateOneWithHud(mapOfTileToOoB(previous), player)
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

func (player *Player) updateSpaceHighlights() map[*Tile]bool { // Returns removed highlights
	previous := player.actions.spaceHighlights
	player.actions.spaceHighlights = map[*Tile]bool{}
	absCoordinatePairs := applyRelativeDistance(player.y, player.x, player.actions.spaceStack.peek().areaOfInfluence)
	for _, pair := range absCoordinatePairs {
		if validCoordinate(pair[0], pair[1], player.stage.tiles) {
			tile := player.stage.tiles[pair[0]][pair[1]]
			player.actions.spaceHighlights[tile] = true
			delete(previous, tile)
		}
	}
	return previous
}

func (player *Player) updateShiftHighlights() map[*Tile]bool { // Returns removed highlights
	previous := player.actions.shiftHighlights
	player.actions.shiftHighlights = map[*Tile]bool{}
	absCoordinatePairs := applyRelativeDistance(player.y, player.x, jumpCross()) // Shift shape lives in multiple places
	for _, pair := range absCoordinatePairs {
		if validCoordinate(pair[0], pair[1], player.stage.tiles) {
			tile := player.stage.tiles[pair[0]][pair[1]]
			player.actions.shiftHighlights[tile] = true
			delete(previous, tile)
		}
	}
	return previous
}

func (player *Player) activatePower() {
	tilesToHighlight := make([]*Tile, 0, len(player.actions.spaceHighlights))
	for tile := range player.actions.spaceHighlights {
		tile.damageAll(25, player)

		tileToHighlight := tile.incrementAndReturnIfFirst()
		if tileToHighlight != nil {
			tilesToHighlight = append(tilesToHighlight, tileToHighlight)
		}

		go tile.tryToNotifyAfter(100) // Flat for player if more powers?
	}
	highlightHtml := sliceOfTileToColoredOoB(tilesToHighlight, randomFieryColor())
	player.stage.updateAll(highlightHtml)

	go player.stage.updateAllWithHudAfterDelay(110) // Undamage screen huds after the explosion

	player.actions.spaceHighlights = map[*Tile]bool{}
	if player.actions.spaceStack.hasPower() {
		player.nextPower()
	}
}

func (player *Player) applyTeleport(teleport *Teleport) {
	player.stageName = teleport.destStage
	player.y = teleport.destY
	player.x = teleport.destX
	player.updateRecord()
	player.assignStageAndListen()
	player.placeOnStage()
}

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
	updateOneWithHud("", player)
}
func (player *Player) hideBoost() {
	player.actions.shiftEngaged = false
	previous := player.actions.shiftHighlights
	player.actions.shiftHighlights = map[*Tile]bool{}
	updateOneWithHud(mapOfTileToOoB(previous), player)
}

/////////////////////////////////////////////////////////////
// Actions

type Actions struct {
	spaceReadyable  bool
	spaceHighlights map[*Tile]bool
	spaceStack      *StackOfPowerUp
	boostCounter    int
	shiftHighlights map[*Tile]bool
	shiftEngaged    bool
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
	first := player.actions.boostCounter == 0
	player.actions.boostCounter += n
	if first {
		player.showBoost()
	}
	updateOne(divPlayerInformation(player), player)
}

func (player *Player) useBoost() {
	player.actions.boostCounter--
	if player.actions.boostCounter == 0 {
		player.hideBoost()
	}
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
	return &Actions{false, map[*Tile]bool{}, &StackOfPowerUp{}, 0, map[*Tile]bool{}, false}
}
