package main

import (
	"sync"
	"testing"
)

func TestActivateHighlightWithMovement_NoConcurrentWrite(t *testing.T) {
	loadFromJson()
	//world := createGameWorld(testdb(), nil)
	world, shutDown := createWorldForTesting()
	defer shutDown()

	testStage := createStageByName("hallway")
	// updatesForPlayer := make(chan []byte)
	// go drainChannel(updatesForPlayer) // needs to be stopped
	// bufferClearChannel := make(chan struct{})
	// go drainChannel(bufferClearChannel)

	player := createTestingPlayer(world, "")
	player.placeOnStage(testStage, 1, 4)

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
