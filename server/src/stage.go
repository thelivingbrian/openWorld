package main

import (
	"fmt"
	"sync"
)

type Stage struct {
	tiles       [][]Tile
	playerMap   map[string]*Player
	playerMutex sync.Mutex
	name        string
}

func (stage *Stage) markAllDirty() {
	for _, player := range stage.playerMap {
		updateScreen(player)
	}
}

func validCoordinate(y int, x int, tiles [][]Tile) bool {
	if y < 0 || y >= len(tiles) {
		return false
	}
	if x < 0 || x >= len(tiles[y]) {
		return false
	}
	return true
}

func move(p *Player, yOffset int, xOffset int) {
	destY := p.y + yOffset
	destX := p.x + xOffset
	if validCoordinate(destY, destX, p.stage.tiles) && walkable(&p.stage.tiles[destY][destX]) {
		currentTile := &p.stage.tiles[p.y][p.x]
		destTile := &p.stage.tiles[destY][destX]
		currentTile.removePlayer(p.id)
		p.y = destY
		p.x = destX
		destTile.addPlayer(p)
		p.stage.markAllDirty()
	}
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

func (stage *Stage) damageAt(coords [][2]int) {
	for _, pair := range coords {
		for _, player := range stage.playerMap { // This is really stupid right? The tile has a playermap?
			if pair[0] == player.y && pair[1] == player.x {
				player.health += -50
				if !player.isAlive() {
					fmt.Println(player.id + " has died")

					deadPlayerTile := &stage.tiles[pair[0]][pair[1]]
					deadPlayerTile.playerMutex.Lock() // break into function, no high level mutexing(?)
					delete(deadPlayerTile.playerMap, player.id)
					deadPlayerTile.playerMutex.Unlock()

					stage.playerMutex.Lock()
					delete(stage.playerMap, player.id)
					stage.playerMutex.Unlock()

					stage.markAllDirty()
					updateScreen(player)
				}
			}
		}
	}
}

func getStageByName(name string) *Stage {
	stageMutex.Lock()
	existingStage, stageExists := stageMap[name]
	if !stageExists {
		newStage := createStageByName(name)
		stagePtr := &newStage
		stageMap[name] = stagePtr
		existingStage = stagePtr
	}
	stageMutex.Unlock()
	return existingStage
}

func getClinic() *Stage {
	return getStageByName("clinic")
}
