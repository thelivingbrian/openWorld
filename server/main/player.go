package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Player struct {
	id        string
	stage     *Stage
	stageName string
	conn      *websocket.Conn
	connLock  sync.Mutex
	x         int
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

func (player *Player) isDead() bool {
	return player.health <= 0
}

func placeOnStage(p *Player) {
	x := p.x // Feels backwards but maybe needed for loading  from db?
	y := p.y
	p.stage.tiles[y][x].addPlayer(p)
	p.stage.playerMap[p.id] = p
	p.stage.markAllDirty()
}

func handleDeathOf(player *Player) {
	// Implement??
}

func respawn(player *Player) {
	clinic := getClinic()
	player.health = 100
	player.stage = clinic
	player.stageName = "clinic"
	player.x = 2
	player.y = 2
	player.actions = createDefaultActions()
	placeOnStage(player)
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
	move(p, -1, 0)
}

func moveSouth(p *Player) {
	move(p, 1, 0)
}

func moveEast(p *Player) {
	move(p, 0, 1)
}

func moveWest(p *Player) {
	move(p, 0, -1)
}

func move(p *Player, yOffset int, xOffset int) {
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

func validCoordinate(y int, x int, tiles [][]*Tile) bool {
	if y < 0 || y >= len(tiles) {
		return false
	}
	if x < 0 || x >= len(tiles[y]) {
		return false
	}
	return true
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

func mapOfTileToOoB(m map[*Tile]bool) string {
	html := ``
	for tile := range m {
		html += htmlFromTile(tile)
	}
	return html
}
