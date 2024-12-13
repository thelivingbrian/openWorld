package main

import (
	"fmt"
	"testing"
	"time"
)

func TestDamageABunchOfPlayers(t *testing.T) {
	// Settings
	movementDelay := 200
	activationDelay := 1000
	playerCount := 500

	// Arrange
	loadFromJson()
	world := createGameWorld(testdb())

	testStage := world.getNamedStageOrDefault("test-walls-interactable")
	testStage.spawn = []SpawnAction{SpawnAction{always, addBoostsAt(11, 13)}}
	if len(world.worldStages) != 1 {
		t.Error("Should have two stages")
	}
	clinic := world.getNamedStageOrDefault("clinic")
	clinic.tiles[12][12].teleport = nil
	if len(world.worldStages) != 2 {
		t.Error("Should have one stage")
	}

	//
	//   . | p p b
	//   p | p . p
	//   . | . . p <- 13,13
	//
	testStage.tiles[13][13].addPowerUpAndNotifyAll(grid9x9)
	testStage.tiles[12][13].addPowerUpAndNotifyAll(grid9x9)
	// 11,13 should have boosts
	testStage.tiles[11][12].addPowerUpAndNotifyAll(grid9x9)
	testStage.tiles[11][11].addPowerUpAndNotifyAll(grid9x9)
	testStage.tiles[12][11].addPowerUpAndNotifyAll(grid9x9)
	testStage.tiles[12][9].addPowerUpAndNotifyAll(grid9x9)

	// Join initial player
	record := PlayerRecord{Username: "test1", Y: 13, X: 13, StageName: testStage.name}
	req := createLoginRequest(record)
	world.addIncoming(req)
	p := world.join(req)
	go drainChannel(p.updates)
	p.placeOnStage(testStage)

	// Get in position
	p.moveEastBoost() // should do nothing
	if p.tile != testStage.tiles[13][13] {
		t.Error("Player should not have moved")
	}
	// collect all of the power ups
	p.moveNorth()
	p.moveNorth()
	p.moveWest()
	p.moveWest()
	p.moveSouth()

	// Spawn all of the clone players
	clones := make([]*Player, playerCount)
	for i := range clones {
		clone, cancel := spawnNewPlayerWithRandomMovement(p, movementDelay)
		clones[i] = clone
		defer cancel()
	}
	if len(p.world.worldPlayers) != 501 {
		t.Error(fmt.Sprintf("Player count should be 501 but is: %d", len(p.world.worldPlayers)))
	}

	// Escape the box
	p.moveWestBoost()
	if p.tile != testStage.tiles[12][9] {
		t.Error("Player should not have moved")
	}

	// Activate every collected power
	fmt.Println("Starting killstreak: ", p.getKillStreakSync(), " / 500")
	for count := 0; count < 7; count++ {
		p.activatePower()
		time.Sleep(time.Duration(activationDelay) * time.Millisecond)
		fmt.Println("current ks: ", p.getKillStreakSync(), " / 500")
	}

	// check each clone is in clinic
	for i := range clones {
		if clones[i].stage.name != "clinic" {
			t.Error(fmt.Sprintf("Clone#%d should be on clinic but is on: %s", i, clones[i].stage.name))
		}
	}

	// check original box is empty of clones
	for dy := 0; dy < 3; dy++ {
		for dx := 0; dx < 3; dx++ {
			yPos, xPos, playerCount := 11+dy, 11+dx, len(testStage.tiles[11+dy][11+dx].playerMap)
			if playerCount != 0 {
				t.Error(fmt.Sprintf("Tile(y:%d x:%d) should have 0 players but has: %d", yPos, xPos, playerCount))
			}
		}
	}

	// check player
	if p.stage != testStage {
		t.Error("Player should be on the test stage")
	}
	if p.getKillStreakSync() < 500 {
		t.Error("Killstreak should be at least 500")
	}

	// For laptop:
	time.Sleep(2000 * time.Millisecond) // slow computer can have trailing kills flowing in

	// respawn using menu
	menu := p.menues["respawn"]
	menu.attemptClick(p, PlayerSocketEvent{Arg0: "0"})
	if p.stage.name != "clinic" {
		t.Error("Player should be in the clinic")
	}
	if p.getKillStreakSync() != 0 {
		t.Error("Killstreak should have reset to 0")
	}
}
