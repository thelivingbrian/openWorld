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

func moveNorth(stage *Stage, p *Player) {
	x := p.x
	y := p.y
	nextTile := &stage.tiles[y-1][x]
	if walkable(nextTile) {
		currentTile := &stage.tiles[y][x]
		currentTile.removePlayer(p.id)

		nextTile.addPlayer(p)

		p.y = y - 1
		stage.markAllDirty()
	} else {
		//nop
	}
}

func moveSouth(stage *Stage, p *Player) {
	x := p.x
	y := p.y

	nextTile := &stage.tiles[y+1][x]
	if walkable(nextTile) {
		currentTile := &stage.tiles[y][x]
		currentTile.removePlayer(p.id)

		nextTile.addPlayer(p)

		p.y = y + 1
		stage.markAllDirty()
	} else {
		//nop
	}
}

func moveEast(stage *Stage, p *Player) {
	x := p.x
	y := p.y
	nextTile := &stage.tiles[y][x+1]
	if walkable(nextTile) {
		currentTile := &stage.tiles[y][x]
		currentTile.removePlayer(p.id)

		nextTile.addPlayer(p)

		p.x = x + 1
		stage.markAllDirty()
	} else {
		//nop
	}
}

func moveWest(stage *Stage, p *Player) {
	x := p.x
	y := p.y
	nextTile := &stage.tiles[y][x-1]
	if walkable(nextTile) {
		currentTile := &stage.tiles[y][x]
		currentTile.removePlayer(p.id)

		nextTile.addPlayer(p)

		p.x = x - 1
		stage.markAllDirty()
	} else {
		//nop
	}
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

func getClinic() *Stage {
	clinic := stageFromArea("clinic")
	return &clinic
}
