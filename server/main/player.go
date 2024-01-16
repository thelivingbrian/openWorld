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
	updateScreenFromScratch(p)
}

func (player *Player) handleDeath() {
	player.removeFromStage()
	respawn(player)
}

func (player *Player) updateRecord() {
	go player.world.db.updateRecordForPlayer(*player)
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

func moveNorth(p *Player) {
	p.move(-1, 0)
}

func moveSouth(p *Player) {
	p.move(1, 0)
}

func moveEast(p *Player) {
	p.move(0, 1)
}

func moveWest(p *Player) {
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
		if p.actions.space {
			oobRemoveHighlights = mapOfTileToOoB(p.setSpaceHighlights())
		}
		if currentStage == p.stage {
			updateOneWithHud(oobRemoveHighlights+htmlFromTile(currentTile), p)
		}
	}
}

func (player *Player) turnSpaceOn() {
	player.actions.space = true
	player.setSpaceHighlights()
	updateOne(hudAsOutOfBound(player), player)
}

func (player *Player) setSpaceHighlights() map[*Tile]bool { // Returns removed highlights
	previous := player.actions.spaceHighlights
	player.actions.spaceHighlights = map[*Tile]bool{}
	absCoordinatePairs := applyRelativeDistance(player.y, player.x, player.actions.spaceShape)
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
	player.actions.space = false

	for tile := range player.actions.spaceHighlights {
		tile.damageAll(25, player)
	}
	htmlRemoveHighlights := mapOfTileToOoB(player.actions.spaceHighlights)

	player.actions.spaceHighlights = map[*Tile]bool{}

	htmlAddHud := hudAsOutOfBound(player)
	player.stage.updates <- Update{player, []byte(htmlRemoveHighlights + htmlAddHud)}
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
	space           bool
	spaceShape      [][2]int
	spaceHighlights map[*Tile]bool
}

func createDefaultActions() *Actions {
	return &Actions{false, grid7x7, map[*Tile]bool{}}
}
