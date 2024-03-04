package main

import (
	"sync"
	"sync/atomic"
	"time"
)

type Teleport struct {
	destStage string
	destY     int
	destX     int
}

type Tile struct {
	material        Material
	playerMap       map[string]*Player
	playerMutex     sync.Mutex
	stage           *Stage
	teleport        *Teleport
	y               int
	x               int
	currentCssClass string
	//originalCssClass string
	eventsInFlight atomic.Int32
	powerUp        *PowerUp
	powerMutex     sync.Mutex
	money          int
	boosts         int
}

func newTile(mat Material, y int, x int) *Tile {
	return &Tile{
		material:    mat,
		playerMap:   make(map[string]*Player),
		playerMutex: sync.Mutex{},
		stage:       nil,
		teleport:    nil,
		y:           y,
		x:           x,
		//originalCssClass: mat.CssColor,
		currentCssClass: mat.CssColor,
		eventsInFlight:  atomic.Int32{},
		powerUp:         nil,
		powerMutex:      sync.Mutex{},
		money:           0}
}

// newTile w/ teleport?

func (tile *Tile) addPlayerAndNotifyOthers(player *Player) {
	tile.addPlayer(player)
	tile.stage.updateAllExcept(playerBox(tile), player)
	//tile.stage.updateAllWithHudExcept(player, []*Tile{tile})
}

func (tile *Tile) addPlayer(player *Player) {
	itemChange := false
	if tile.powerUp != nil {
		// This should be mutexed I think
		powerUp := tile.powerUp
		tile.powerUp = nil
		player.actions.spaceStack.push(powerUp)
		itemChange = true
	}
	if tile.money != 0 {
		// I tex you tex
		player.setMoney(player.money + tile.money)
		tile.money = 0
		itemChange = true
	}
	if tile.boosts > 0 {
		// We all tex
		player.addBoosts(tile.boosts)
		tile.boosts = 0
		itemChange = true
	}
	if itemChange {
		player.stage.updateAll(svgFromTile(tile))
	}
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

func (tile *Tile) removePlayerAndNotifyOthers(player *Player) {
	tile.removePlayer(player.id)
	tile.stage.updateAllExcept(playerBox(tile), player)
	//tile.stage.updateAllWithHudExcept(player, []*Tile{tile})
}

func (tile *Tile) removePlayer(playerId string) {
	tile.playerMutex.Lock()
	delete(tile.playerMap, playerId)
	tile.playerMutex.Unlock() // Defer instead?

	samplePlayerRemaining := tile.getAPlayer()

	if samplePlayerRemaining == nil {
		tile.currentCssClass = tile.material.CssColor
		//fmt.Println(tile.material.CssColor)
	} else {
		tile.currentCssClass = cssClassFromHealth(samplePlayerRemaining)
	}
}

func (tile *Tile) getAPlayer() *Player {
	tile.playerMutex.Lock()
	defer tile.playerMutex.Unlock()
	for _, player := range tile.playerMap {
		return player
	}
	return nil
}

func (tile *Tile) incrementAndReturnIfFirst() *Tile {
	if tile.eventsInFlight.Load() == 0 {
		tile.eventsInFlight.Add(1)
		return tile
	} else {
		tile.eventsInFlight.Add(1)
		return nil
	}
}

func (tile *Tile) tryToNotifyAfter(delay int) {
	time.Sleep(time.Millisecond * time.Duration(delay))
	tile.eventsInFlight.Add(-1)
	if tile.eventsInFlight.Load() == 0 {
		// return string instead?
		//tile.stage.updateAll(htmlFromTile(tile))
		tile.stage.updateAllWithHud([]*Tile{tile})
	}
}

func (tile *Tile) damageAll(dmg int, initiator *Player) {
	first := true
	for _, player := range tile.playerMap {
		if player == initiator {
			continue // Race condition nonsense
		}
		survived := player.addToHealth(-dmg)
		tile.currentCssClass = cssClassFromHealth(player)
		if !survived {
			tile.currentCssClass = tile.material.CssColor
			tile.money += player.money / 2 // Use Observer, return diff
			player.money = player.money / 2
			// Maybe should just pass in required fields?
			go player.world.db.saveKillEvent(tile, initiator, player)
		}
		if first {
			first = !survived // Gross but this ensures that surviving players aren't hidden by death
			// Does multiple updates could be improved
			tile.stage.updateAllWithHudExcept(player, []*Tile{tile})
		}
	}
}

func walkable(tile *Tile) bool {
	return tile.material.Walkable
}

func (tile *Tile) addPowerUpAndNotifyAll(player *Player, shape [][2]int) {
	tile.powerUp = &PowerUp{shape, [4]int{100, 100, 100, 100}}
	html := htmlFromTile(tile)
	tile.stage.updateAllWithHudExcept(player, []*Tile{tile})
	updateOne(html, player) // Hides player token
}

func (tile *Tile) addBoostsAndNotifyAll(player *Player) {
	//fmt.Println("Adding Boost")
	tile.boosts += 5
	html := htmlFromTile(tile)
	tile.stage.updateAllWithHudExcept(player, []*Tile{tile})
	updateOne(html, player)

}

func cssClassFromHealth(player *Player) string {
	// >120 indicator
	// Middle range choosen color? or only in safe
	if player.health > 50 {
		return "red"
	}
	if player.health >= 0 {
		return "dark-red"
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

func mapOfTileToArray(m map[*Tile]bool) []*Tile {
	out := make([]*Tile, len(m))
	for tile := range m {
		out = append(out, tile)
	}
	return out
}

func sliceOfTileToColoredOoB(tiles []*Tile, cssClass string) string {
	html := ``
	for _, tile := range tiles {
		html += oobHighlightBox(tile, cssClass)
	}
	return html
}

func colorOf(tile *Tile) string {
	return tile.currentCssClass
}

// Used?
func colorArray(row []Tile) []string {
	var output []string = make([]string, len(row))
	for i := range row {
		output[i] = colorOf(&row[i])
	}
	return output
}
