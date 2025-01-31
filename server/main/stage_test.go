package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestEnsureInteractableWillPush(t *testing.T) {
	loadFromJson()
	testStage := createStageByName("test-walls-interactable")
	updatesForPlayer := make(chan []byte)
	go drainChannel(updatesForPlayer)
	bufferClearChannel := make(chan struct{})
	go drainChannel(bufferClearChannel)

	player := &Player{
		id:                "tp",
		stage:             testStage,
		actions:           createDefaultActions(),
		health:            100,
		updates:           updatesForPlayer,
		clearUpdateBuffer: bufferClearChannel,
		tangible:          true,
	}
	player.placeOnStage(testStage, 14, 1)

	if len(player.stage.tiles[14][1].playerMap) == 0 {
		t.Error("Player did not spawn at correct location")
	}

	if player.stage.tiles[14][2].interactable == nil || !player.stage.tiles[14][2].interactable.pushable {
		t.Error("test-walls-interactable should have pushable at 14,2")
	}

	player.moveEast()
	player.moveEast()
	player.moveWest()
	player.moveNorth()

	if player.stage.tiles[14][4].interactable == nil {
		t.Error("Interactable did not push")
	}

	if player.stage.tiles[14][2].interactable != nil {
		t.Error("Interactable still at starting location despite being pushed")
	}

	if len(player.stage.tiles[13][2].playerMap) == 0 {
		t.Error("Player has not moved correctly:")
		fmt.Printf("Y%dX%d", player.getTileSync().y, player.getTileSync().x)
	}
}

func TestSurroundedPushableSquare(t *testing.T) {
	loadFromJson()
	testStage := createStageByName("test-walls-interactable")
	updatesForPlayer := make(chan []byte)
	go drainChannel(updatesForPlayer)
	bufferClearChannel := make(chan struct{})
	go drainChannel(bufferClearChannel)

	if testStage.tiles[14][2].interactable == nil ||
		testStage.tiles[3][7].interactable == nil ||
		testStage.tiles[3][8].interactable == nil ||
		testStage.tiles[4][7].interactable == nil ||
		testStage.tiles[4][8].interactable == nil {
		t.Error("Initial state of test-walls-interactable does not have correct 5 interactables")
	}

	// Place players around the 2x2 square of pushable tiles (3,7) (3,8) (4,7) (4,8)
	players := []*Player{
		{id: "p0", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
		{id: "p1", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
		{id: "p2", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
		{id: "p3", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
		{id: "p4", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
		{id: "p5", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
		{id: "p6", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
		{id: "p7", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
	}

	positions := []struct {
		y, x int
	}{
		{2, 7},
		{2, 8},
		{3, 6},
		{4, 6},
		{3, 9},
		{4, 9},
		{5, 7},
		{5, 8},
	}

	for i, player := range players {
		player.placeOnStage(testStage, positions[i].y, positions[i].x)
	}

	// Act
	players[0].moveSouth() // p1 pushes from (2,7) to (3,7)
	players[1].moveSouth() // p2 pushes from (2,8) to (3,3)
	players[2].moveEast()  // p3 pushes from (3,6) to (3,7)
	players[3].moveEast()
	players[4].moveWest() // p5 pushes from (3,9) . . .
	players[5].moveWest()
	players[6].moveNorth()
	players[7].moveNorth()

	// Assert - Exact positions are known
	if testStage.tiles[14][2].interactable == nil ||
		testStage.tiles[2][7].interactable == nil ||
		testStage.tiles[3][7].interactable == nil ||
		testStage.tiles[3][8].interactable == nil ||
		testStage.tiles[4][6].interactable == nil {
		t.Error("Final state of test-walls-interactable does not have correct 5 interactables")
	}
}

func TestSurroundedPushableSquareMultipleThreads(t *testing.T) {
	loadFromJson()
	testStage := createStageByName("test-walls-interactable")
	updatesForPlayer := make(chan []byte)
	go drainChannel(updatesForPlayer)
	bufferClearChannel := make(chan struct{})
	go drainChannel(bufferClearChannel)

	if testStage.tiles[14][2].interactable == nil ||
		testStage.tiles[3][7].interactable == nil ||
		testStage.tiles[3][8].interactable == nil ||
		testStage.tiles[4][7].interactable == nil ||
		testStage.tiles[4][8].interactable == nil {
		t.Error("Initial state of test-walls-interactable does not have correct 5 interactables")
	}

	// Place players around the 2x2 square of pushable tiles (3,7) (3,8) (4,7) (4,8)
	players := []*Player{
		{id: "p0", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
		{id: "p1", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
		{id: "p2", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
		{id: "p3", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
		{id: "p4", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
		{id: "p5", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
		{id: "p6", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
		{id: "p7", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
	}

	positions := []struct {
		y, x int
	}{
		{2, 7},
		{2, 8},
		{3, 6},
		{4, 6},
		{3, 9},
		{4, 9},
		{5, 7},
		{5, 8},
	}

	for i, player := range players {
		player.placeOnStage(testStage, positions[i].y, positions[i].x)
	}

	var wg sync.WaitGroup
	wg.Add(len(players))

	// Initial push from players
	go func(wg *sync.WaitGroup) { defer wg.Done(); players[0].moveSouth() }(&wg)
	go func(wg *sync.WaitGroup) { defer wg.Done(); players[1].moveSouth() }(&wg)
	go func(wg *sync.WaitGroup) { defer wg.Done(); players[2].moveEast() }(&wg)
	go func(wg *sync.WaitGroup) { defer wg.Done(); players[3].moveEast() }(&wg)
	go func(wg *sync.WaitGroup) { defer wg.Done(); players[4].moveWest() }(&wg)
	go func(wg *sync.WaitGroup) { defer wg.Done(); players[5].moveWest() }(&wg)
	go func(wg *sync.WaitGroup) { defer wg.Done(); players[6].moveNorth() }(&wg)
	go func(wg *sync.WaitGroup) { defer wg.Done(); players[7].moveNorth() }(&wg)

	wg.Wait()

	// Count all interactables on the stage
	totalInteractables := 0
	for y := range testStage.tiles {
		for x := range testStage.tiles[y] {
			if testStage.tiles[y][x].interactable != nil {
				//fmt.Printf("found: y:%d x:%d\n", y, x)
				totalInteractables++
			}
		}
	}

	// Assert
	if totalInteractables != 5 {
		t.Errorf("Expected 5 interactables on the stage, found %d", totalInteractables)
	}
}

func TestEnsureNoInteractableDuplication(t *testing.T) {
	loadFromJson()
	testStage := createStageByName("test-walls-interactable")
	updatesForPlayer := make(chan []byte)
	go drainChannel(updatesForPlayer)
	bufferClearChannel := make(chan struct{})
	go drainChannel(bufferClearChannel)

	if testStage.tiles[14][2].interactable == nil ||
		testStage.tiles[3][7].interactable == nil ||
		testStage.tiles[3][8].interactable == nil ||
		testStage.tiles[4][7].interactable == nil ||
		testStage.tiles[4][8].interactable == nil {
		t.Error("Initial state of test-walls-interactable does not have correct 5 interactables")
	}

	testStage.tiles[3][8].interactable = nil
	testStage.tiles[4][8].interactable = nil
	testStage.tiles[14][2].interactable = nil
	testStage.tiles[3][7].interactable = &Interactable{pushable: true}
	testStage.tiles[4][7].interactable = &Interactable{pushable: true}
	testStage.tiles[5][7].interactable = &Interactable{pushable: true}
	testStage.tiles[6][7].interactable = &Interactable{pushable: true}
	testStage.tiles[7][7].interactable = &Interactable{pushable: true}
	testStage.tiles[8][7].interactable = &Interactable{pushable: true}
	testStage.tiles[9][7].interactable = &Interactable{pushable: true}
	testStage.tiles[10][7].interactable = &Interactable{pushable: true}
	testStage.tiles[11][7].interactable = &Interactable{pushable: true}
	testStage.tiles[12][7].interactable = &Interactable{pushable: true}
	testStage.tiles[13][7].interactable = &Interactable{pushable: true}

	// Place 2 players at ends of long interactable line
	players := []*Player{
		{id: "p0", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
		{id: "p1", stage: testStage, updates: updatesForPlayer, clearUpdateBuffer: bufferClearChannel, actions: createDefaultActions(), health: 100, tangible: true},
	}

	positions := []struct {
		y, x int
	}{
		{2, 7},
		{14, 7},
	}

	for i, player := range players {
		player.placeOnStage(testStage, positions[i].y, positions[i].x)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func(wg *sync.WaitGroup) { defer wg.Done(); players[0].moveSouth() }(&wg)
	go func(wg *sync.WaitGroup) { defer wg.Done(); players[1].moveNorth() }(&wg)

	wg.Wait()

	// Count all interactables on the stage
	totalInteractables := 0
	for y := range testStage.tiles {
		for x := range testStage.tiles[y] {
			if testStage.tiles[y][x].interactable != nil {
				//fmt.Printf("found: y:%d x:%d\n", y, x)
				totalInteractables++
			}
		}
	}

	// Assert
	if totalInteractables != 11 {
		t.Errorf("Expected 11 interactables on the stage, found %d", totalInteractables)
	}
}
