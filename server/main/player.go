package main

import (
	"fmt"
	"log"
	"sync"
	"time"

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
	fmt.Println("updating from scratch")
	updateScreenFromScratch(p) // This is using an old method for computing the highlights (Which weirdly works because space highlights have not yet been set)
	fmt.Println("Adding power up")
	p.stage.tiles[4][4].addPowerUpAndNotifyAll(p) // This is completing before the space highlights are being set after a teleport at the end of move()
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

func (p *Player) move(yOffset int, xOffset int) {
	destY := p.y + yOffset
	destX := p.x + xOffset
	if validCoordinate(destY, destX, p.stage.tiles) && walkable(p.stage.tiles[destY][destX]) {
		currentTile := p.stage.tiles[p.y][p.x]
		destTile := p.stage.tiles[destY][destX]

		currentStage := p.stage                    // Stage may change as result of teleport or etc
		currentTile.removePlayerAndNotifyOthers(p) // The routines coming in can race where the first successfully removes and both add
		destTile.addPlayerAndNotifyOthers(p)

		oobRemoveHighlights := ""
		if p.actions.spaceActive {
			oobRemoveHighlights = mapOfTileToOoB(p.setSpaceHighlights())
		}
		if currentStage == p.stage {
			updateOneWithHud(oobRemoveHighlights+htmlFromTile(currentTile), p)
		}
	}
}

func (player *Player) turnSpaceOn() {
	player.actions.spaceActive = true

	player.setSpaceHighlights()
	updateOne(hudAsOutOfBound(player), player)
}

func (player *Player) setSpaceHighlights() map[*Tile]bool { // Returns removed highlights
	fmt.Println("Setting Space Highlights")
	previous := player.actions.spaceHighlights
	player.actions.spaceHighlights = map[*Tile]bool{}
	absCoordinatePairs := applyRelativeDistance(player.y, player.x, player.actions.spacePower.areaOfInfluence)
	for _, pair := range absCoordinatePairs {
		if validCoordinate(pair[0], pair[1], player.stage.tiles) {
			tile := player.stage.tiles[pair[0]][pair[1]]
			player.actions.spaceHighlights[tile] = true
			delete(previous, tile)
		}
	}
	return previous
}

func (player *Player) turnSpaceOff() {
	player.actions.spaceActive = false

	tilesToHighlight := make([]*Tile, 0, len(player.actions.spaceHighlights))
	for tile := range player.actions.spaceHighlights {
		tile.damageAll(25, player)

		tileToHighlight := tile.incrementAndReturnIfFirst()
		if tileToHighlight != nil {
			tilesToHighlight = append(tilesToHighlight, tileToHighlight)
		}

		go tile.tryToNotifyAfter(100)
	}
	highlightHtml := sliceOfTileToColoredOoB(tilesToHighlight, randomFieryColor())
	player.stage.updateAll(highlightHtml)

	go player.stage.updateAllWithHudAfterDelay(110)

	player.actions.spaceHighlights = map[*Tile]bool{}
	if player.actions.spaceStack.hasPower() {
		player.actions.spacePower = player.actions.spaceStack.pop()
		go func() {
			time.Sleep(250 * time.Millisecond)
			player.turnSpaceOn()
		}()
	} else {
		player.actions.spaceReadyable = false
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

/////////////////////////////////////////////////////////////
// Actions

type Actions struct {
	spaceReadyable bool
	spaceActive    bool
	//spaceShape      [][2]int // *PowerUp Instead
	spaceHighlights map[*Tile]bool
	spacePower      *PowerUp // Need to do this or add a peek method
	spaceStack      *StackOfPowerUp
}

type PowerUp struct {
	areaOfInfluence [][2]int
	damageAtRadius  [4]int
}

type StackOfPowerUp struct {
	powers     []*PowerUp
	powerMutex sync.Mutex
}

func (stack *StackOfPowerUp) hasPower() bool {
	stack.powerMutex.Lock()
	defer stack.powerMutex.Unlock()
	return len(stack.powers) > 0
}

func (stack *StackOfPowerUp) pop() *PowerUp {
	if stack.hasPower() {
		stack.powerMutex.Lock()
		defer stack.powerMutex.Unlock()
		out := stack.powers[len(stack.powers)-1]
		stack.powers = stack.powers[:len(stack.powers)-1]
		return out
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
	return &Actions{false, false, map[*Tile]bool{}, nil, &StackOfPowerUp{}}
}
