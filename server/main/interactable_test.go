package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestEnsureInteractableWillPush(t *testing.T) {
	loadFromJson()
	testStage := createStageByName("test-walls-interactable")
	updatesForPlayer := make(chan []byte)
	defer close(updatesForPlayer)
	go drainChannel(updatesForPlayer)

	player := &Player{
		id:       "tp",
		actions:  createDefaultActions(),
		updates:  updatesForPlayer,
		tangible: true,
		camera:   newCamera(updatesForPlayer),
	}
	player.placeOnStage(testStage, 14, 1)

	if len(player.getTileSync().stage.tiles[14][1].characterMap) == 0 {
		t.Error("Player did not spawn at correct location")
	}

	if player.getTileSync().stage.tiles[14][2].interactable == nil || !player.getTileSync().stage.tiles[14][2].interactable.pushable {
		t.Error("test-walls-interactable should have pushable at 14,2")
	}

	moveEast(player)
	moveEast(player)
	moveWest(player)
	moveNorth(player)

	if player.getTileSync().stage.tiles[14][4].interactable == nil {
		t.Error("Interactable did not push")
	}

	if player.getTileSync().stage.tiles[14][2].interactable != nil {
		t.Error("Interactable still at starting location despite being pushed")
	}

	if len(player.getTileSync().stage.tiles[13][2].characterMap) == 0 {
		t.Error("Player has not moved correctly:")
		fmt.Printf("Y%dX%d", player.getTileSync().y, player.getTileSync().x)
	}
}

func TestSurroundedPushableSquare(t *testing.T) {
	loadFromJson()
	testStage := createStageByName("test-walls-interactable")
	updatesForPlayer := make(chan []byte)
	defer close(updatesForPlayer)
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
		{id: "p0", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
		{id: "p1", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
		{id: "p2", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
		{id: "p3", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
		{id: "p4", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
		{id: "p5", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
		{id: "p6", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
		{id: "p7", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
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
	moveSouth(players[0]) // p1 pushes from (2,7) to (3,7)
	moveSouth(players[1]) // p2 pushes from (2,8) to (3,3)
	moveEast(players[2])  // p3 pushes from (3,6) to (3,7)
	moveEast(players[3])
	moveWest(players[4]) // p5 pushes from (3,9) . . .
	moveWest(players[5])
	moveNorth(players[6])
	moveNorth(players[7])

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
	defer close(updatesForPlayer)
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
		{id: "p0", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
		{id: "p1", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
		{id: "p2", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
		{id: "p3", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
		{id: "p4", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
		{id: "p5", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
		{id: "p6", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
		{id: "p7", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
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
	go func(wg *sync.WaitGroup) { defer wg.Done(); moveSouth(players[0]) }(&wg)
	go func(wg *sync.WaitGroup) { defer wg.Done(); moveSouth(players[1]) }(&wg)
	go func(wg *sync.WaitGroup) { defer wg.Done(); moveEast(players[2]) }(&wg)
	go func(wg *sync.WaitGroup) { defer wg.Done(); moveEast(players[3]) }(&wg)
	go func(wg *sync.WaitGroup) { defer wg.Done(); moveWest(players[4]) }(&wg)
	go func(wg *sync.WaitGroup) { defer wg.Done(); moveWest(players[5]) }(&wg)
	go func(wg *sync.WaitGroup) { defer wg.Done(); moveNorth(players[6]) }(&wg)
	go func(wg *sync.WaitGroup) { defer wg.Done(); moveNorth(players[7]) }(&wg)

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
	defer close(updatesForPlayer)
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

	// Place 2 players at ends of long interactable line
	players := []*Player{
		{id: "p0", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
		{id: "p1", updates: updatesForPlayer, actions: createDefaultActions(), tangible: true, camera: newCamera(updatesForPlayer)},
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

	go func(wg *sync.WaitGroup) { defer wg.Done(); moveSouth(players[0]) }(&wg)
	go func(wg *sync.WaitGroup) { defer wg.Done(); moveNorth(players[1]) }(&wg)

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

func TestInteractableStateGates(t *testing.T) {
	incoming := &Interactable{state: "door-open"}

	if !interactableStateIs("door-open")(incoming, nil) {
		t.Fatal("interactableStateIs should match exact incoming state")
	}
	if interactableStateIs("door-closed")(incoming, nil) {
		t.Fatal("interactableStateIs should fail when states differ")
	}
	if !interactableStateIsNot("door-closed")(incoming, nil) {
		t.Fatal("interactableStateIsNot should pass when states differ")
	}
	if interactableStateIsNot("door-open")(incoming, nil) {
		t.Fatal("interactableStateIsNot should fail when states are equal")
	}
	if !interactableStateContains("open")(incoming, nil) {
		t.Fatal("interactableStateContains should match substring")
	}
	if interactableStateContains("sealed")(incoming, nil) {
		t.Fatal("interactableStateContains should fail for missing substring")
	}

	if interactableStateIs("door-open")(nil, nil) {
		t.Fatal("state gates should not match nil incoming interactable")
	}
	if interactableStateIs("")(incoming, nil) || interactableStateIsNot("")(incoming, nil) || interactableStateContains("")(incoming, nil) {
		t.Fatal("state gates should not match empty configured state fragments")
	}
}

func TestResolveReactionRulesWithStateGates(t *testing.T) {
	rules := []ReactionRule{
		{
			ReactsWith:     "interactableStateIs",
			ReactsWithArgs: []string{"armed"},
			Reaction:       "pass",
		},
		{
			ReactsWith:     "interactableStateContains",
			ReactsWithArgs: []string{"open"},
			Reaction:       "pass",
		},
	}

	resolved := resolveReactionRules(rules)
	if len(resolved) != 2 {
		t.Fatalf("expected 2 resolved rules, got %d", len(resolved))
	}

	if !resolved[0].ReactsWith(&Interactable{state: "armed"}, nil) {
		t.Fatal("interactableStateIs rule should match expected state")
	}
	if resolved[0].ReactsWith(&Interactable{state: "disarmed"}, nil) {
		t.Fatal("interactableStateIs rule should fail on non-matching state")
	}
	if !resolved[1].ReactsWith(&Interactable{state: "airlock-open"}, nil) {
		t.Fatal("interactableStateContains rule should match substring")
	}
}

func TestResolveReactionRulesWithTransmitPushAll(t *testing.T) {
	rules := []ReactionRule{
		{
			ReactsWith: "interactableIsNil",
			Reaction:   "transmitPushAll",
		},
	}

	resolved := resolveReactionRules(rules)
	if len(resolved) != 1 {
		t.Fatalf("expected 1 resolved rule, got %d", len(resolved))
	}
	if resolved[0].ReactionWithOffset == nil {
		t.Fatal("expected transmitPushAll to resolve as an offset-aware reaction")
	}
}

func TestResolveReactionRulesWithFilteredTransmitPush(t *testing.T) {
	rules := []ReactionRule{
		{
			ReactsWith:   "interactableIsNil",
			Reaction:     "transmitPushByState",
			ReactionArgs: []string{"armed"},
		},
		{
			ReactsWith:   "interactableIsNil",
			Reaction:     "transmitPushByName",
			ReactionArgs: []string{"box"},
		},
	}

	resolved := resolveReactionRules(rules)
	if len(resolved) != 2 {
		t.Fatalf("expected 2 resolved rules, got %d", len(resolved))
	}
	if resolved[0].ReactionWithOffset == nil || resolved[1].ReactionWithOffset == nil {
		t.Fatal("expected filtered transmit push reactions to resolve as offset-aware reactions")
	}
}

func TestTransmitPushAllMovesOtherInteractables(t *testing.T) {
	area := Area{
		Name: "transmit-push-all-test",
		Tiles: [][]Material{{
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{
				Name:     "transmitter",
				CssClass: "transmitter",
				Walkable: true,
				ReactionRules: []ReactionRule{
					{ReactsWith: "interactableIsNil", Reaction: "transmitPushAll"},
				},
			},
			{
				Name:     "box",
				CssClass: "box",
				Pushable: true,
				Walkable: true,
			},
			nil,
		}},
	}

	stage := createStageFromArea(area)
	if stage == nil {
		t.Fatal("expected stage")
	}

	transmitterTile := stage.tiles[0][1]
	transmitter := transmitterTile.interactable
	if transmitter == nil {
		t.Fatal("expected transmitter interactable at 0,1")
	}

	initiator := &Player{
		world:        &World{worldStages: map[string]*Stage{}},
		playerStages: map[string]*Stage{},
	}
	if !transmitter.React(nil, initiator, transmitterTile, 0, 1) {
		t.Fatal("expected transmit reaction to trigger")
	}

	if stage.tiles[0][2].interactable != nil {
		t.Fatal("expected box to be moved from source tile")
	}
	if stage.tiles[0][3].interactable == nil || stage.tiles[0][3].interactable.name != "box" {
		t.Fatal("expected box to be transmitted one tile east")
	}
	if transmitterTile.interactable == nil || transmitterTile.interactable.name != "transmitter" {
		t.Fatal("expected transmitter to remain in place")
	}
}

func TestTransmitPushAllDoesNotDoublePushWhenMovingRight(t *testing.T) {
	area := Area{
		Name: "transmit-push-all-right-order",
		Tiles: [][]Material{{
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{
				Name:     "transmitter",
				CssClass: "transmitter",
				Walkable: true,
				ReactionRules: []ReactionRule{
					{ReactsWith: "interactableIsNil", Reaction: "transmitPushAll"},
				},
			},
			{
				Name:     "box",
				CssClass: "box",
				Pushable: true,
				Walkable: true,
			},
			nil,
			nil,
		}},
	}

	stage := createStageFromArea(area)
	if stage == nil {
		t.Fatal("expected stage")
	}

	transmitterTile := stage.tiles[0][1]
	transmitter := transmitterTile.interactable
	if transmitter == nil {
		t.Fatal("expected transmitter interactable at 0,1")
	}

	initiator := &Player{world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{}}
	if !transmitter.React(nil, initiator, transmitterTile, 0, 1) {
		t.Fatal("expected transmit reaction to trigger")
	}

	if stage.tiles[0][3].interactable == nil || stage.tiles[0][3].interactable.name != "box" {
		t.Fatal("expected box to move one tile east to 0,3")
	}
	if stage.tiles[0][4].interactable != nil {
		t.Fatal("expected box not to be pushed twice to 0,4")
	}
}

func TestTransmitPushAllDoesNotDoublePushWhenMovingDown(t *testing.T) {
	area := Area{
		Name: "transmit-push-all-down-order",
		Tiles: [][]Material{
			{{Walkable: true}},
			{{Walkable: true}},
			{{Walkable: true}},
			{{Walkable: true}},
			{{Walkable: true}},
		},
		Interactables: [][]*InteractableDescription{
			{nil},
			{{
				Name:     "transmitter",
				CssClass: "transmitter",
				Walkable: true,
				ReactionRules: []ReactionRule{
					{ReactsWith: "interactableIsNil", Reaction: "transmitPushAll"},
				},
			}},
			{{
				Name:     "box",
				CssClass: "box",
				Pushable: true,
				Walkable: true,
			}},
			{nil},
			{nil},
		},
	}

	stage := createStageFromArea(area)
	if stage == nil {
		t.Fatal("expected stage")
	}

	transmitterTile := stage.tiles[1][0]
	transmitter := transmitterTile.interactable
	if transmitter == nil {
		t.Fatal("expected transmitter interactable at 1,0")
	}

	initiator := &Player{world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{}}
	if !transmitter.React(nil, initiator, transmitterTile, 1, 0) {
		t.Fatal("expected transmit reaction to trigger")
	}

	if stage.tiles[3][0].interactable == nil || stage.tiles[3][0].interactable.name != "box" {
		t.Fatal("expected box to move one tile south to 3,0")
	}
	if stage.tiles[4][0].interactable != nil {
		t.Fatal("expected box not to be pushed twice to 4,0")
	}
}

func TestTransmitPushByStateMovesOnlyMatchingState(t *testing.T) {
	area := Area{
		Name: "transmit-push-by-state",
		Tiles: [][]Material{{
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{
				Name:     "transmitter",
				CssClass: "transmitter",
				Walkable: true,
				ReactionRules: []ReactionRule{
					{ReactsWith: "interactableIsNil", Reaction: "transmitPushByState", ReactionArgs: []string{"armed"}},
				},
			},
			{
				Name:     "armed-box",
				State:    "armed",
				CssClass: "box",
				Pushable: true,
				Walkable: true,
			},
			nil,
			{
				Name:     "idle-box",
				State:    "idle",
				CssClass: "box",
				Pushable: true,
				Walkable: true,
			},
			nil,
			nil,
			nil,
		}},
	}

	stage := createStageFromArea(area)
	if stage == nil {
		t.Fatal("expected stage")
	}

	transmitterTile := stage.tiles[0][1]
	transmitter := transmitterTile.interactable
	if transmitter == nil {
		t.Fatal("expected transmitter interactable at 0,1")
	}

	initiator := &Player{world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{}}
	if !transmitter.React(nil, initiator, transmitterTile, 0, 1) {
		t.Fatal("expected transmit reaction to trigger")
	}

	if stage.tiles[0][3].interactable == nil || stage.tiles[0][3].interactable.name != "armed-box" {
		t.Fatal("expected armed-box to move one tile east")
	}
	if stage.tiles[0][4].interactable == nil || stage.tiles[0][4].interactable.name != "idle-box" {
		t.Fatal("expected idle-box to remain in place")
	}
}

func TestTransmitPushByNameMovesOnlyMatchingName(t *testing.T) {
	area := Area{
		Name: "transmit-push-by-type",
		Tiles: [][]Material{{
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{
				Name:     "transmitter",
				CssClass: "transmitter",
				Walkable: true,
				ReactionRules: []ReactionRule{
					{ReactsWith: "interactableIsNil", Reaction: "transmitPushByName", ReactionArgs: []string{"box"}},
				},
			},
			{
				Name:     "box",
				CssClass: "box",
				Pushable: true,
				Walkable: true,
			},
			nil,
			{
				Name:     "barrel",
				CssClass: "barrel",
				Pushable: true,
				Walkable: true,
			},
			nil,
		}},
	}

	stage := createStageFromArea(area)
	if stage == nil {
		t.Fatal("expected stage")
	}

	transmitterTile := stage.tiles[0][1]
	transmitter := transmitterTile.interactable
	if transmitter == nil {
		t.Fatal("expected transmitter interactable at 0,1")
	}

	initiator := &Player{world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{}}
	if !transmitter.React(nil, initiator, transmitterTile, 0, 1) {
		t.Fatal("expected transmit reaction to trigger")
	}

	if stage.tiles[0][3].interactable == nil || stage.tiles[0][3].interactable.name != "box" {
		t.Fatal("expected box to move one tile east")
	}
	if stage.tiles[0][4].interactable == nil || stage.tiles[0][4].interactable.name != "barrel" {
		t.Fatal("expected barrel to remain in place")
	}
}

func TestTransmitPushAllConcurrentDirectionsCompletesWithoutDeadlock(t *testing.T) {
	area := Area{
		Name: "transmit-push-all-concurrency",
		Tiles: [][]Material{
			{{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}},
			{{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}},
			{{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}},
			{{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}},
			{{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}},
		},
		Interactables: [][]*InteractableDescription{
			{nil, nil, nil, nil, nil},
			{nil, nil, nil, nil, nil},
			{nil, nil, {
				Name:     "transmitter",
				CssClass: "transmitter",
				Walkable: true,
				ReactionRules: []ReactionRule{
					{ReactsWith: "interactableIsNil", Reaction: "transmitPushAll"},
				},
			}, {
				Name:     "box-east",
				CssClass: "box",
				Pushable: true,
				Walkable: true,
			}, nil},
			{nil, nil, {
				Name:     "box-south",
				CssClass: "box",
				Pushable: true,
				Walkable: true,
			}, nil, nil},
			{nil, nil, nil, nil, nil},
		},
	}

	stage := createStageFromArea(area)
	if stage == nil {
		t.Fatal("expected stage")
	}

	transmitterTile := stage.tiles[2][2]
	transmitter := transmitterTile.interactable
	if transmitter == nil {
		t.Fatal("expected transmitter interactable at 2,2")
	}

	const workerCount = 8
	const iterations = 150
	directions := [][2]int{{0, 1}, {1, 0}, {0, -1}, {-1, 0}}

	start := make(chan struct{})
	var wg sync.WaitGroup
	errCh := make(chan string, workerCount)

	for i := 0; i < workerCount; i++ {
		yOff := directions[i%len(directions)][0]
		xOff := directions[i%len(directions)][1]
		wg.Add(1)
		go func(index, yOff, xOff int) {
			defer wg.Done()
			<-start
			initiator := &Player{world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{}}
			for j := 0; j < iterations; j++ {
				if !transmitter.React(nil, initiator, transmitterTile, yOff, xOff) {
					errCh <- fmt.Sprintf("worker %d: transmit reaction did not trigger", index)
					return
				}
			}
		}(i, yOff, xOff)
	}

	close(start)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		close(errCh)
		for err := range errCh {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("concurrent transmitPushAll appears deadlocked or stalled")
	}

	if transmitterTile.interactable == nil || transmitterTile.interactable.name != "transmitter" {
		t.Fatal("expected transmitter to remain at source after concurrent pushes")
	}
}

func TestTransmitPushStickyCollisionDoesNotCollapseInteractables(t *testing.T) {
	area := Area{
		Name: "transmit-push-sticky-collision",
		Tiles: [][]Material{{
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
			{Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{
				Name:     "transmitter",
				CssClass: "transmitter",
				Walkable: true,
				ReactionRules: []ReactionRule{
					{ReactsWith: "interactableIsNil", Reaction: "transmitPushAll"},
				},
			},
			{
				Name:         "sticky-a",
				CssClass:     "sticky",
				Pushable:     true,
				Walkable:     true,
				StickyGroups: []string{"sticky"},
			},
			nil,
			{
				Name:         "sticky-b",
				CssClass:     "sticky",
				Pushable:     true,
				Walkable:     true,
				StickyGroups: []string{"sticky"},
			},
			{
				Name:     "wall",
				CssClass: "wall",
				Pushable: false,
				Walkable: false,
			},
		}},
	}

	stage := createStageFromArea(area)
	if stage == nil {
		t.Fatal("expected stage")
	}

	transmitterTile := stage.tiles[0][1]
	transmitter := transmitterTile.interactable
	if transmitter == nil {
		t.Fatal("expected transmitter interactable at 0,1")
	}

	initiator := &Player{world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{}}

	if !transmitter.React(nil, initiator, transmitterTile, 0, 1) {
		t.Fatal("expected first transmit reaction to trigger")
	}
	if !transmitter.React(nil, initiator, transmitterTile, 0, 1) {
		t.Fatal("expected second transmit reaction to trigger")
	}

	stickyCount := 0
	for _, tile := range stage.tiles[0] {
		if tile.interactable != nil && (tile.interactable.name == "sticky-a" || tile.interactable.name == "sticky-b") {
			stickyCount++
		}
	}

	if stickyCount != 2 {
		t.Fatalf("expected 2 sticky interactables after transmitted collision, got %d", stickyCount)
	}
	if stage.tiles[0][3].interactable == nil || stage.tiles[0][4].interactable == nil {
		t.Fatal("expected sticky interactables to remain as two occupied tiles at 0,3 and 0,4")
	}
}

///////////////////////////////////////////////////////////////////
// Sticky Group (Polyomino) Push Tests

func TestStickyPairPushesEast(t *testing.T) {
	// [_] [A] [B] [_]  -- push east -->  [_] [_] [A] [B]
	area := Area{
		Name: "sticky-pair-east",
		Tiles: [][]Material{{
			{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{Name: "sA", CssClass: "a", Pushable: true, StickyGroups: []string{"sticky"}},
			{Name: "sB", CssClass: "b", Pushable: true, StickyGroups: []string{"sticky"}},
			nil,
		}},
	}

	stage := createStageFromArea(area)
	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	p := &Player{
		id: "tp", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}
	p.placeOnStage(stage, 0, 0)
	moveEast(p)

	if stage.tiles[0][1].interactable != nil {
		t.Fatal("expected tile 0,1 to be empty after push")
	}
	if ia := stage.tiles[0][2].interactable; ia == nil || ia.name != "sA" {
		t.Fatal("expected sA at 0,2")
	}
	if ia := stage.tiles[0][3].interactable; ia == nil || ia.name != "sB" {
		t.Fatal("expected sB at 0,3")
	}
}

func TestStickyInteriorNilIncomingPushDoesNotDropMember(t *testing.T) {
	area := Area{
		Name: "sticky-interior-nil-incoming",
		Tiles: [][]Material{
			{{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}},
			{{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}},
		},
		Interactables: [][]*InteractableDescription{
			{{Name: "a", CssClass: "a", Pushable: true, StickyGroups: []string{"sticky"}}, {Name: "b", CssClass: "b", Pushable: true, StickyGroups: []string{"sticky"}}, nil, nil},
			{{Name: "c", CssClass: "c", Pushable: true, StickyGroups: []string{"sticky"}}, nil, nil, nil},
		},
	}

	stage := createStageFromArea(area)
	if stage == nil {
		t.Fatal("expected stage")
	}

	p := &Player{world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{}}

	// Start from interior tile b with nil incoming (transmit-style behavior).
	if ok := p.push(stage.tiles[0][1], nil, 0, 1); !ok {
		t.Fatal("expected interior nil-incoming sticky push east to succeed")
	}

	count := 0
	for y := range stage.tiles {
		for x := range stage.tiles[y] {
			if stage.tiles[y][x].interactable != nil {
				count++
			}
		}
	}

	if count != 3 {
		t.Fatalf("expected all 3 sticky members to remain after push, got %d", count)
	}
	if ia := stage.tiles[0][1].interactable; ia == nil || ia.name != "a" {
		t.Fatal("expected a at 0,1")
	}
	if ia := stage.tiles[0][2].interactable; ia == nil || ia.name != "b" {
		t.Fatal("expected b at 0,2")
	}
	if ia := stage.tiles[1][1].interactable; ia == nil || ia.name != "c" {
		t.Fatal("expected c at 1,1")
	}
}

func TestStickyPairPushesWest(t *testing.T) {
	// [_] [A] [B] [_]  -- push west <--  [A] [B] [_] [_]
	area := Area{
		Name: "sticky-pair-west",
		Tiles: [][]Material{{
			{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{Name: "sA", CssClass: "a", Pushable: true, StickyGroups: []string{"sticky"}},
			{Name: "sB", CssClass: "b", Pushable: true, StickyGroups: []string{"sticky"}},
			nil,
		}},
	}

	stage := createStageFromArea(area)
	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	p := &Player{
		id: "tp", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}
	p.placeOnStage(stage, 0, 3)
	moveWest(p)

	if ia := stage.tiles[0][0].interactable; ia == nil || ia.name != "sA" {
		t.Fatal("expected sA at 0,0")
	}
	if ia := stage.tiles[0][1].interactable; ia == nil || ia.name != "sB" {
		t.Fatal("expected sB at 0,1")
	}
	if stage.tiles[0][2].interactable != nil {
		t.Fatal("expected tile 0,2 to be empty after push")
	}
}

func TestStickyLShapePushesSouth(t *testing.T) {
	// L-shape: A at (0,1), B at (1,1), C at (1,2)
	// Push south from (0,0)->moveEast onto A
	//
	// Before:        After:
	//  _ A _ _        _ _ _ _
	//  _ B C _        _ A _ _
	//  _ _ _ _        _ B C _
	//  _ _ _ _        _ _ _ _
	area := Area{
		Name: "sticky-L-south",
		Tiles: [][]Material{
			{{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}},
			{{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}},
			{{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}},
			{{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}},
		},
		Interactables: [][]*InteractableDescription{
			{nil, {Name: "sA", CssClass: "a", Pushable: true, StickyGroups: []string{"sticky"}}, nil, nil},
			{nil, {Name: "sB", CssClass: "b", Pushable: true, StickyGroups: []string{"sticky"}}, {Name: "sC", CssClass: "c", Pushable: true, StickyGroups: []string{"sticky"}}, nil},
			{nil, nil, nil, nil},
			{nil, nil, nil, nil},
		},
	}

	stage := createStageFromArea(area)
	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	// Player pushes south from above A
	p := &Player{
		id: "tp", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}

	// Place player at (0,0), then push towards A via moving east (which triggers push on A's tile)
	// Actually, to push south, player needs to be north of A and move south.
	// A is at (0,1). We can't be at (-1,1). Let's rearrange.
	// Put player at (0,2) and have it push west into C. But we want south push.

	// Let's test via direct push call instead:
	aTile := stage.tiles[0][1]
	p.placeOnStage(stage, 0, 0)

	// Direct push: push A's tile southward
	result := p.push(aTile, nil, 1, 0)
	if !result {
		t.Fatal("expected sticky L-shape push south to succeed")
	}

	if stage.tiles[0][1].interactable != nil {
		t.Fatal("expected 0,1 to be empty")
	}
	if ia := stage.tiles[1][1].interactable; ia == nil || ia.name != "sA" {
		t.Fatal("expected sA at 1,1")
	}
	if ia := stage.tiles[2][1].interactable; ia == nil || ia.name != "sB" {
		t.Fatal("expected sB at 2,1")
	}
	if ia := stage.tiles[2][2].interactable; ia == nil || ia.name != "sC" {
		t.Fatal("expected sC at 2,2")
	}
}

func TestStickyGroupBlockedByWall(t *testing.T) {
	// [wall] [A] [B] [_]  -- push west should fail (wall blocks A)
	area := Area{
		Name: "sticky-blocked-wall",
		Tiles: [][]Material{{
			{Walkable: false}, {Walkable: true}, {Walkable: true}, {Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{Name: "sA", CssClass: "a", Pushable: true, StickyGroups: []string{"sticky"}},
			{Name: "sB", CssClass: "b", Pushable: true, StickyGroups: []string{"sticky"}},
			nil,
		}},
	}

	stage := createStageFromArea(area)
	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	p := &Player{
		id: "tp", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}
	p.placeOnStage(stage, 0, 3)
	moveWest(p)

	// Group should NOT have moved
	if ia := stage.tiles[0][1].interactable; ia == nil || ia.name != "sA" {
		t.Fatal("expected sA to remain at 0,1 (blocked by wall)")
	}
	if ia := stage.tiles[0][2].interactable; ia == nil || ia.name != "sB" {
		t.Fatal("expected sB to remain at 0,2 (blocked by wall)")
	}
}

func TestStickyGroupBlockedByOccupiedTile(t *testing.T) {
	// [_] [A] [B] [nonSticky] -- push east should fail (static block blocks B)
	area := Area{
		Name: "sticky-blocked-occupied",
		Tiles: [][]Material{{
			{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{Name: "sA", CssClass: "a", Pushable: true, StickyGroups: []string{"sticky"}},
			{Name: "sB", CssClass: "b", Pushable: true, StickyGroups: []string{"sticky"}},
			{Name: "blocker", CssClass: "x"},
		}},
	}

	stage := createStageFromArea(area)
	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	p := &Player{
		id: "tp", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}
	p.placeOnStage(stage, 0, 0)
	moveEast(p)

	// Group should NOT have moved
	if ia := stage.tiles[0][1].interactable; ia == nil || ia.name != "sA" {
		t.Fatal("expected sA to remain at 0,1")
	}
	if ia := stage.tiles[0][2].interactable; ia == nil || ia.name != "sB" {
		t.Fatal("expected sB to remain at 0,2")
	}
	if ia := stage.tiles[0][3].interactable; ia == nil || ia.name != "blocker" {
		t.Fatal("expected blocker to remain at 0,3")
	}
}

func TestStickyGroupBlockedByEdge(t *testing.T) {
	// [A] [B] -- push west should fail (edge of stage blocks A)
	area := Area{
		Name: "sticky-blocked-edge",
		Tiles: [][]Material{{
			{Walkable: true}, {Walkable: true}, {Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			{Name: "sA", CssClass: "a", Pushable: true, StickyGroups: []string{"sticky"}},
			{Name: "sB", CssClass: "b", Pushable: true, StickyGroups: []string{"sticky"}},
			nil,
		}},
	}

	stage := createStageFromArea(area)
	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	p := &Player{
		id: "tp", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}
	p.placeOnStage(stage, 0, 2)
	moveWest(p)

	// Group should NOT have moved
	if ia := stage.tiles[0][0].interactable; ia == nil || ia.name != "sA" {
		t.Fatal("expected sA to remain at 0,0")
	}
	if ia := stage.tiles[0][1].interactable; ia == nil || ia.name != "sB" {
		t.Fatal("expected sB to remain at 0,1")
	}
}

func TestStickyGroupCrossesStageBoundaryEast(t *testing.T) {
	leftArea := Area{
		Name: "sticky-boundary-left",
		East: "sticky-boundary-right",
		Tiles: [][]Material{{
			{Walkable: true}, {Walkable: true}, {Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{Name: "sA", CssClass: "a", Pushable: true, StickyGroups: []string{"sticky"}},
			{Name: "sB", CssClass: "b", Pushable: true, StickyGroups: []string{"sticky"}},
		}},
	}

	rightArea := Area{
		Name: "sticky-boundary-right",
		West: "sticky-boundary-left",
		Tiles: [][]Material{{
			{Walkable: true}, {Walkable: true}, {Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil, nil, nil,
		}},
	}

	leftStage := createStageFromArea(leftArea)
	rightStage := createStageFromArea(rightArea)

	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	p := &Player{
		id: "tp", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{
			leftStage.name:  leftStage,
			rightStage.name: rightStage,
		}},
		playerStages: map[string]*Stage{},
	}
	p.placeOnStage(leftStage, 0, 0)

	if !p.push(leftStage.tiles[0][1], nil, 0, 1) {
		t.Fatal("expected sticky push east across stage boundary to succeed")
	}

	if leftStage.tiles[0][1].interactable != nil {
		t.Fatal("expected left stage tile 0,1 to be empty after boundary push")
	}
	if ia := leftStage.tiles[0][2].interactable; ia == nil || ia.name != "sA" {
		t.Fatal("expected sA at left stage 0,2")
	}
	if ia := rightStage.tiles[0][0].interactable; ia == nil || ia.name != "sB" {
		t.Fatal("expected sB at right stage 0,0")
	}
}

func TestBallPushesIntoStickyGroup(t *testing.T) {
	// [_] [ball] [sA] [sB] [_]  -- push east -->  [_] [_] [ball] [sA] [sB]
	area := Area{
		Name: "ball-into-sticky",
		Tiles: [][]Material{{
			{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{Name: "ball", CssClass: "ball", Pushable: true},
			{Name: "sA", CssClass: "a", Pushable: true, StickyGroups: []string{"sticky"}},
			{Name: "sB", CssClass: "b", Pushable: true, StickyGroups: []string{"sticky"}},
			nil,
		}},
	}

	stage := createStageFromArea(area)
	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	p := &Player{
		id: "tp", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}
	p.placeOnStage(stage, 0, 0)
	moveEast(p)

	if stage.tiles[0][1].interactable != nil {
		t.Fatal("expected tile 0,1 to be empty")
	}
	if ia := stage.tiles[0][2].interactable; ia == nil || ia.name != "ball" {
		t.Fatal("expected ball at 0,2")
	}
	if ia := stage.tiles[0][3].interactable; ia == nil || ia.name != "sA" {
		t.Fatal("expected sA at 0,3")
	}
	if ia := stage.tiles[0][4].interactable; ia == nil || ia.name != "sB" {
		t.Fatal("expected sB at 0,4")
	}
}

func TestBallPushesIntoStickyGroupWest(t *testing.T) {
	// [_] [sA] [sB] [ball] [_]  -- push west -->  [sA] [sB] [ball] [_] [_]
	area := Area{
		Name: "ball-into-sticky-west",
		Tiles: [][]Material{{
			{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{Name: "sA", CssClass: "a", Pushable: true, StickyGroups: []string{"sticky"}},
			{Name: "sB", CssClass: "b", Pushable: true, StickyGroups: []string{"sticky"}},
			{Name: "ball", CssClass: "ball", Pushable: true},
			nil,
		}},
	}

	stage := createStageFromArea(area)
	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	p := &Player{
		id: "tp", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}
	p.placeOnStage(stage, 0, 4)
	moveWest(p)

	if stage.tiles[0][3].interactable != nil {
		t.Fatal("expected tile 0,3 to be empty")
	}
	if ia := stage.tiles[0][2].interactable; ia == nil || ia.name != "ball" {
		t.Fatal("expected ball at 0,2")
	}
	if ia := stage.tiles[0][1].interactable; ia == nil || ia.name != "sB" {
		t.Fatal("expected sB at 0,1")
	}
	if ia := stage.tiles[0][0].interactable; ia == nil || ia.name != "sA" {
		t.Fatal("expected sA at 0,0")
	}
}

func TestPushableNonStickyPushedByGroup(t *testing.T) {
	// [_] [sA] [pushable] [sB] [_]
	// pushable is NOT sticky, so sA and sB should move independently
	area := Area{
		Name: "non-sticky-gap",
		Tiles: [][]Material{{
			{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{Name: "sA", CssClass: "a", Pushable: true, StickyGroups: []string{"sticky"}},
			{Name: "mid", CssClass: "m", Pushable: true},
			{Name: "sB", CssClass: "b", Pushable: true, StickyGroups: []string{"sticky"}},
			nil,
		}},
	}

	stage := createStageFromArea(area)
	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	p := &Player{
		id: "tp", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}
	p.placeOnStage(stage, 0, 0)
	moveEast(p)

	// Entire row should shift because mid is pushable.
	if ia := stage.tiles[0][2].interactable; ia == nil || ia.name != "sA" {
		t.Fatal("expected sA to shift to 0,2 (blocked by non-sticky mid)")
	}
	if ia := stage.tiles[0][3].interactable; ia == nil || ia.name != "mid" {
		t.Fatal("expected mid to remain at 0,3")
	}
}

func TestStickyStateSwitchWithStates(t *testing.T) {
	// Verify that sticky propagates through interactable states.
	area := Area{
		Name: "sticky-state-test",
		Tiles: [][]Material{{
			{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{
				Name:         "sA",
				CssClass:     "a",
				Pushable:     true,
				StickyGroups: []string{"sticky"},
				DefaultState: "default",
				States: map[string]InteractableStateDescription{
					"default": {CssClass: "a", Pushable: true, StickyGroups: []string{"sticky"}},
					"inert":   {CssClass: "a-off", Pushable: false},
				},
			},
			{Name: "sB", CssClass: "b", Pushable: true, StickyGroups: []string{"sticky"}},
			nil,
		}},
	}

	stage := createStageFromArea(area)
	sATile := stage.tiles[0][1]
	sA := sATile.interactable
	if sA == nil || len(sA.stickGroups) == 0 {
		t.Fatal("expected sA to be sticky in default state")
	}

	// Switch to inert state
	sA.applyState("inert")
	if len(sA.stickGroups) > 0 {
		t.Fatal("expected sA to NOT be sticky in inert state")
	}

	// Push should treat sA as non-sticky (and it's not pushable either in inert state)
	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	p := &Player{
		id: "tp", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}
	p.placeOnStage(stage, 0, 0)
	moveEast(p)

	// sA should be blocking (not pushable, not sticky, not walkable in inert)
	if ia := stage.tiles[0][1].interactable; ia == nil || ia.name != "sA" {
		t.Fatal("expected inert sA to remain at 0,1")
	}
}

func TestStickySquarePushesNorth(t *testing.T) {
	// 2x2 square of sticky blocks pushed north
	//  _ _ _ _       _ _ _ _
	//  _ _ _ _  -->  _ A B _
	//  _ A B _       _ C D _
	//  _ C D _       _ _ _ _
	//  _ _ _ _       _ _ _ _
	area := Area{
		Name: "sticky-square-north",
		Tiles: [][]Material{
			{{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}},
			{{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}},
			{{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}},
			{{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}},
			{{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}},
		},
		Interactables: [][]*InteractableDescription{
			{nil, nil, nil, nil},
			{nil, nil, nil, nil},
			{nil, {Name: "sA", CssClass: "a", Pushable: true, StickyGroups: []string{"sticky"}}, {Name: "sB", CssClass: "b", Pushable: true, StickyGroups: []string{"sticky"}}, nil},
			{nil, {Name: "sC", CssClass: "c", Pushable: true, StickyGroups: []string{"sticky"}}, {Name: "sD", CssClass: "d", Pushable: true, StickyGroups: []string{"sticky"}}, nil},
			{nil, nil, nil, nil},
		},
	}

	stage := createStageFromArea(area)
	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	p := &Player{
		id: "tp", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}
	p.placeOnStage(stage, 4, 1)
	moveNorth(p) // push into sC at (3,1)

	if ia := stage.tiles[1][1].interactable; ia == nil || ia.name != "sA" {
		t.Fatalf("expected sA at 1,1, got %v", ia)
	}
	if ia := stage.tiles[1][2].interactable; ia == nil || ia.name != "sB" {
		t.Fatalf("expected sB at 1,2, got %v", ia)
	}
	if ia := stage.tiles[2][1].interactable; ia == nil || ia.name != "sC" {
		t.Fatalf("expected sC at 2,1, got %v", ia)
	}
	if ia := stage.tiles[2][2].interactable; ia == nil || ia.name != "sD" {
		t.Fatalf("expected sD at 2,2, got %v", ia)
	}
	if stage.tiles[3][1].interactable != nil || stage.tiles[3][2].interactable != nil {
		t.Fatal("expected old positions (3,1) and (3,2) to be empty")
	}
}

func TestStickyGroupBlockedByNonPushableMember(t *testing.T) {
	area := Area{
		Name: "sticky-non-pushable-member",
		Tiles: [][]Material{{
			{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{Name: "sA", CssClass: "a", Pushable: true, StickyGroups: []string{"sticky"}},
			{Name: "sB", CssClass: "b", Pushable: false, StickyGroups: []string{"sticky"}},
			nil,
		}},
	}

	stage := createStageFromArea(area)
	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	p := &Player{
		id: "tp", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}
	p.placeOnStage(stage, 0, 0)
	moveEast(p)

	if ia := stage.tiles[0][1].interactable; ia == nil || ia.name != "sA" {
		t.Fatal("expected sA to remain at 0,1")
	}
	if ia := stage.tiles[0][2].interactable; ia == nil || ia.name != "sB" {
		t.Fatal("expected non-pushable sticky sB to remain at 0,2")
	}
}

func TestNonWalkableStickyBlocksMovement(t *testing.T) {
	area := Area{
		Name: "sticky-non-walkable",
		Tiles: [][]Material{{
			{Walkable: true}, {Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{Name: "sA", CssClass: "a", Pushable: false, Walkable: false, StickyGroups: []string{"sticky"}},
		}},
	}

	stage := createStageFromArea(area)
	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	p := &Player{
		id: "tp", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}
	p.placeOnStage(stage, 0, 0)
	moveEast(p)

	if p.tile != stage.tiles[0][0] {
		t.Fatal("expected player to remain in place when blocked by non-walkable sticky")
	}
	if ia := stage.tiles[0][1].interactable; ia == nil || ia.name != "sA" {
		t.Fatal("expected sticky interactable to remain at 0,1")
	}
}

func TestStickyGroupBlockedByPlayerOccupant(t *testing.T) {
	area := Area{
		Name: "sticky-blocked-by-player",
		Tiles: [][]Material{{
			{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{Name: "sA", CssClass: "a", Pushable: true, StickyGroups: []string{"sticky"}},
			{Name: "sB", CssClass: "b", Pushable: true, StickyGroups: []string{"sticky"}},
			nil,
		}},
	}

	stage := createStageFromArea(area)
	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	pusher := &Player{
		id: "pusher", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}
	blocker := &Player{
		id: "blocker", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}

	pusher.placeOnStage(stage, 0, 0)
	blocker.placeOnStage(stage, 0, 3)
	moveEast(pusher)

	if ia := stage.tiles[0][1].interactable; ia == nil || ia.name != "sA" {
		t.Fatal("expected sA to remain at 0,1")
	}
	if ia := stage.tiles[0][2].interactable; ia == nil || ia.name != "sB" {
		t.Fatal("expected sB to remain at 0,2")
	}
}

func TestStickyGroupMovesAsPolyominoGroup(t *testing.T) {
	area := Area{
		Name: "sticky-group-polyomino",
		Tiles: [][]Material{{
			{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{Name: "g1-a", CssClass: "f", Pushable: true, StickyGroups: []string{"g1"}},
			{Name: "g1-b", CssClass: "f", Pushable: true, StickyGroups: []string{"g1"}},
			nil,
		}},
	}

	stage := createStageFromArea(area)
	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	p := &Player{
		id: "tp", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}
	p.placeOnStage(stage, 0, 0)
	moveEast(p)

	if stage.tiles[0][1].interactable != nil {
		t.Fatal("expected original sticky-group head tile to be empty")
	}
	if ia := stage.tiles[0][2].interactable; ia == nil || ia.name != "g1-a" {
		t.Fatal("expected first sticky-group member at 0,2")
	}
	if ia := stage.tiles[0][3].interactable; ia == nil || ia.name != "g1-b" {
		t.Fatal("expected second sticky-group member at 0,3")
	}
}

func TestStickyIgnoresGroupAndLinksOtherSticky(t *testing.T) {
	area := Area{
		Name: "sticky-overrides-group",
		Tiles: [][]Material{{
			{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{Name: "sticky-a", CssClass: "a", Pushable: true, StickyGroups: []string{"sticky", "g1"}},
			{Name: "sticky-b", CssClass: "b", Pushable: true, StickyGroups: []string{"sticky", "g2"}},
			nil,
		}},
	}

	stage := createStageFromArea(area)
	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	p := &Player{
		id: "tp", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}
	p.placeOnStage(stage, 0, 0)
	moveEast(p)

	if stage.tiles[0][1].interactable != nil {
		t.Fatal("expected original sticky tile to be empty")
	}
	if ia := stage.tiles[0][2].interactable; ia == nil || ia.name != "sticky-a" {
		t.Fatal("expected sticky-a at 0,2")
	}
	if ia := stage.tiles[0][3].interactable; ia == nil || ia.name != "sticky-b" {
		t.Fatal("expected sticky-b at 0,3")
	}
}

func TestStickyGroupsSupportsMultipleMemberships(t *testing.T) {
	// [_] [B:left] [A:left,right] [C:right] [_] -- push east --> [_] [_] [B] [A] [C]
	area := Area{
		Name: "stick-group-multi-membership",
		Tiles: [][]Material{{
			{Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true}, {Walkable: true},
		}},
		Interactables: [][]*InteractableDescription{{
			nil,
			{Name: "b", CssClass: "b", Pushable: true, StickyGroups: []string{"left"}},
			{Name: "a", CssClass: "a", Pushable: true, StickyGroups: []string{"left", "right"}},
			{Name: "c", CssClass: "c", Pushable: true, StickyGroups: []string{"right"}},
			nil,
		}},
	}

	stage := createStageFromArea(area)
	ch := make(chan []byte, 128)
	defer close(ch)
	go drainChannel(ch)

	p := &Player{
		id: "tp", updates: ch, actions: createDefaultActions(),
		tangible: true, camera: newCamera(ch),
		world: &World{worldStages: map[string]*Stage{}}, playerStages: map[string]*Stage{},
	}
	p.placeOnStage(stage, 0, 0)
	moveEast(p)

	if stage.tiles[0][1].interactable != nil {
		t.Fatal("expected original head tile to be empty")
	}
	if ia := stage.tiles[0][2].interactable; ia == nil || ia.name != "b" {
		t.Fatal("expected b at 0,2")
	}
	if ia := stage.tiles[0][3].interactable; ia == nil || ia.name != "a" {
		t.Fatal("expected a at 0,3")
	}
	if ia := stage.tiles[0][4].interactable; ia == nil || ia.name != "c" {
		t.Fatal("expected c at 0,4")
	}
}
