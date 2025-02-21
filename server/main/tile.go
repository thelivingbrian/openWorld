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
	boostsMutex       sync.Mutex
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

	// technically can race, e.g. with interactable reaction
	if tile.bottomText != "" {
		player.updateBottomText(tile.bottomText)
	}

	if tile.collectItemsForPlayer(player) {
		sendSoundToPlayer(player, "money")
		player.stage.updateAll(svgFromTile(tile))
	}

	// player's tile lock should be held
	tile.playerMap[player.id] = player
	player.tile = tile

	if tile.teleport != nil {
		if tile.teleport.confirmation {
			player.menues["teleport"] = continueTeleporting(tile.teleport)
			turnMenuOnByName(player, "teleport")
		} else {
			// new routine prevents deadlock
			go player.applyTeleport(tile.teleport)
		}
	}
}

func (tile *Tile) collectItemsForPlayer(player *Player) bool {
	itemChange := false
	// Single item mutex?
	tile.powerMutex.Lock()
	defer tile.powerMutex.Unlock()
	tile.moneyMutex.Lock()
	defer tile.moneyMutex.Unlock()
	tile.boostsMutex.Lock()
	defer tile.boostsMutex.Unlock()
	if tile.powerUp != nil {
		powerUp := tile.powerUp
		tile.powerUp = nil
		addPowerToStack(player, powerUp)
		itemChange = true
	}
	if tile.money != 0 {
		player.addMoneyAndUpdate(tile.money)
		tile.money = 0
		itemChange = true
	}
	if tile.boosts > 0 {
		player.addBoostsAndUpdate(tile.boosts)
		tile.boosts = 0
		itemChange = true
	}
	return itemChange
}

func (tile *Tile) removePlayerAndNotifyOthers(player *Player) (success bool) {
	success = tryRemovePlayer(tile, player)
	if success {
		tile.stage.updateAllExcept(playerBox(tile), player)
	} else {
		// Possible under what circumstance ?
		logger.Warn().Msg("WARN : FAILED TO REMOVE PLAYER :(")
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
	if target == initiator {
		return false
	}
	target.tangibilityLock.Lock()
	defer target.tangibilityLock.Unlock()
	if !target.tangible {
		return false
	}
	if isInClinicOrInfirmary(target) {
		return false
	}
	if target.getTeamNameSync() == initiator.getTeamNameSync() {
		return false
	}

	location := target.getTileSync()
	fatal := damagePlayerAndHandleDeath(target, dmg) // Death can't handle until return to tangibility
	if fatal {
		initiator.incrementKillCount()
		initiator.incrementKillStreak()
		initiator.updateRecord()
		go initiator.world.db.saveKillEvent(location, initiator, target)
	}
	return fatal
}
func damagePlayerAndHandleDeath(player *Player, dmg int) bool {
	flashBackgroundColor(player, "twilight")
	fatal := reduceHealthAndCheckFatal(player, dmg)
	if fatal {
		handleDeath(player)
	} else {
		player.updatePlayerHud()
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
		tile.stage.updateAll(lockedInteractableBox(tile))
	}
}

func destroyInteractable(tile *Tile, _ *Player) {
	// *Player is a placeholder for initiator/destroyer in future
	tile.interactableMutex.Lock()
	defer tile.interactableMutex.Unlock()
	if tile.interactable != nil {
		tile.interactable = nil
		tile.stage.updateAll(lockedInteractableBox(tile))
	}
}

func trySetInteractable(tile *Tile, i *Interactable) bool {
	ownLock := tile.interactableMutex.TryLock()
	if !ownLock {
		return false
	}
	defer tile.interactableMutex.Unlock()
	if tile.interactable != nil {
		return false
	}
	tile.interactable = i
	return true
}

func tryGetInteractable(tile *Tile) *Interactable {
	ownLock := tile.interactableMutex.TryLock()
	if !ownLock {
		return nil
	}
	defer tile.interactableMutex.Unlock()
	return tile.interactable
}

/////////////////////////////////////////////////////////////
// Utilities

func walkable(tile *Tile) bool {
	if tile == nil {
		return false
	}
	if !tile.material.Walkable {
		return false
	}
	tile.interactableMutex.Lock()
	defer tile.interactableMutex.Unlock()
	if tile.interactable != nil {
		return tile.interactable.pushable || tile.interactable.walkable

	}
	return true
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

func getVanNeumannNeighborsOfTile(tile *Tile) []*Tile {
	out := make([]*Tile, 0)
	for _, yOff := range []int{-1, 1} {
		y := tile.y + yOff
		x := tile.x
		if y >= 0 && y < len(tile.stage.tiles) && x >= 0 && x < len(tile.stage.tiles[y]) {
			out = append(out, tile.stage.tiles[y][x])
		}
	}
	for _, xOff := range []int{-1, 1} {
		y := tile.y
		x := tile.x + xOff
		if y >= 0 && y < len(tile.stage.tiles) && x >= 0 && x < len(tile.stage.tiles[y]) {
			out = append(out, tile.stage.tiles[y][x])
		}
	}
	return out
}

func getTilesInRadius(tile *Tile, r int) []*Tile {
	out := make([]*Tile, 0)
	for i := -r; i <= r; i++ {
		for j := -r; j <= r; j++ {
			y := tile.y + i
			x := tile.x + j
			if y >= 0 && y < len(tile.stage.tiles) && x >= 0 && x < len(tile.stage.tiles[y]) {
				out = append(out, tile.stage.tiles[y][x])
			}
		}
	}
	return out
}

func damageAndIndicate(tiles []*Tile, initiator *Player, stage *Stage, damage int) {
	for _, tile := range tiles {
		tile.damageAll(damage, initiator)
		destroyFragileInteractable(tile, initiator)
		tile.eventsInFlight.Add(1)
		go tile.tryToNotifyAfter(100)
	}
	damageBoxes := sliceOfTileToWeatherBoxes(tiles, randomFieryColor())
	stage.updateAll(damageBoxes)
}

/////////////////////////////////////////////////////////////////
// Observers / Item state

// / These need to get looked at (? mutex?)
func (tile *Tile) addPowerUpAndNotifyAll(shape [][2]int) {
	tile.powerMutex.Lock()
	tile.powerUp = &PowerUp{shape}
	tile.powerMutex.Unlock()
	tile.stage.updateAll(svgFromTile(tile))
}

func (tile *Tile) addBoostsAndNotifyAll() {
	tile.boostsMutex.Lock()
	tile.boosts += 10
	tile.boostsMutex.Unlock()
	tile.stage.updateAll(svgFromTile(tile))
}

func (tile *Tile) addMoneyAndNotifyAll(amount int) {
	tile.addMoneyAndNotifyAllExcept(amount, nil)
}

func (tile *Tile) addMoneyAndNotifyAllExcept(amount int, player *Player) {
	tile.moneyMutex.Lock()
	tile.money += amount
	tile.moneyMutex.Unlock()
	tile.stage.updateAllExcept(svgFromTile(tile), player)
}
