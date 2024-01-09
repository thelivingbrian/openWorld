package main

import (
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
	stage       *Stage
	teleport    *Teleport
	// Coords
	y int
	x int
	// Display
	currentCssClass  string
	originalCssClass string
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
	return &Tile{mat, make(map[string]*Player), sync.Mutex{}, nil, nil, y, x, mat.CssClassName, mat.CssClassName}
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
		tile.currentCssClass = tile.originalCssClass //.material.CssClassName
	}

	return htmlFromTile(tile)
}

func (tile *Tile) addPlayer(player *Player) string {
	if tile.teleport != nil {
		tile.stage.removePlayerById(player.id)
		tile.removePlayer(player.id)

		/*player.y = tile.teleport.destY
		player.x = tile.teleport.destX
		player.stageName = tile.teleport.destStage

		stageMutex.Lock()
		existingStage, stageExists := stageMap[player.stageName]
		if !stageExists {
			fmt.Println("New Stage")
			existingStage = createStageAndHandleUpdates(player.stageName)
		}
		stageMutex.Unlock()
		*/

		existingStage := getStageByName(tile.teleport.destStage)
		placeOnTile(player, existingStage.tiles[tile.teleport.destY][tile.teleport.destX])
		//player.stage = existingStage
		//placeOnStage(player)
	} else {
		tile.playerMutex.Lock()
		tile.playerMap[player.id] = player
		tile.playerMutex.Unlock()
		player.y = tile.y
		player.x = tile.x
		tile.currentCssClass = cssClassFromHealth(player)
	}
	return htmlFromTile(tile)
}

func placeOnTile(player *Player, tile *Tile) {
	player.stage = tile.stage
	tile.addPlayer(player)
	tile.stage.addPlayer(player)
	tile.stage.markAllDirty()
}

func cssClassFromHealth(player *Player) string {
	if player.health >= 80 {
		return "green"
	}
	if player.health >= 60 {
		return "lime"
	}
	if player.health >= 40 {
		return "yellow"
	}
	if player.health >= 20 {
		return "orange"
	}
	if player.health >= 0 {
		return "red"
	}
	return "blue" //shouldn't happen but want to be visible
}

func (tile *Tile) damageAll(dmg int) {
	first := true
	for _, player := range tile.playerMap {
		player.health += -dmg
		tile.currentCssClass = cssClassFromHealth(player)
		if player.isDead() {
			tile.removePlayer(player.id)
			tile.stage.removePlayerById(player.id)
			respawn(player)
		}
		if first {
			first = false
			tile.stage.updateAll(htmlFromTile(tile))
		}
	}
}
