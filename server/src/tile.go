package main

import (
	"fmt"
	"sync"
)

type Teleport struct {
	destStage string
	destY     int
	destX     int
}

type Tile struct {
	material    Material
	playerMap   map[string]*Player
	playerMutex sync.Mutex
	Teleport    *Teleport
	// Items and coords?
}

func colorOf(tile *Tile) string {
	if len(tile.playerMap) > 0 {
		return "blue"
	}

	return tile.material.CssClassName
}

func colorArray(row []Tile) []string {
	var output []string = make([]string, len(row))
	for i := range row {
		output[i] = colorOf(&row[i])
	}
	return output
}

func newTile(mat Material) Tile {
	return Tile{mat, make(map[string]*Player), sync.Mutex{}, nil}
}

// newTile w/ teleport?

func walkable(tile *Tile) bool {
	return tile.material.Walkable
}

func (tile *Tile) removePlayer(playerId string) {
	tile.playerMutex.Lock()
	delete(tile.playerMap, playerId)
	tile.playerMutex.Unlock()
}

func (tile *Tile) addPlayer(player *Player) {
	if tile.Teleport != nil {
		player.y = tile.Teleport.destY
		player.x = tile.Teleport.destX
		player.stageName = tile.Teleport.destStage

		stageMutex.Lock()
		existingStage, stageExists := stageMap[player.stageName]
		if !stageExists {
			fmt.Println("New Stage")
			newStage := createStageByName(player.stageName)
			stagePtr := &newStage
			stageMap[player.stageName] = stagePtr
			existingStage = stagePtr
		}
		stageMutex.Unlock()

		player.stage = existingStage
		placeOnStage(player)
	} else {
		tile.playerMutex.Lock()
		tile.playerMap[player.id] = player
		tile.playerMutex.Unlock()
	}
}
