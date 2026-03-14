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
