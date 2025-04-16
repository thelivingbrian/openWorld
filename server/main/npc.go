package main

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type Character interface {
	receiveDamageFrom(initiator *Player, dmg int) bool
	getIconSync() string
	getTileSync() *Tile
	fetchStageSync(stagename string) *Stage
	transferPlayer(source, dest *Tile)
	push(tile *Tile, incoming *Interactable, yOff, xOff int) bool
}

type NonPlayer struct {
	id       string
	icon     string
	iconLock sync.Mutex
	world    *World
	tile     *Tile
	tileLock sync.Mutex
	health   atomic.Int32
	money    atomic.Int32
	boosts   atomic.Int32
}

func (npc *NonPlayer) getIconSync() string {
	health := npc.health.Load()
	if health <= 50 {
		return "dim-" + npc.icon
	}
	return npc.icon
}
func (npc *NonPlayer) getTileSync() *Tile {
	npc.tileLock.Lock()
	defer npc.tileLock.Unlock()
	return npc.tile
}

func (npc *NonPlayer) receiveDamageFrom(initiator *Player, dmg int) bool {
	if npc.health.Add(int32(-dmg)) < 0 {
		npc.tileLock.Lock()
		defer npc.tileLock.Unlock()
		if tryRemoveCharacter(npc.tile, npc.id) {
			npc.tile.stage.updateAll(CharacterBox(npc.tile))
		} else {
			logger.Warn().Msg("FAILED TO REMOVE AN NPC")
		}
		return true
	}
	return false
}

func (npc *NonPlayer) fetchStageSync(stagename string) *Stage {
	npc.world.wStageMutex.Lock()
	defer npc.world.wStageMutex.Unlock()
	stage, ok := npc.world.worldStages[stagename]
	if ok && stage != nil {
		return stage
	}
	// stagename + team || stagename + rand

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

func (npc *NonPlayer) transferPlayer(source, dest *Tile) {
	if transferNPCBetweenTiles(npc, source, dest) {
		updateAllAfterMovement(dest, source)
	}
}

func transferNPCBetweenTiles(npc *NonPlayer, source, dest *Tile) bool {
	npc.tileLock.Lock()
	defer npc.tileLock.Unlock()

	if !tryRemoveCharacter(source, npc.id) {
		return false
	}

	addLockedNPCToTile(npc, dest)
	//dest.addLockedCharacterToTile(npc)
	return true
}

func (p *NonPlayer) push(tile *Tile, incoming *Interactable, yOff, xOff int) bool { // Returns if given interacable successfully pushed
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

	if tile.interactable.reactions != nil {
		// Reactions are undefined for npc
		return false
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

func spawnNewNPCWithRandomMovement(ref *Player, interval int) (*NonPlayer, context.CancelFunc) {
	username := uuid.New().String()
	refTile := ref.getTileSync()
	npc := &NonPlayer{
		id:     username,
		world:  ref.world,
		icon:   "red r0 black-b thick",
		health: atomic.Int32{},
	}
	npc.health.Store(int32(100))
	addNPCToTile(npc, refTile)

	ctx, cancel := context.WithCancel(context.Background())
	// go func(ctx context.Context) {
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			return
	// 		default:
	// 			<-newPlayer.updates
	// 		}
	// 	}
	// }(ctx)
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
