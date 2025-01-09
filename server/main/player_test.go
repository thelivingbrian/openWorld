package main

import (
	"sync"
	"testing"
)

func TestMoveNorthBoostWithValidNorthernNeighbor(t *testing.T) {
	loadFromJson()
	world := createGameWorld(testdb())
	testStage := createStageByName("hallway")
	updatesForPlayer := make(chan []byte)
	go drainChannel(updatesForPlayer)
	bufferClearChannel := make(chan struct{})
	go drainChannel(bufferClearChannel)

	player := &Player{
		id:    "tp",
		stage: testStage,
		//stageName:         testStage.name,
		x:                 4,
		y:                 1,
		actions:           createDefaultActions(),
		health:            100,
		updates:           updatesForPlayer,
		clearUpdateBuffer: bufferClearChannel,
		world:             world,
		tangible:          true,
	}
	player.placeOnStage(testStage)

	//fmt.Println(player.tile)

	// Act
	player.addBoosts(5)
	player.moveNorthBoost()

	// Assert
	// if player.stageName != "hallway2" {
	// 	t.Error("player stageName should be hallway2 but is: " + player.stageName)
	// }

	if player.stage.name != "hallway2" {
		t.Error("player.stage.name should be hallway2")
	}

	if player.y != 7 || player.x != 4 {
		t.Error("Player should be at y:7 x:4")

	}
}

func TestActivateHighlightWithMovement_NoConcurrentWrite(t *testing.T) {
	loadFromJson()
	world := createGameWorld(testdb())
	testStage := createStageByName("hallway")
	updatesForPlayer := make(chan []byte)
	go drainChannel(updatesForPlayer)
	bufferClearChannel := make(chan struct{})
	go drainChannel(bufferClearChannel)

	player := &Player{
		id:                "tp",
		stage:             testStage,
		x:                 4,
		y:                 1,
		actions:           createDefaultActions(),
		health:            100,
		updates:           updatesForPlayer,
		clearUpdateBuffer: bufferClearChannel,
		world:             world,
		tangible:          true,
	}
	player.placeOnStage(testStage)

	powerUp1 := &PowerUp{areaOfInfluence: grid9x9}
	powerUp2 := &PowerUp{areaOfInfluence: grid5x5}
	powerUp3 := &PowerUp{areaOfInfluence: jumpCross()}
	// seems sufficient for roughly 50% trigger rate
	reps := 50
	for i := 0; i < reps; i++ {
		player.actions.spaceStack.push(powerUp1)
		player.actions.spaceStack.push(powerUp2)
		player.actions.spaceStack.push(powerUp3)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < reps; i++ {
			wg.Add(4) // Add 4 waits for each iteration of the loop
			go func() { defer wg.Done(); player.moveNorth() }()
			go func() { defer wg.Done(); player.moveSouth() }()
			go func() { defer wg.Done(); player.moveEast() }()
			go func() { defer wg.Done(); player.moveWest() }()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < reps*3; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); player.activatePower() }()
		}
	}()

	wg.Wait()

	// No throw indicates success.
}

func (p *Player) placeOnStage(stage *Stage) {
	placePlayerOnStageAt(p, stage, p.y, p.x)
}
