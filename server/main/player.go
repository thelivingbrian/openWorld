package main

import (
	"github.com/gorilla/websocket"
)

type Player struct {
	id        string
	stage     *Stage
	stageName string
	conn      *websocket.Conn
	x         int
	y         int
	actions   *Actions
	health    int
}

type Actions struct {
	space bool
}

func (player *Player) isDead() bool {
	return player.health <= 0
}

func placeOnStage(p *Player) {
	x := p.x
	y := p.y
	p.stage.tiles[y][x].addPlayer(p) // add p method
	p.stage.playerMap[p.id] = p      // needed?
	//fmt.Println(len(p.stage.playerMap))
	p.stage.markAllDirty()
}

func handleDeathOf(player *Player) {
	clinic := getClinic()
	player.health = 100
	player.stage = clinic
	player.stageName = "clinic"
	player.x = 2
	player.y = 2
	placeOnStage(player)
}

func updateScreenWithStarter(player *Player, html string, playerUpdates chan Update) {
	if player.isDead() {
		handleDeathOf(player)
		return
	}
	html += hudAsOutOfBound(player)
	playerUpdates <- Update{player, []byte(html)}
}

func updateScreenFromScratch(player *Player, playerUpdates chan Update) {
	if player.isDead() {
		handleDeathOf(player)
		return
	}
	playerUpdates <- Update{player, htmlFromPlayer(player)}
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
		currentTile.removePlayer(p.id)
		p.y = destY // Don't like this here, move to addPlayer?
		p.x = destX
		destTile.addPlayer(p)
		p.stage.markAllDirty()
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
