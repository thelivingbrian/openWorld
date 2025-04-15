package main

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

type Character interface {
	receiveDamageFrom(initiator *Player, dmg int) bool
	getIconSync() string
}

type NonPlayer struct {
	id       string
	icon     string
	tile     *Tile
	tileLock sync.Mutex
	health   atomic.Int32
	money    atomic.Int32
	boosts   atomic.Int32
}

func (npc *NonPlayer) getIconSync() string {
	return npc.icon
}

func (npc *NonPlayer) receiveDamageFrom(initiator *Player, dmg int) bool {
	if npc.health.Add(int32(-dmg)) < 0 {
		return true
	}
	return false
}

func spawnNewNPCWithRandomMovement(ref *Player, interval int) (*NonPlayer, context.CancelFunc) {
	username := uuid.New().String()
	refTile := ref.getTileSync()
	npc := &NonPlayer{
		id:     username,
		icon:   "red r0 black-b thick",
		health: atomic.Int32{},
	}
	npc.health.Store(int32(100))
	addNPCToTile(refTile, npc)

	_, cancel := context.WithCancel(context.Background())
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
	// go func(ctx context.Context) {
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			return
	// 		default:
	// 			time.Sleep(time.Duration(interval) * time.Millisecond)
	// 			randn := rand.Intn(5000)

	// 			if randn%4 == 0 {
	// 				newPlayer.moveNorth()
	// 			}
	// 			if randn%4 == 1 {
	// 				newPlayer.moveSouth()
	// 			}
	// 			if randn%4 == 2 {
	// 				newPlayer.moveEast()
	// 			}
	// 			if randn%4 == 3 {
	// 				newPlayer.moveWest()
	// 			}
	// 		}
	// 	}
	// }(ctx)
	return npc, cancel
}
