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
	material         Material
	playerMap        map[string]*Player
	playerMutex      sync.Mutex
	stage            *Stage
	teleport         *Teleport
	y                int
	x                int
	currentCssClass  string
	originalCssClass string
}

func newTile(mat Material, y int, x int) *Tile {
	return &Tile{mat, make(map[string]*Player), sync.Mutex{}, nil, nil, y, x, mat.CssClassName, mat.CssClassName}
}

// newTile w/ teleport?

func (tile *Tile) addPlayer(player *Player) string {
	if tile.teleport != nil {
		// Add on new stage // Not always a new stage?
		player.removeFromStage()
		player.applyTeleport(tile.teleport)
		// player apply teleport
		// Use place on stage have it do tile stuff
		//existingStage := getStageByName(tile.teleport.destStage)
		//player.placeOnTile(existingStage.tiles[tile.teleport.destY][tile.teleport.destX])
	} else {
		tile.playerMutex.Lock()
		tile.playerMap[player.id] = player
		tile.playerMutex.Unlock()
		player.y = tile.y
		player.x = tile.x
		player.tile = tile
		tile.currentCssClass = cssClassFromHealth(player)
	}
	return htmlFromTile(tile)
}

func (tile *Tile) removePlayer(playerId string) string {
	tile.playerMutex.Lock()
	delete(tile.playerMap, playerId)
	tile.playerMutex.Unlock()

	if len(tile.playerMap) == 0 {
		tile.currentCssClass = tile.originalCssClass
	}

	return htmlFromTile(tile)
}

func (tile *Tile) damageAll(dmg int) {
	first := true
	for _, player := range tile.playerMap {
		player.health += -dmg
		updateOne(divPlayerInformation(player), player)
		tile.currentCssClass = cssClassFromHealth(player) // Sketchy?
		// Observer for player health handles all of this
		if player.isDead() {
			// Remove dead player from tile()
			tile.removePlayer(player.id)
			tile.stage.removePlayerById(player.id)
			respawn(player)
		}
		if first {
			first = false
			tile.stage.updateAllExcept(htmlFromTile(tile), player)
		}
	}
}

func walkable(tile *Tile) bool {
	return tile.material.Walkable
}

func cssClassFromHealth(player *Player) string {
	// >120 indicator
	// Middle range choosen color? or only in safe
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

func validCoordinate(y int, x int, tiles [][]*Tile) bool {
	if y < 0 || y >= len(tiles) {
		return false
	}
	if x < 0 || x >= len(tiles[y]) {
		return false
	}
	return true
}

func mapOfTileToOoB(m map[*Tile]bool) string {
	html := ``
	for tile := range m {
		html += htmlFromTile(tile)
	}
	return html
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
