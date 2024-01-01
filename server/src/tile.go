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
	//Game properties
	material    Material
	playerMap   map[string]*Player
	playerMutex sync.Mutex
	Teleport    *Teleport
	// Items and coords?
	// Display
	CurrentCssClass string
}

func colorOf(tile *Tile) string {
	return tile.CurrentCssClass // Maybe like the old way better with the player count logic here
}

func colorArray(row []Tile) []string {
	var output []string = make([]string, len(row))
	for i := range row {
		output[i] = colorOf(&row[i])
	}
	return output
}

func newTile(mat Material) Tile {
	return Tile{mat, make(map[string]*Player), sync.Mutex{}, nil, mat.CssClassName}
}

// newTile w/ teleport?

func walkable(tile *Tile) bool {
	return tile.material.Walkable
}

func (tile *Tile) removePlayer(playerId string) {
	tile.playerMutex.Lock()
	delete(tile.playerMap, playerId)
	tile.playerMutex.Unlock()

	if len(tile.playerMap) == 0 {
		tile.CurrentCssClass = tile.material.CssClassName
	}
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
		tile.CurrentCssClass = "blue"
	}
}
