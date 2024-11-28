package main

import (
	"fmt"
	"math/rand"
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
	boosts            int
	htmlTemplate      string
	bottomText        string
}

type Teleport struct {
	destStage    string
	destY        int
	destX        int
	sourceStage  string
	confirmation bool
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

func makeTileTemplate(mat Material, y, x int) string {
	tileCoord := fmt.Sprintf("%d-%d", y, x)
	cId := "c" + tileCoord // This is used to identify the entire square
	tId := "t" + tileCoord // This is used to identify the top highlight box
	placeHold := "%s"      // later becomes player, interactable, svg, and weather boxes

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
	return fmt.Sprintf(template, cId, mat.CssColor, floor1css, floor2css, placeHold, placeHold, placeHold, ceil1css, ceil2css, placeHold, tId)
}

// newTile w/ teleport?

func (tile *Tile) addPlayerAndNotifyOthers(player *Player) {
	tile.addPlayer(player)
	tile.stage.updateAllExcept(playerBox(tile), player)
}

func (tile *Tile) addPlayer(player *Player) {
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
		player.stage.updateAll(svgFromTile(tile))
	}
	if tile.teleport == nil {
		// Worth it? no - but maybe with dupe logic?
		//if player.getHealthSync() == 0 {
		//	return
		//}
		tile.playerMutex.Lock()
		tile.playerMap[player.id] = player
		tile.playerMutex.Unlock()
		player.y = tile.y
		player.x = tile.x
		// Is reverse order better?
		player.tileLock.Lock()
		defer player.tileLock.Unlock()
		player.tile = tile
	} else {
		if tile.teleport.confirmation {
			player.menues["teleport"] = continueTeleporting(tile.teleport)
			turnMenuOn(player, "teleport")
		} else {
			player.applyTeleport(tile.teleport)
		}
	}
}

func (tile *Tile) removePlayerAndNotifyOthers(player *Player) {
	tile.removePlayer(player.id)
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

func (tile *Tile) tryToNotifyAfter(delay int) {
	time.Sleep(time.Millisecond * time.Duration(delay))
	tile.eventsInFlight.Add(-1)
	if tile.eventsInFlight.Load() == 0 {
		// blue trsp20 for gloom
		tile.stage.updateAll(weatherBox(tile, ""))
	}
}

func (tile *Tile) damageAll(dmg int, initiator *Player) {
	survivors := false
	// player map needs mutex ?
	for _, player := range tile.playerMap {
		if player.getTeamNameSync() == initiator.getTeamNameSync() {
			continue
		}
		survived := player.addToHealth(-dmg)
		survivors = survivors || survived
		if !survived {
			initiator.incrementKillCount()
			initiator.incrementKillStreak()
			initiator.updateRecord()
			go player.world.db.saveKillEvent(tile, initiator, player) // Maybe should just pass in required fields?
		}
	}
	if survivors {
		tile.stage.updateAll(playerBox(tile))
	}
}

func destroyInteractable(tile *Tile, _ *Player) {
	// *Player is a placeholder for initiator/destroyer in future
	tile.interactableMutex.Lock()
	defer tile.interactableMutex.Unlock()
	if tile.interactable != nil && tile.interactable.fragile {
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

	if tile.interactable == nil {
		return tile.material.Walkable
	} else {
		// stops obstruction by pushables
		return tile.interactable.pushable
	}
}

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
	tile.money += amount
	tile.stage.updateAll(svgFromTile(tile))
}

///

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

////////////////////////////////////////////
// Interactables and reactions

type Interactable struct {
	name      string
	pushable  bool
	cssClass  string
	fragile   bool
	reactions []InteractableReaction
}

type InteractableReaction struct {
	ReactsWith func(*Interactable) bool
	Reaction   func(incoming *Interactable, initiatior *Player, location *Tile)
}

var interactableReactions = map[string][]InteractableReaction{
	"black-hole": []InteractableReaction{InteractableReaction{ReactsWith: everything, Reaction: eat}},
	"fuchsia-goal": []InteractableReaction{
		InteractableReaction{ReactsWith: matchesName("fuchsia-ball"), Reaction: scoreGoalForTeam("sky-blue")},
		InteractableReaction{ReactsWith: matchesName("sky-blue-ball"), Reaction: spawnMoney([]int{10, 20, 50})},
	},
	"sky-blue-goal": []InteractableReaction{
		InteractableReaction{ReactsWith: matchesName("sky-blue-ball"), Reaction: scoreGoalForTeam("fuchsia")},
		InteractableReaction{ReactsWith: matchesName("fuchsia-ball"), Reaction: spawnMoney([]int{10, 20, 50})},
	},
}

func (source *Interactable) React(incoming *Interactable, initiatior *Player, location *Tile) bool {
	if source.reactions == nil {
		return false
	}
	for i := range source.reactions {
		if source.reactions[i].ReactsWith != nil && source.reactions[i].ReactsWith(incoming) {
			source.reactions[i].Reaction(incoming, initiatior, location)
			return true
		}
	}
	return false
}

// Gates
func everything(*Interactable) bool {
	return true
}

func matchesCssClass(cssClass string) func(*Interactable) bool {
	return func(i *Interactable) bool {
		return i.cssClass == cssClass
	}
}

func matchesName(name string) func(*Interactable) bool {
	return func(i *Interactable) bool {
		if i == nil {
			return false
		}
		return i.name == name
	}
}

// Actions
func eat(*Interactable, *Player, *Tile) {
	// incoming interactable is discarded
}

func scoreGoalForTeam(team string) func(*Interactable, *Player, *Tile) {
	return func(i *Interactable, p *Player, t *Tile) {
		p.world.leaderBoard.scoreboard.Increment(team)
		if team == p.getTeamNameSync() {
			// Otherwise you have scored a goal for a different team
			p.incrementGoalsScored()
			p.updateRecord()
		}
		fmt.Println(p.world.leaderBoard.scoreboard.GetScore(team))
	}
}

func spawnMoney(amounts []int) func(*Interactable, *Player, *Tile) {
	return func(i *Interactable, p *Player, t *Tile) {
		tiles := walkableTiles(t.stage.tiles)
		count := len(tiles)
		if count == 0 {
			return
		}
		for i := range amounts {
			randn := rand.Intn(count)
			tiles[randn].addMoneyAndNotifyAll(amounts[i])
		}
	}
}
