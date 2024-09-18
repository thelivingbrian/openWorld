package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Teleport struct {
	destStage string
	destY     int
	destX     int
}

type Interactable struct {
	pushable bool
	cssClass string
	// reactive
	// fragile
}

type Tile struct {
	material          Material
	playerMap         map[string]*Player
	playerMutex       sync.Mutex
	interactable      *Interactable
	interactableMutex sync.Mutex
	stage             *Stage
	teleport          *Teleport // Should/could be interactable?
	y                 int
	x                 int
	eventsInFlight    atomic.Int32
	powerUp           *PowerUp
	powerMutex        sync.Mutex
	money             int
	boosts            int
	htmlTemplate      string
	// messageText
}

func newTile(mat Material, y int, x int, defaultTileColor string) *Tile {
	if mat.CssColor == "" {
		mat.CssColor = defaultTileColor
	}
	return &Tile{
		material:       mat,
		playerMap:      make(map[string]*Player),
		playerMutex:    sync.Mutex{},
		stage:          nil,
		teleport:       nil,
		y:              y,
		x:              x,
		eventsInFlight: atomic.Int32{},
		powerUp:        nil,
		powerMutex:     sync.Mutex{},
		money:          0,
		htmlTemplate:   makeTileTemplate(mat, y, x)}
}

func makeTileTemplate(mat Material, y, x int) string {
	tileCoord := fmt.Sprintf("%d-%d", y, x)
	cId := "c" + tileCoord // This is used to identify the entire square
	hId := "t" + tileCoord // This is used to identify the top highlight box
	placeHold := "%s"      // later becomes user, player, interactable, and svg boxes

	floor1css := ""
	if mat.Floor1Css != "" {
		floor1css = fmt.Sprintf(`<div class="box floor1 %s"></div>`, mat.Floor1Css)
	}

	floor2css := ""
	if mat.Floor2Css != "" {
		floor2css = fmt.Sprintf(`<div class="box floor2 %s"></div>`, mat.Floor2Css)
	}

	ceil1css := ""
	if mat.Ceiling1Css != "" {
		ceil1css = fmt.Sprintf(`<div class="box ceiling1 %s"></div>`, mat.Ceiling1Css)
	}

	ceil2css := ""
	if mat.Ceiling2Css != "" {
		ceil2css = fmt.Sprintf(`<div class="box ceiling2 %s"></div>`, mat.Ceiling2Css)
	}

	template := `<div id="%s" class="grid-square %s">				
					%s
					%s
					%s
					%s
					%s
					%s
					%s
					%s
					<div id="%s" class="box top"></div>
				</div>`
	return fmt.Sprintf(template, cId, mat.CssColor, floor1css, floor2css, placeHold, placeHold, placeHold, placeHold, ceil1css, ceil2css, hId)
}

// newTile w/ teleport?

func (tile *Tile) addPlayerAndNotifyOthers(player *Player) {
	tile.addPlayer(player)
	tile.stage.updateAllExcept(playerBox(tile), player)
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
	} else {
		// Add on new stage // Not always a new stage?
		player.removeFromStage()
		player.applyTeleport(tile.teleport)
	}
}

func (tile *Tile) removePlayerAndNotifyOthers(player *Player) {
	tile.removePlayer(player.id)
	// have take a tile not player?
	tile.stage.updateAllExcept(playerBox(tile), player)
}

func (tile *Tile) removePlayer(playerId string) {
	tile.playerMutex.Lock()
	delete(tile.playerMap, playerId)
	tile.playerMutex.Unlock() // Defer instead?
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
		tile.stage.updateAllWithHud([]*Tile{tile})
	}
}

func (tile *Tile) damageAll(dmg int, initiator *Player) {
	survivors := false
	for _, player := range tile.playerMap {
		if player == initiator {
			continue // Race condition nonsense, can this still happen?
		}
		survived := player.addToHealth(-dmg)
		survivors = survivors || survived
		if !survived {
			initiator.incrementKillStreak()
			go player.world.db.saveKillEvent(tile, initiator, player) // Maybe should just pass in required fields?
		}
	}
	if survivors {
		tile.stage.updateAll(playerBox(tile))
	}
}

func halveMoneyOf(player *Player) int {
	currentMoney := player.getMoneySync()
	newValue := currentMoney / 2
	player.setMoney(newValue)
	return newValue
}

func walkable(tile *Tile) bool {
	return tile.material.Walkable
}

/// These need to get looked at (? mutex?)

func (tile *Tile) addPowerUpAndNotifyAll(shape [][2]int) {
	tile.powerUp = &PowerUp{shape}
	tile.stage.updateAll(svgFromTile(tile))
}

func (tile *Tile) addBoostsAndNotifyAll() {
	tile.boosts += 10
	tile.stage.updateAll(svgFromTile(tile))
}

func (tile *Tile) addMoneyAndNotifyAll(amount int) {
	tile.money += amount
	tile.stage.updateAll(svgFromTile(tile))
}

///

func cssClassFromHealth(player *Player) string {
	// >120 indicator
	// Middle range choosen color? or only in safe
	if player.health > 50 {
		return "red r0"
	}
	if player.health >= 0 {
		return "dark-red r0"
	}
	return "blue" // shouldn't happen but want to be visible
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

func validityByAxis(y int, x int, tiles [][]*Tile) (bool, bool) {
	invalidY, invalidX := false, false
	if y < 0 || y >= len(tiles) {
		invalidY = true
	}
	if x < 0 || x >= len(tiles[0]) { // Not the best, assumes rectangular grid
		invalidX = true
	}
	return invalidY, invalidX
}

func mapOfTileToArray(m map[*Tile]bool) []*Tile {
	out := make([]*Tile, 0)
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
