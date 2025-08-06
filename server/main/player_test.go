package main

import (
	"testing"
)

func TestMoveNorthBoostWithValidNorthernNeighbor(t *testing.T) {
	loadFromJson()
	world, shutDown := createWorldForTesting()
	defer shutDown()
	player := createTestingPlayer(world, "")
	defer close(player.updates)

	testStage := createStageByName("hallway")
	player.placeOnStage(testStage, 1, 4)

	// Act
	player.addBoostsAndUpdate(5)
	player.moveNorthBoost()

	if player.getTileSync().stage.name != "hallway2" {
		t.Error("player.stage.name should be hallway2")
	}

	if player.getTileSync().y != 7 || player.getTileSync().x != 4 {
		t.Error("Player should be at y:7 x:4")

	}
}

func (p *Player) placeOnStage(stage *Stage, y, x int) {
	placePlayerOnStageAt(p, stage, y, x)
}

// Utilities
func createTestingPlayer(world *World, user string) *Player {
	updatesForPlayer := make(chan []byte)
	go drainChannel(updatesForPlayer)

	id := "tp" + user
	tp := &Player{
		id:           id,
		username:     user,
		actions:      createDefaultActions(),
		updates:      updatesForPlayer,
		tangible:     true,
		playerStages: map[string]*Stage{},
		world:        world,
		camera:       newCamera(updatesForPlayer),
	}
	tp.health.Store(100)
	world.worldPlayers[id] = tp
	return tp
}

func (player *Player) getKillStreakSync() int {
	return int(player.killstreak.Load())
}
