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

func (tile *Tile) addPlayerAndNotifyOthers(player *Player) {
	tile.addPlayer(player)
	tile.stage.updateAllExcept(htmlFromTile(tile), player)
}

func (tile *Tile) addPlayer(player *Player) {
	if tile.teleport == nil {
		tile.playerMutex.Lock()
		tile.playerMap[player.id] = player
		tile.playerMutex.Unlock()
		player.y = tile.y
		player.x = tile.x
		player.tile = tile
		tile.currentCssClass = cssClassFromHealth(player)
	} else {
		// Add on new stage // Not always a new stage?
		player.removeFromStage()
		player.applyTeleport(tile.teleport)
	}
}

func (tile *Tile) removePlayer(playerId string) {
	tile.playerMutex.Lock()
	delete(tile.playerMap, playerId)
	tile.playerMutex.Unlock() // Defer instead?

	if len(tile.playerMap) == 0 {
		tile.currentCssClass = tile.originalCssClass
	}
	// else need to find another players health
}

func (tile *Tile) removePlayerAndNotifyOthers(player *Player) {
	tile.removePlayer(player.id)
	tile.stage.updateAllExcept(htmlFromTile(tile), player)
}

func (tile *Tile) damageAll(dmg int, initiator *Player) {
	first := true
	for _, player := range tile.playerMap {
		survived := player.addToHealth(-dmg)
		tile.currentCssClass = cssClassFromHealth(player)
		if !survived {
			tile.currentCssClass = tile.originalCssClass
			// Copying these structs in by vaue makes a compiler warning
			// Maybe should just pass in required fields?
			// Or this is fine event though it isn't really a snapshot
			go player.world.db.saveKillEvent(tile, initiator, player)
		}
		if first {
			first = !survived // Gross but this ensures that surviving players aren't hidden by death
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
