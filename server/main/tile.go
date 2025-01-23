package main

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Tile struct {
	material          Material
	playerMap         map[string]*Player
	playerMutex       sync.Mutex
	interactable      *Interactable
	interactableMutex sync.Mutex
	stage             *Stage
	teleport          *Teleport // Should/could be interactable? - questionable
	y                 int
	x                 int
	eventsInFlight    atomic.Int32
	powerUp           *PowerUp
	powerMutex        sync.Mutex
	money             int
	moneyMutex        sync.Mutex
	boosts            int
	htmlTemplate      string
	bottomText        string
}

type Teleport struct {
	destStage          string
	destY              int
	destX              int
	sourceStage        string
	confirmation       bool
	rejectInteractable bool
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
		htmlTemplate:   makeTileTemplate(mat, y, x),
		bottomText:     mat.DisplayText, // Pre-process needed *String to have option of null?
	}
}

////////////////////////////////////////////////
// HTML

func makeTileTemplate(mat Material, y, x int) string {
	tileCoord := fmt.Sprintf("%d-%d", y, x)
	cId := "c" + tileCoord // This is used to identify the entire square
	placeHold := "%s"      // later becomes player, interactable, svg, weather, and highlight boxes

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
					%s
				</div>`
	return fmt.Sprintf(template, cId, mat.CssColor, floor1css, floor2css, placeHold, placeHold, placeHold, ceil1css, ceil2css, placeHold, placeHold)
}

////////////////////////////////////////////////////////////
// Players

func (tile *Tile) addPlayerAndNotifyOthers(player *Player) {
	player.tileLock.Lock()
	defer player.tileLock.Unlock()
	tile.addLockedPlayertoTile(player)
	tile.stage.updateAllExcept(playerBox(tile), player)
}

func (tile *Tile) addLockedPlayertoTile(player *Player) {
	tile.playerMutex.Lock()
	defer tile.playerMutex.Unlock()

	itemChange := false
	if tile.bottomText != "" {
		player.updateBottomText(tile.bottomText)
	}
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
		// locks with transfer across stages
		go player.stage.updateAll(svgFromTile(tile))
	}

	// players tile lock should be held
	tile.playerMap[player.id] = player
	player.tile = tile

	if tile.teleport != nil {
		if tile.teleport.confirmation {
			player.menues["teleport"] = continueTeleporting(tile.teleport)
			turnMenuOnByName(player, "teleport")
		} else {
			// new routine prevents deadlock // still needed ? test.
			go player.applyTeleport(tile.teleport)
		}
	}
}

func (tile *Tile) removePlayerAndNotifyOthers(player *Player) (success bool) {
	success = tryRemovePlayer(tile, player)
	if success {
		tile.stage.updateAllExcept(playerBox(tile), player)
	} else {
		fmt.Println("WARN : FAILED TO REMOVE PLAYER :(")
	}
	return success
}

func tryRemovePlayer(tile *Tile, player *Player) bool {
	tile.playerMutex.Lock()
	defer tile.playerMutex.Unlock()

	_, foundOnTile := tile.playerMap[player.id]
	if !foundOnTile {
		return false
	}

	delete(tile.playerMap, player.id)
	return true
}

func (tile *Tile) getAPlayer() *Player {
	tile.playerMutex.Lock()
	defer tile.playerMutex.Unlock()
	for _, player := range tile.playerMap {
		return player
	}
	return nil
}

/////////////////////////////////////////////////////////////////////
// Damage

func (tile *Tile) damageAll(dmg int, initiator *Player) {
	fatalities := false
	for _, player := range tile.copyOfPlayers() {
		fatalities = damageTargetOnBehalfOf(player, initiator, dmg) || fatalities
	}
	if fatalities {
		tile.stage.updateAll(playerBox(tile))
	}
}

func (tile *Tile) copyOfPlayers() []*Player {
	players := make([]*Player, 0)
	tile.playerMutex.Lock()
	for _, player := range tile.playerMap {
		players = append(players, player)
	}
	tile.playerMutex.Unlock()
	return players
}

func damageTargetOnBehalfOf(target, initiator *Player, dmg int) bool {
	target.tangibilityLock.Lock()
	if !target.tangible {
		target.tangibilityLock.Unlock()
		return false
	}
	target.tangibilityLock.Unlock()
	if isInClinicOrInfirmary(target) {
		return false
	}
	if target.getTeamNameSync() == initiator.getTeamNameSync() {
		return false
	}

	location := target.getTileSync()
	fatal := damagePlayerAndHandleDeath(target, dmg)
	if fatal {
		initiator.incrementKillCount()
		initiator.incrementKillStreak()
		initiator.updateRecord()
		go initiator.world.db.saveKillEvent(location, initiator, target)
	}
	return fatal
}
func damagePlayerAndHandleDeath(player *Player, dmg int) bool {
	fatal := reduceHealthAndCheckFatal(player, dmg)
	if fatal {
		handleDeath(player)
	} else {
		player.updateInformation()
	}
	return fatal
}

func reduceHealthAndCheckFatal(player *Player, dmg int) bool {
	player.healthLock.Lock()
	defer player.healthLock.Unlock()
	oldHealth := player.health
	newHealth := oldHealth - dmg
	player.health = newHealth

	// negative health is invincibility, alternative is killstreak for killing a zombie
	fatal := oldHealth > 0 && newHealth <= 0
	return fatal
}

func isInClinicOrInfirmary(p *Player) bool {
	stagename := p.getStageNameSync()
	if stagename == "clinic" {
		return true
	}
	if strings.HasPrefix(stagename, "infirmary") {
		return true
	}
	return false
}

////////////////////////////////////////////////////////////////////////
//  Notify

func (tile *Tile) tryToNotifyAfter(delay int) {
	time.Sleep(time.Millisecond * time.Duration(delay))
	if tile.eventsInFlight.Add(-1) == 0 {
		// blue trsp20 for gloom
		tile.stage.updateAll(weatherBox(tile, ""))
	}
}

//////////////////////////////////////////////////////////////////////
// Interactables

func destroyFragileInteractable(tile *Tile, _ *Player) {
	// *Player is a placeholder for initiator/destroyer in future
	tile.interactableMutex.Lock()
	defer tile.interactableMutex.Unlock()
	if tile.interactable != nil && tile.interactable.fragile {
		tile.interactable = nil
		tile.stage.updateAll(interactableBox(tile))
	}
}

func destroyInteractable(tile *Tile, _ *Player) {
	// *Player is a placeholder for initiator/destroyer in future
	tile.interactableMutex.Lock()
	defer tile.interactableMutex.Unlock()
	if tile.interactable != nil {
		tile.interactable = nil
		tile.stage.updateAll(interactableBox(tile))
	}
}

func halveMoneyOf(player *Player) int {
	currentMoney := player.getMoneySync()
	newValue := currentMoney / 2
	player.setMoney(newValue)
	return newValue
}

func walkable(tile *Tile) bool {
	if tile == nil {
		return false
	}
	tile.interactableMutex.Lock()
	defer tile.interactableMutex.Unlock()

	// why?
	if tile.interactable == nil {
		return tile.material.Walkable
	} else {
		// pushable must (?) be walkable to prevent blocking of players and interactables in corners
		// non-pushable may still be walkable
		return tile.interactable.pushable || tile.interactable.walkable
	}
}

/////////////////////////////////////////////////////////////
// Utilities

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

func sliceOfTileToWeatherBoxes(tiles []*Tile, cssClass string) string {
	html := ``
	for _, tile := range tiles {
		html += weatherBox(tile, cssClass)
	}
	return html
}

func sliceOfTileToHighlightBoxes(tiles []*Tile, cssClass string) string {
	html := ``
	for _, tile := range tiles {
		html += oobHighlightBox(tile, cssClass)
	}
	return html
}

func everyOtherTileOnStage(tile *Tile) []*Tile {
	out := make([]*Tile, 0)
	for i := range tile.stage.tiles {
		for j := range tile.stage.tiles[i] {
			if tile != tile.stage.tiles[i][j] {
				out = append(out, tile.stage.tiles[i][j])
			}
		}
	}
	return out
}

/////////////////////////////////////////////////////////////////
// Observers / Item state

// / These need to get looked at (? mutex?)
func (tile *Tile) addPowerUpAndNotifyAll(shape [][2]int) {
	tile.powerUp = &PowerUp{shape}
	tile.stage.updateAll(svgFromTile(tile))
}

func (tile *Tile) addBoostsAndNotifyAll() {
	tile.boosts += 10
	tile.stage.updateAll(svgFromTile(tile))
}

func (tile *Tile) addMoneyAndNotifyAll(amount int) {
	tile.addMoneyAndNotifyAllExcept(amount, nil)
}

func (tile *Tile) addMoneyAndNotifyAllExcept(amount int, player *Player) {
	tile.moneyMutex.Lock()
	defer tile.moneyMutex.Unlock()
	tile.money += amount
	tile.stage.updateAllExcept(svgFromTile(tile), player)
}
