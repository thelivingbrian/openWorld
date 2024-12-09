package main

import (
	"fmt"
	"testing"
	"time"
)

func TestDamageABunchOfPlayers(t *testing.T) {
	loadFromJson()
	world := createGameWorld(testdb())

	clinic := world.getNamedStageOrDefault("clinic")
	clinic.tiles[12][12].teleport = nil
	if len(world.worldStages) != 1 {
		t.Error("Should have one stage")
	}

	testStage := world.getNamedStageOrDefault("test-walls-interactable")
	testStage.spawn = []SpawnAction{SpawnAction{always, addBoostsAt(11, 13)}}

	// . | . p b
	// p | p p p
	// . | p p p
	testStage.tiles[13][13].addPowerUpAndNotifyAll(grid9x9)
	testStage.tiles[12][13].addPowerUpAndNotifyAll(grid9x9)
	// 11,13 should have boosts
	testStage.tiles[11][12].addPowerUpAndNotifyAll(grid9x9)
	testStage.tiles[12][12].addPowerUpAndNotifyAll(grid9x9)
	testStage.tiles[13][12].addPowerUpAndNotifyAll(grid9x9)
	testStage.tiles[13][11].addPowerUpAndNotifyAll(grid9x9)
	testStage.tiles[12][11].addPowerUpAndNotifyAll(grid9x9)
	testStage.tiles[12][9].addPowerUpAndNotifyAll(grid9x9)
	if len(world.worldStages) != 2 {
		t.Error("Should have two stages")
	}

	p := world.join(&PlayerRecord{Username: "test1", Y: 13, X: 13, StageName: testStage.name})
	go drainChannel(p.updates)
	p.placeOnStage(testStage)

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

	var clones [500]*Player
	for i := range clones {
		clones[i] = spawnNewPlayerWithRandomMovement(p)
	}

	fmt.Println("current players", len(p.world.worldPlayers))

	if len(p.world.worldPlayers) != 501 {
		t.Error(fmt.Sprintf("Player count should be 501 but is: %d", len(p.world.worldPlayers)))
	}

	p.moveWestBoost()
	if p.tile != testStage.tiles[12][9] {
		t.Error("Player should not have moved")
	}

	for count := 0; count < 9; count++ {
		p.activatePower()
		time.Sleep(1000 * time.Millisecond)
		fmt.Println("current ks", p.getKillStreakSync())
	}

	for i := range clones {
		if clones[i].stage.name != "clinic" {
			t.Error(fmt.Sprintf("Clone#%d should be on clinic but is on: %s", i, clones[i].stage.name))
		}
	}

	fmt.Println("0" + p.stage.name)
	menu := p.menues["respawn"]
	menu.attemptClick(p, PlayerSocketEvent{Arg0: "0"})
	fmt.Println("1" + p.stage.name)
}
