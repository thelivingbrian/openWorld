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
	characterMap      map[string]Character
	CharacterMutex    sync.Mutex
	interactable      *Interactable
	interactableMutex sync.Mutex
	stage             *Stage
	teleport          *Teleport // Should/could be interactable? - questionable
	y                 int
	x                 int
	eventsInFlight    atomic.Int32
	itemMutex         sync.Mutex
	powerUp           *PowerUp
	money             int
	boosts            int
	quickSwapTemplate string
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

func newTile(mat Material, y int, x int) *Tile {
	return &Tile{
		material:          mat,
		characterMap:      make(map[string]Character),
		CharacterMutex:    sync.Mutex{},
		stage:             nil,
		teleport:          nil,
		y:                 y,
		x:                 x,
		eventsInFlight:    atomic.Int32{},
		itemMutex:         sync.Mutex{},
		powerUp:           nil,
		boosts:            0,
		money:             0,
		quickSwapTemplate: makeQuickSwapTemplate(mat, y, x),
		bottomText:        mat.DisplayText, // Pre-process needed *String to have option of null?
	}
}

////////////////////////////////////////////////
// Quickswaps

func makeQuickSwapTemplate(mat Material, y, x int) string {
	// weird pattern here
	placeHold := "%s" // later becomes player, interactable, svg, weather, and highlight boxes

	out := ""
	out += swapToken(y, x, "Lg1", "g1", mat.Ground1Css)
	out += swapToken(y, x, "Lg2", "g2", mat.Ground2Css)
	out += swapToken(y, x, "Lf1", "f1", mat.Floor1Css)
	out += swapToken(y, x, "Lf2", "f2", mat.Floor2Css)
	out += placeHold
	out += placeHold
	out += placeHold
	out += swapToken(y, x, "Lc1", "c1", mat.Ceiling1Css)
	out += swapToken(y, x, "Lc2", "c2", mat.Ceiling2Css)
	out += placeHold
	out += placeHold

	return out
}

func swapToken(y, x int, prefix, zIndex, color string) string {
	return fmt.Sprintf(`[~ id="%s-%d-%d" class="box %s %s"]`, prefix, y, x, zIndex, color)
}

////////////////////////////////////////////////////////////
// Add Character

// Player

func (tile *Tile) addPlayerAndNotifyOthers(player *Player) {
	player.tileLock.Lock()
	defer player.tileLock.Unlock()
	tile.addLockedPlayerToTile(player)
	tile.stage.updateAllExcept(characterBox(tile), player)
}

func (tile *Tile) addLockedPlayerToTile(player *Player) {
	tile.CharacterMutex.Lock()
	defer tile.CharacterMutex.Unlock()

	// technically can race, e.g. with interactable reaction
	if tile.bottomText != "" {
		player.updateBottomText(tile.bottomText)
	}

	if tile.collectItemsForPlayer(player) {
		sendSoundToPlayer(player, "money")
		tile.stage.updateAll(svgFromTile(tile))
	}

	// player's tile lock should be held
	tile.characterMap[player.id] = player
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

	tile.itemMutex.Lock()
	defer tile.itemMutex.Unlock()
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

// npc

func addNPCAndNotifyOthers(npc *NonPlayer, tile *Tile) {
	npc.tileLock.Lock()
	defer npc.tileLock.Unlock()
	addLockedNPCToTile(npc, tile)
	tile.stage.updateAll(characterBox(tile))
}

func addLockedNPCToTile(npc *NonPlayer, tile *Tile) {
	tile.CharacterMutex.Lock()
	defer tile.CharacterMutex.Unlock()
	if collectItemNPC(tile, npc) {
		tile.stage.updateAll(svgFromTile(tile))
	}
	tile.characterMap[npc.id] = npc
	npc.tile = tile
	// No teleport for npc currently
}

func collectItemNPC(tile *Tile, npc *NonPlayer) bool {
	itemChange := false

	tile.itemMutex.Lock()
	defer tile.itemMutex.Unlock()
	if tile.powerUp != nil {
		tile.powerUp = nil
		// Activate power
		itemChange = true
	}
	if tile.money != 0 {
		npc.money.Add(int32(tile.money))
		tile.money = 0
		itemChange = true
	}
	if tile.boosts > 0 {
		npc.boosts.Add(int32(tile.boosts))
		tile.boosts = 0
		itemChange = true
	}
	return itemChange
}

func (tile *Tile) removePlayerAndNotifyOthers(player *Player) (success bool) {
	success = tryRemoveCharacterById(tile, player.id)
	if success {
		tile.stage.updateAllExcept(characterBox(tile), player)
	} else {
		// Possible under what circumstance :
		//   Handle death can race with logout to produce this (harmlessly?)
		logger.Warn().Msg("WARN : FAILED TO REMOVE PLAYER :(")
	}
	return success
}

func tryRemoveCharacterById(tile *Tile, id string) bool {
	tile.CharacterMutex.Lock()
	defer tile.CharacterMutex.Unlock()

	_, foundOnTile := tile.characterMap[id]
	if !foundOnTile {
		return false
	}

	delete(tile.characterMap, id)
	return true
}

func (tile *Tile) getACharacter() Character {
	tile.CharacterMutex.Lock()
	defer tile.CharacterMutex.Unlock()
	for _, player := range tile.characterMap {
		return player
	}
	return nil
}

/////////////////////////////////////////////////////////////////////
// Damage

func damageAndIndicate(tiles []*Tile, initiator Character, stage *Stage, damage int) {
	for _, tile := range tiles {
		tile.damageAll(damage, initiator)
		destroyFragileInteractable(tile, initiator)
		tile.eventsInFlight.Add(1)
		go tile.tryToNotifyAfter(100)
	}
	damageBoxes := sliceOfTileToWeatherBoxes(tiles, randomFieryColor())
	stage.updateAll(damageBoxes + soundTriggerByName("explosion"))
}

func (tile *Tile) damageAll(dmg int, initiator Character) {
	for _, character := range tile.copyOfCharacters() {
		character.takeDamageFrom(initiator, dmg)
	}
	tile.stage.updateAll(characterBox(tile))

}

func (tile *Tile) copyOfCharacters() []Character {
	players := make([]Character, 0)
	tile.CharacterMutex.Lock()
	for _, player := range tile.characterMap {
		players = append(players, player)
	}
	tile.CharacterMutex.Unlock()
	return players
}

func damagePlayerAndHandleDeath(player *Player, dmg int) bool {
	flashBackgroundColorIfTangible(player, "twilight")
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

func updateStreakIfTangible(player *Player) {
	ownLock := player.tangibilityLock.TryLock()
	if !ownLock || !player.tangible {
		return
	}
	defer player.tangibilityLock.Unlock()
	player.streakLock.Lock()
	defer player.streakLock.Unlock()
	html := spanStreak(player.killstreak)
	updateOne(html, player)
}

func safeFromDamage(tile *Tile) bool {
	if tile == nil {
		return true
	}
	stagename := tile.stage.name
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
		tile.stage.updateAll(weatherBox(tile, tile.stage.weather))
	}
}

//////////////////////////////////////////////////////////////////////
// Interactables

func replaceNilInteractable(tile *Tile, incoming *Interactable) bool {
	if incoming == nil {
		return true
	}
	if tile.material.Walkable { // Prevents lock contention from using Walkable()
		setLockedInteractableAndUpdate(tile, incoming)
		return true
	}

	return false
}

func setLockedInteractableAndUpdate(tile *Tile, incoming *Interactable) {
	tile.interactable = incoming
	tile.stage.updateAll(interactableBoxSpecific(tile.y, tile.x, tile.interactable))
}

func destroyInteractable(tile *Tile, _ *Player) {
	// *Player is a placeholder for initiator/destroyer in future
	tile.interactableMutex.Lock()
	defer tile.interactableMutex.Unlock()
	if tile.interactable != nil {
		tile.interactable = nil
		tile.stage.updateAll(interactableBoxSpecific(tile.y, tile.x, tile.interactable))
	}
}

func destroyFragileInteractable(tile *Tile, _ Character) {
	// *Player is a placeholder for initiator/destroyer in future
	tile.interactableMutex.Lock()
	defer tile.interactableMutex.Unlock()
	if tile.interactable != nil && tile.interactable.fragile {
		tile.interactable = nil
		tile.stage.updateAll(interactableBoxSpecific(tile.y, tile.x, tile.interactable))
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

func hasTeleport(tile *Tile) bool {
	if tile == nil || tile.teleport == nil {
		return false
	}
	return true
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

// use for airlock?
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

/////////////////////////////////////////////////////////////////
// Observers / Item state

func (tile *Tile) addPowerUpAndNotifyAll(shape [][2]int) {
	tile.addPowerUp(shape)
	tile.stage.updateAll(svgFromTile(tile))
}

func (tile *Tile) addPowerUp(shape [][2]int) {
	tile.itemMutex.Lock()
	defer tile.itemMutex.Unlock()
	tile.powerUp = &PowerUp{shape}
}

func (tile *Tile) addBoostsAndNotifyAll() {
	tile.addBoosts(10)
	tile.stage.updateAll(svgFromTile(tile))
}

func (tile *Tile) addBoosts(amount int) {
	tile.itemMutex.Lock()
	defer tile.itemMutex.Unlock()
	tile.boosts += amount
}

func (tile *Tile) addMoneyAndNotifyAll(amount int) {
	tile.addMoneyAndNotifyAllExcept(amount, nil)
}

func (tile *Tile) addMoneyAndNotifyAllExcept(amount int, player *Player) {
	tile.addMoney(amount)
	tile.stage.updateAllExcept(svgFromTile(tile), player)
}

func (tile *Tile) addMoney(amount int) {
	tile.itemMutex.Lock()
	defer tile.itemMutex.Unlock()
	tile.money += amount
}
