package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestEnsureInteractableWillPush(t *testing.T) {
	loadFromJson()
	testStage := createStageByName("test-walls-interactable")
	updatesForPlayer := make(chan Update)
	go drainChannel(updatesForPlayer)

	player := &Player{
		id:        "tp",
		stage:     testStage,
		stageName: testStage.name,
		x:         1,
		y:         14,
		actions:   createDefaultActions(),
		health:    100,
		updates:   updatesForPlayer,
	}
	player.placeOnStage()

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
		fmt.Printf("Y%dX%d", player.y, player.x)
	}
}

func TestSurroundedPushableSquare(t *testing.T) {
	loadFromJson()
	testStage := createStageByName("test-walls-interactable")
	updatesForPlayer := make(chan Update)
	go drainChannel(updatesForPlayer)

	if testStage.tiles[14][2].interactable == nil ||
		testStage.tiles[3][7].interactable == nil ||
		testStage.tiles[3][8].interactable == nil ||
		testStage.tiles[4][7].interactable == nil ||
		testStage.tiles[4][8].interactable == nil {
		t.Error("Initial state of test-walls-interactable does not have correct 5 interactables")
	}

	// Place players around the 2x2 square of pushable tiles (3,7) (3,8) (4,7) (4,8)
	players := []*Player{
		{id: "p0", stage: testStage, stageName: testStage.name, y: 2, x: 7, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p1", stage: testStage, stageName: testStage.name, y: 2, x: 8, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p2", stage: testStage, stageName: testStage.name, y: 3, x: 6, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p3", stage: testStage, stageName: testStage.name, y: 4, x: 6, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p4", stage: testStage, stageName: testStage.name, y: 3, x: 9, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p5", stage: testStage, stageName: testStage.name, y: 4, x: 9, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p6", stage: testStage, stageName: testStage.name, y: 5, x: 7, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p7", stage: testStage, stageName: testStage.name, y: 5, x: 8, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
	}

	for _, player := range players {
		player.placeOnStage()
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
		testStage.tiles[3][7].interactable == nil ||
		testStage.tiles[4][9].interactable == nil ||
		testStage.tiles[5][7].interactable == nil ||
		testStage.tiles[5][8].interactable == nil {
		t.Error("Final state of test-walls-interactable does not have correct 5 interactables")
	}
}
func TestSurroundedPushableSquareMultipleThreads(t *testing.T) {
	loadFromJson()
	testStage := createStageByName("test-walls-interactable")
	updatesForPlayer := make(chan Update)
	go drainChannel(updatesForPlayer)

	if testStage.tiles[14][2].interactable == nil ||
		testStage.tiles[3][7].interactable == nil ||
		testStage.tiles[3][8].interactable == nil ||
		testStage.tiles[4][7].interactable == nil ||
		testStage.tiles[4][8].interactable == nil {
		t.Error("Initial state of test-walls-interactable does not have correct 5 interactables")
	}

	// Place players around the 2x2 square of pushable tiles (3,7) (3,8) (4,7) (4,8)
	players := []*Player{
		{id: "p0", stage: testStage, stageName: testStage.name, y: 2, x: 7, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p1", stage: testStage, stageName: testStage.name, y: 2, x: 8, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p2", stage: testStage, stageName: testStage.name, y: 3, x: 6, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p3", stage: testStage, stageName: testStage.name, y: 4, x: 6, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p4", stage: testStage, stageName: testStage.name, y: 3, x: 9, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p5", stage: testStage, stageName: testStage.name, y: 4, x: 9, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p6", stage: testStage, stageName: testStage.name, y: 5, x: 7, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p7", stage: testStage, stageName: testStage.name, y: 5, x: 8, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
	}

	for _, player := range players {
		player.placeOnStage()
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
	updatesForPlayer := make(chan Update)
	go drainChannel(updatesForPlayer)

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

	// Place players around the 2x2 square of pushable tiles (3,7) (3,8) (4,7) (4,8)
	players := []*Player{
		{id: "p0", stage: testStage, stageName: testStage.name, y: 2, x: 7, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p1", stage: testStage, stageName: testStage.name, y: 2, x: 8, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p2", stage: testStage, stageName: testStage.name, y: 3, x: 6, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p3", stage: testStage, stageName: testStage.name, y: 4, x: 6, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p4", stage: testStage, stageName: testStage.name, y: 3, x: 9, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p5", stage: testStage, stageName: testStage.name, y: 4, x: 9, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p6", stage: testStage, stageName: testStage.name, y: 5, x: 7, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p7", stage: testStage, stageName: testStage.name, y: 5, x: 8, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
		{id: "p8", stage: testStage, stageName: testStage.name, y: 14, x: 7, updates: updatesForPlayer, actions: createDefaultActions(), health: 100},
	}

	for _, player := range players {
		player.placeOnStage()
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func(wg *sync.WaitGroup) { defer wg.Done(); players[0].moveSouth() }(&wg)
	go func(wg *sync.WaitGroup) { defer wg.Done(); players[8].moveNorth() }(&wg)

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

// Inclusion function using the logistic function
func inclusionProbability(d, r, fuzz float64) float64 {
	return 1 / (1 + math.Exp((d-r)/fuzz))
	/*
		if d < r-fuzz {
			return 1.0
		} else if d > r+fuzz {
			return 0.0
		} else {
			return 1 - (d-(r-fuzz))/(2*fuzz)
		}
	*/
}

func TestCircleGeneration(t *testing.T) {
	n := 90             // Size of the grid
	r := float64(n) / 3 // Radius of the circle
	fuzz := 4.7         // Fuzz factor; adjust this to vary sharpness

	// Center of the circle
	cx, cy := float64(n)/2, float64(n)/2

	// Initialize the bit array
	grid := make([][]int, n)
	for i := range grid {
		grid[i] = make([]int, n)
	}

	rand.Seed(time.Now().UnixNano())

	// Fill the grid
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			dx := float64(i) - cx
			dy := float64(j) - cy
			d := math.Hypot(dx, dy)
			p := inclusionProbability(d, r, fuzz)
			if rand.Float64() < p {
				grid[i][j] = 1
			} else {
				grid[i][j] = 0
			}
		}
	}

	// For demonstration, print the grid to the console
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if grid[i][j] == 1 {
				fmt.Print("██")
			} else {
				fmt.Print("  ")
			}
		}
		fmt.Println()
	}
}
