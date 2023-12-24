package main

import (
	"sync"
)

type Tile struct {
	material    int
	playerMap   map[string]*Player
	playerMutex sync.Mutex
	// Items and coords?
}

func colorOf(tile *Tile) string {
	if len(tile.playerMap) > 0 {
		return "blue"
	}
	if tile.material == 0 {
		return "half-gray"
	}
	return ""
}

func colorArray(row []Tile) []string {
	var output []string = make([]string, len(row))
	for i := range row {
		output[i] = colorOf(&row[i])
	}
	return output
}

func newTile(mat int) Tile {
	return Tile{mat, make(map[string]*Player), sync.Mutex{}}
}

func walkable(tile *Tile) bool {
	return tile.material > 50
}

func (tile *Tile) removePlayer(playerId string) {
	tile.playerMutex.Lock()
	delete(tile.playerMap, playerId)
	tile.playerMutex.Unlock()
}

func (tile *Tile) addPlayer(player *Player) {
	tile.playerMutex.Lock()
	tile.playerMap[player.id] = player
	tile.playerMutex.Unlock()
}
