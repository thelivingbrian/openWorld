package main

import (
	"fmt"

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
	p.stage.tiles[y][x].addPlayer(p)
	p.stage.playerMap[p.id] = p
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

func updateScreenWithStarter(player *Player, html string) {
	if player.isDead() {
		handleDeathOf(player)
		return
	}
	html += hudAsOutOfBound(player)
	player.stage.updates <- Update{player, []byte(html)}
}

func updateScreenFromScratch(player *Player) {
	if player.isDead() {
		handleDeathOf(player)
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
		currentStage := p.stage
		oobPrevious := currentTile.removePlayer(p.id)
		//p.y = destY // Don't like this here, move to addPlayer?
		//p.x = destX
		oobNext := destTile.addPlayer(p)
		fmt.Println(currentStage.name)
		currentStage.updateAll(oobPrevious + oobNext)
		//p.stage.markAllDirty()
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
