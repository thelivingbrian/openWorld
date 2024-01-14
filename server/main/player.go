package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Player struct {
	id        string
	username  string
	stage     *Stage // may not need
	tile      *Tile
	stageName string
	conn      *websocket.Conn
	connLock  sync.Mutex
	x         int // Definitely may not need
	y         int
	actions   *Actions
	health    int
	money     int
}

type Actions struct {
	space           bool
	spaceShape      [][2]int
	spaceHighlights map[*Tile]bool
}

func createDefaultActions() *Actions {
	return &Actions{false, grid7x7, map[*Tile]bool{}}
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

func (player *Player) addToHealth(n int) {
	player.setHealth(player.health + n)
}

func (p *Player) assignStage() {
	fmt.Println("Getting Stage")
	existingStage := getStageByName(p.stageName)
	if existingStage == nil {
		playerMutex.Lock()
		defer playerMutex.Unlock()
		delete(playerMap, p.id)
		log.Fatal("Stage Not Found.")

	}
	p.stage = existingStage
}

func (p *Player) placeOnStage() {

	fmt.Println("Assigning stage and tile")
	//p.stage = existingStage
	p.stage.playerMutex.Lock()
	p.stage.playerMap[p.id] = p
	p.stage.playerMutex.Unlock()

	p.stage.tiles[p.y][p.x].addPlayer(p)
	fmt.Println("Added Player")
	p.stage.markAllDirty()
	fmt.Println("Marked dirty")
}

/*
func (player *Player) placeOnTile(tile *Tile) {
	player.stage = tile.stage
	tile.addPlayer(player)
	tile.stage.addPlayer(player)
	tile.stage.markAllDirty()
}
*/

func (player *Player) handleDeath() {
	player.removeFromStage()
	respawn(player)
}

func (player *Player) removeFromStage() {
	player.tile.removePlayer(player.id)
	player.stage.removePlayerById(player.id)
}

func respawn(player *Player) {
	//clinic := getClinic()
	player.health = 100
	//player.stage = clinic
	player.stageName = "clinic"
	player.x = 2
	player.y = 2
	player.actions = createDefaultActions()
	player.assignStage()
	player.placeOnStage()
}

func updateScreenWithStarter(player *Player, html string) {
	if player.isDead() {
		respawn(player)
		return
	}
	html += hudAsOutOfBound(player)
	player.stage.updates <- Update{player, []byte(html)}
}

func updateScreenFromScratch(player *Player) {
	if player.isDead() {
		respawn(player)
		return
	}
	player.stage.updates <- Update{player, htmlFromPlayer(player)}
}

func oobUpdateWithHud(player *Player, update string) {
	player.stage.updates <- Update{player, []byte(update + hudAsOutOfBound(player))}
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

		currentStage := p.stage // Stage may change as result of teleport or etc
		oobAll := currentTile.removePlayer(p.id)
		oobAll = oobAll + destTile.addPlayer(p)
		currentStage.updateAllExcept(oobAll, p)

		oobRemoveHighlights := ""
		if p.actions.space {
			oobRemoveHighlights = mapOfTileToOoB(p.setSpaceHighlights())
		}
		if currentStage == p.stage {
			updateOneWithHud(oobAll+oobRemoveHighlights, p)
		}
	}
}

func (player *Player) turnSpaceOn() {
	player.actions.space = true
	player.setSpaceHighlights()

	player.stage.updates <- Update{player, []byte(hudAsOutOfBound(player))}
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
		tile.damageAll(25)
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
	player.assignStage()
	player.placeOnStage()
}
