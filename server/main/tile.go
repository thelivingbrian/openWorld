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
	teleport    *Teleport
	// Coords
	y int
	x int
	// Display
	currentCssClass string
}

func colorOf(tile *Tile) string {
	return tile.currentCssClass // Maybe like the old way better with the player count logic here
}

func colorArray(row []Tile) []string {
	var output []string = make([]string, len(row))
	for i := range row {
		output[i] = colorOf(&row[i])
	}
	return output
}

func newTile(mat Material, y int, x int) *Tile {
	return &Tile{mat, make(map[string]*Player), sync.Mutex{}, nil, y, x, mat.CssClassName}
}

// newTile w/ teleport?

func walkable(tile *Tile) bool {
	return tile.material.Walkable
}

func (tile *Tile) removePlayer(playerId string) string {
	tile.playerMutex.Lock()
	delete(tile.playerMap, playerId)
	tile.playerMutex.Unlock()

	if len(tile.playerMap) == 0 {
		tile.currentCssClass = tile.material.CssClassName
	}

	return htmlFromTile(tile)
}

func (tile *Tile) addPlayer(player *Player) string {
	if tile.teleport != nil {
		player.y = tile.teleport.destY
		player.x = tile.teleport.destX
		player.stageName = tile.teleport.destStage

		stageMutex.Lock()
		existingStage, stageExists := stageMap[player.stageName]
		if !stageExists {
			fmt.Println("New Stage")
			existingStage = createStageByName(player.stageName)
			//stagePtr := &newStage
			//stageMap[player.stageName] = stagePtr
			//existingStage = stagePtr
		}
		stageMutex.Unlock()

		player.stage = existingStage
		placeOnStage(player)
	} else {
		tile.playerMutex.Lock()
		tile.playerMap[player.id] = player
		tile.playerMutex.Unlock()
		player.y = tile.y
		player.x = tile.x
		tile.currentCssClass = "blue"
	}
	return htmlFromTile(tile)
}
