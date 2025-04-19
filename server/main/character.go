package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type Character interface {
	getName() string
	getIconSync() string
	getTileSync() *Tile
	getTeamNameSync() string
	fetchStageSync(stagename string) *Stage
	transferBetween(source, dest *Tile)
	push(tile *Tile, incoming *Interactable, yOff, xOff int) bool
	takeDamageFrom(initiator Character, dmg int)
	incrementKillCount()
	incrementKillStreak()
	updateRecord()
}

// Move
func moveNorth(character Character) {
	move(character, -1, 0)
}

func moveSouth(character Character) {
	move(character, 1, 0)
}

func moveEast(character Character) {
	move(character, 0, 1)
}

func moveWest(character Character) {
	move(character, 0, -1)
}

func move(character Character, yOffset int, xOffset int) {
	sourceTile := character.getTileSync()
	character.push(sourceTile, nil, yOffset, xOffset)
	destTile := getRelativeTile(sourceTile, yOffset, xOffset, character)
	character.push(destTile, nil, yOffset, xOffset)
	if walkable(destTile) {
		character.transferBetween(sourceTile, destTile)
	}
}

// Push
func pushTeleport(character Character, tile *Tile, incoming *Interactable, yOff, xOff int) bool {
	if tile.teleport.rejectInteractable {
		return false
	}
	if canBeTeleported(incoming) {
		stage := character.fetchStageSync(tile.teleport.destStage)
		if !validCoordinate(tile.teleport.destY+yOff, tile.teleport.destX+xOff, stage) {
			return false
		}
		return character.push(stage.tiles[tile.teleport.destY+yOff][tile.teleport.destX+xOff], incoming, yOff, xOff)
	}
	return false
}

func canBeTeleported(interactable *Interactable) bool {
	if interactable == nil {
		return false
	}
	return !interactable.rejectTeleport
}

// Teleport
func applyTeleport(character Character, teleport *Teleport) {
	stage := character.fetchStageSync(teleport.destStage)
	if !validCoordinate(teleport.destY, teleport.destX, stage) {
		logger.Error().Msg(fmt.Sprint("Fatal: Invalid coords from teleport: ", teleport.destStage, teleport.destY, teleport.destX))
		return
	}
	// Is using getTileSync a risk with the menu teleport authorizer?
	character.transferBetween(character.getTileSync(), stage.tiles[teleport.destY][teleport.destX])
}

// Juke
func jukeRight(yOff, xOff int, character Character) {
	rel, rot := getRelativeAndRotate(yOff, xOff, character, true)
	swapIfEmpty(rel, rot)
}

func jukeLeft(yOff, xOff int, character Character) {
	rel, rot := getRelativeAndRotate(yOff, xOff, character, false)
	swapIfEmpty(rel, rot)
}

func tryJukeNorth(prev string, character Character) {
	switch prev {
	case "a": // coming from West - turning north
		jukeRight(0, -1, character)
	case "d": // coming from East - turning north
		jukeLeft(0, 1, character)
	}
}

func tryJukeSouth(prev string, character Character) {
	switch prev {
	case "d": // came from East  - turning South
		jukeRight(0, 1, character)
	case "a": // came from West - turning South
		jukeLeft(0, -1, character)
	}
}

func tryJukeWest(prev string, character Character) {
	switch prev {
	case "s": // came from South - turning west
		jukeRight(1, 0, character)
	case "w": // came from North - turning west
		jukeLeft(-1, 0, character)
	}
}

func tryJukeEast(prev string, character Character) {
	switch prev {
	case "w": // came from North - turning east
		jukeRight(-1, 0, character)
	case "s": // came from South - turning east
		jukeLeft(1, 0, character)
	}
}

/////////////////////////////////////////////////////////////////////
//  Player

func (player *Player) getName() string {
	return player.username
}

func (player *Player) getIconSync() string {
	player.viewLock.Lock()
	defer player.viewLock.Unlock()
	return player.icon
}

func (player *Player) getTileSync() *Tile {
	player.tileLock.Lock()
	defer player.tileLock.Unlock()
	return player.tile
}

func (player *Player) fetchStageSync(stagename string) *Stage {
	player.world.wStageMutex.Lock()
	defer player.world.wStageMutex.Unlock()
	stage, ok := player.world.worldStages[stagename]
	if ok && stage != nil {
		return stage
	}
	// stagename + team || stagename + rand

	player.pStageMutex.Lock()
	defer player.pStageMutex.Unlock()
	stage, ok = player.playerStages[stagename]
	if ok && stage != nil {
		return stage
	}

	area, success := areaFromName(stagename)
	if !success {
		return nil
	}

	stage = createStageFromArea(area) // can create empty stage
	if area.LoadStrategy == "" {
		player.world.worldStages[stagename] = stage
	}
	if area.LoadStrategy == "Personal" {
		player.playerStages[stagename] = stage
	}
	if area.LoadStrategy == "Individual" {
		// no-op : stage will load fresh each time
	}

	return stage
}

func (p *Player) transferBetween(source, dest *Tile) {
	if source.stage == dest.stage {
		if transferPlayerWithinStage(p, source, dest) {
			updateOthersAfterMovement(p, dest, source)
			updatePlayerAfterMovement(p, dest, source)
		}
	} else {
		if transferPlayerAcrossStages(p, source, dest) {
			spawnItemsFor(p, dest.stage)
			updateOthersAfterMovement(p, dest, source)
			updatePlayerAfterStageChange(p)
		}
	}
}
func transferPlayerWithinStage(p *Player, source, dest *Tile) bool {
	p.tileLock.Lock()
	defer p.tileLock.Unlock()

	if !tryRemoveCharacterById(source, p.id) {
		return false
	}

	dest.addLockedPlayerToTile(p)
	return true
}

func transferPlayerAcrossStages(p *Player, source, dest *Tile) bool {
	p.stageLock.Lock() // No need for this ?
	defer p.stageLock.Unlock()
	p.tileLock.Lock()
	defer p.tileLock.Unlock()

	if !tryRemoveCharacterById(source, p.id) {
		return false
	}

	p.stage.removeLockedPlayerById(p.id)
	p.stage = dest.stage

	dest.stage.addLockedPlayer(p)
	dest.addLockedPlayerToTile(p)
	return true
}

func (p *Player) push(tile *Tile, incoming *Interactable, yOff, xOff int) bool { // Returns if given interacable successfully pushed
	// Do not nil check incoming interactable here.
	// incoming = nil is valid and will continue a push chain
	// e.g. by taking this tile's interactable and pushing it forward
	if tile == nil {
		return false
	}

	if hasTeleport(tile) {
		return pushTeleport(p, tile, incoming, yOff, xOff)
	}

	ownLock := tile.interactableMutex.TryLock()
	if !ownLock {
		return false // Tile is already locked by another operation
	}
	defer tile.interactableMutex.Unlock()

	if tile.interactable == nil {
		return replaceNilInteractable(tile, incoming)
	}

	if tile.interactable.React(incoming, p, tile, yOff, xOff) {
		return true
	}

	if tile.interactable.pushable {
		nextTile := getRelativeTile(tile, yOff, xOff, p)
		if nextTile != nil {
			if p.push(nextTile, tile.interactable, yOff, xOff) {
				setLockedInteractableAndUpdate(tile, incoming)
				return true
			}
		}
	}
	return false
}

func (target *Player) takeDamageFrom(initiator Character, dmg int) {
	location := target.getTileSync()
	if safe(location, target, initiator) {
		return
	}

	fatal := damagePlayerAndHandleDeath(target, dmg)
	if fatal {
		initiator.incrementKillCount()
		initiator.incrementKillStreak()
		initiator.updateRecord()

		go target.world.db.saveKillEvent(location, initiator, target)
	}
}

func safe(location *Tile, partyA, partyB Character) bool {
	if safeFromDamage(location) {
		return true
	}
	if partyA.getTeamNameSync() == partyB.getTeamNameSync() {
		return true
	}
	return false
}

/////////////////////////////////////////////////////////////////////
//  Non-Player

type NonPlayer struct {
	id         string
	team       string
	teamLock   sync.Mutex
	icon       string
	iconLow    string
	world      *World
	tile       *Tile
	tileLock   sync.Mutex
	health     atomic.Int32
	money      atomic.Int32
	boosts     atomic.Int32
	killCount  atomic.Int32
	killStreak atomic.Int32
}

func (npc *NonPlayer) getName() string {
	return npc.id
}

func (npc *NonPlayer) getIconSync() string {
	health := npc.health.Load()
	if health <= 50 {
		return npc.iconLow
	}
	return npc.icon
}
func (npc *NonPlayer) getTileSync() *Tile {
	npc.tileLock.Lock()
	defer npc.tileLock.Unlock()
	return npc.tile
}

func (npc *NonPlayer) getTeamNameSync() string {
	npc.teamLock.Lock()
	defer npc.teamLock.Unlock()
	return npc.team
}

func (npc *NonPlayer) fetchStageSync(stagename string) *Stage {
	npc.world.wStageMutex.Lock()
	defer npc.world.wStageMutex.Unlock()
	stage, ok := npc.world.worldStages[stagename]
	if ok && stage != nil {
		return stage
	}

	area, success := areaFromName(stagename)
	if !success {
		return nil
	}

	stage = createStageFromArea(area) // can create empty stage
	if area.LoadStrategy == "" {
		npc.world.worldStages[stagename] = stage
	}
	if area.LoadStrategy == "Personal" {
		// npc does not have personal stages
		return nil
	}
	if area.LoadStrategy == "Individual" {
		// npc does not have individual stages
		return nil
	}

	return stage
}

func (npc *NonPlayer) transferBetween(source, dest *Tile) {
	if transferNPCBetweenTiles(npc, source, dest) {
		updateAllAfterMovement(dest, source)
	}
}

func transferNPCBetweenTiles(npc *NonPlayer, source, dest *Tile) bool {
	npc.tileLock.Lock()
	defer npc.tileLock.Unlock()

	if !tryRemoveCharacterById(source, npc.id) {
		return false
	}

	addLockedNPCToTile(npc, dest)
	return true
}

func (npc *NonPlayer) push(tile *Tile, incoming *Interactable, yOff, xOff int) bool {
	if tile == nil { // incoming = nil is valid
		return false
	}

	if hasTeleport(tile) {
		return pushTeleport(npc, tile, incoming, yOff, xOff)
	}

	ownLock := tile.interactableMutex.TryLock()
	if !ownLock {
		return false // Tile is already locked by another operation
	}
	defer tile.interactableMutex.Unlock()

	if tile.interactable == nil {
		return replaceNilInteractable(tile, incoming)
	}

	if tile.interactable.reactions != nil {
		// Reactions are undefined for npc
		return false
	}

	if tile.interactable.pushable {
		nextTile := getRelativeTile(tile, yOff, xOff, npc)
		if nextTile != nil {
			if npc.push(nextTile, tile.interactable, yOff, xOff) {
				setLockedInteractableAndUpdate(tile, incoming)
				return true
			}
		}
	}
	return false
}

func (npc *NonPlayer) takeDamageFrom(initiator Character, dmg int) {
	if safe(npc.getTileSync(), npc, initiator) {
		return
	}
	currentHealth := npc.health.Add(-int32(dmg))
	previousHealth := currentHealth + int32(dmg)
	if currentHealth <= 0 && previousHealth > 0 {
		initiator.incrementKillStreak()
		removeNpc(npc)
	}
}

func removeNpc(npc *NonPlayer) {
	npc.tileLock.Lock()
	defer npc.tileLock.Unlock()
	if !tryRemoveCharacterById(npc.tile, npc.id) {
		logger.Error().Msg("Error - FAILED TO REMOVE NPC")
		return
	}
	sound := soundTriggerByName("clink")
	npc.tile.stage.updateAll(sound)
}

func (npc *NonPlayer) incrementKillCount() {
	npc.killCount.Add(1)
}

func (npc *NonPlayer) incrementKillStreak() {
	npc.killStreak.Add(1)
}

func (npc *NonPlayer) updateRecord() {
	// Do Nothing
}

// Spawn NPC

func spawnNewNPCWithRandomMovement(ref *Player, interval int) (*NonPlayer, context.CancelFunc) {
	username := uuid.New().String()
	refTile := ref.getTileSync()
	npc := &NonPlayer{
		id:      username,
		world:   ref.world,
		icon:    "red-b thick r0",
		iconLow: "dark-red-b thick r0",
		health:  atomic.Int32{},
	}
	npc.health.Store(int32(100))

	addNPCAndNotifyOthers(npc, refTile)

	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Duration(interval) * time.Millisecond)
				randn := rand.Intn(5000)

				if randn%4 == 0 {
					moveNorth(npc)
					activatePower(npc)
				}
				if randn%4 == 1 {
					moveSouth(npc)
				}
				if randn%4 == 2 {
					moveEast(npc)
				}
				if randn%4 == 3 {
					moveWest(npc)
				}
			}
		}
	}(ctx)
	return npc, cancel
}

func activatePower(npc *NonPlayer) {
	shapes := [][][2]int{grid9x9, grid5x5}
	currentTile := npc.getTileSync()
	rand := rand.Intn(2)
	absCoordinatePairs := applyRelativeDistance(currentTile.y, currentTile.x, shapes[rand])
	tiles := make([]*Tile, 0)
	for _, pair := range absCoordinatePairs {
		if validCoordinate(pair[0], pair[1], currentTile.stage) {
			tile := currentTile.stage.tiles[pair[0]][pair[1]]
			tiles = append(tiles, tile)
		}
	}
	damageAndIndicate(tiles, npc, currentTile.stage, 50)
	// damage tiles
}

//////////////////////////////////////////////////////////////////////
// Rotation

func rotate(character Character, orientClockwise bool) {
	sourceTile := character.getTileSync()
	n := getRelativeTile(sourceTile, -1, 0, character)
	s := getRelativeTile(sourceTile, 1, 0, character)
	e := getRelativeTile(sourceTile, 0, 1, character)
	w := getRelativeTile(sourceTile, 0, -1, character)
	var path []*Tile
	if orientClockwise {
		path = []*Tile{n, e, s, w}
	} else {
		path = []*Tile{n, w, s, e}
	}

	cycleInteractableList(path)
}

func cycleInteractableList(path []*Tile) {
	for i := 0; i < len(path); i++ {
		if path[i] == nil {
			continue
		}
		ownLock := path[i].interactableMutex.TryLock()
		if !ownLock {
			continue
		}

		ok, last, depth := cycleForward(path, i, 0)
		if ok {
			setLockedInteractableAndUpdate(path[i], last)

		}
		path[i].interactableMutex.Unlock() // cannot defer
		i += depth
	}
}

func cycleForward(path []*Tile, index, depth int) (bool, *Interactable, int) {
	final := depth == len(path)-1
	if final {
		return true, path[index].interactable, depth + 1
	}

	if !(path[index].interactable == nil) && !path[index].interactable.pushable {
		return false, nil, depth
	}

	next := mod(index+1, len(path))
	if path[next] == nil {
		return false, nil, depth + 1
	}

	ownTarget := path[next].interactableMutex.TryLock()
	if !ownTarget {
		return false, nil, depth + 1
	}
	defer path[next].interactableMutex.Unlock()

	if path[next].interactable == nil {
		ok := replaceNilInteractable(path[next], path[index].interactable)
		return ok, nil, depth + 1
	}
	if !path[next].interactable.pushable {
		return false, nil, depth + 1
	}

	ok, out, newDepth := cycleForward(path, index+1, depth+1)
	if ok {
		setLockedInteractableAndUpdate(path[next], path[index].interactable)
	}
	return ok, out, newDepth
}

/// New

func swapIfEmpty(source, target *Tile) {
	ownSource := source.interactableMutex.TryLock()
	if !ownSource {
		return
	}
	defer source.interactableMutex.Unlock()
	if source.interactable == nil || !source.interactable.pushable {
		return
	}
	ownTarget := target.interactableMutex.TryLock()
	if !ownTarget {
		return
	}
	defer target.interactableMutex.Unlock()
	if target.interactable != nil {
		return
	}
	if replaceNilInteractable(target, source.interactable) {
		setLockedInteractableAndUpdate(source, nil)
	}
}
