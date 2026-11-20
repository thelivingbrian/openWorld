package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func frontierFixture(t *testing.T) (*World, map[string]Area) {
	t.Helper()
	col, err := loadSeedCollection("../../tools/main/data/collections", "bloop")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile("../../tools/main/data/colors/colors.json")
	palette, err := decodeWorldPalette(data)
	if err != nil {
		t.Fatal(err)
	}
	files, err := compileCollectionFiles(col, palette)
	if err != nil {
		t.Fatal(err)
	}
	var compiled []Area
	if err = json.Unmarshal(files["areas.json"], &compiled); err != nil {
		t.Fatal(err)
	}
	byName := map[string]Area{}
	for _, a := range compiled {
		byName[a.Name] = a
	}
	data, _ = os.ReadFile("../../tools/main/data/collections/bloop/manifest.json")
	var m WorldManifest
	if err = json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if err = validateWorldRules(m, byName); err != nil {
		t.Fatal(err)
	}
	return &World{App: App{config: &Configuration{manifest: m}}, worldPlayers: map[string]*Player{}, worldStages: map[string]*Stage{}}, byName
}
func fixturePlayer(world *World, stage *Stage, id string, y, x int) *Player {
	p := &Player{id: id, username: id, team: id, world: world, tangible: true, actions: createDefaultActions(), updates: make(chan []byte, 10000), camera: newCamera(make(chan []byte, 1000))}
	p.health.Store(125)
	p.killstreak.Store(4)
	p.tile = stage.tiles[y][x]
	stage.playerMap[id] = p
	p.tile.characterMap[id] = p
	return p
}
func moveFixturePlayer(p *Player, y, x int) {
	old := p.getTileSync()
	old.CharacterMutex.Lock()
	delete(old.characterMap, p.id)
	old.CharacterMutex.Unlock()
	p.tileLock.Lock()
	p.tile = old.stage.tiles[y][x]
	p.tile.CharacterMutex.Lock()
	p.tile.characterMap[p.id] = p
	p.tile.CharacterMutex.Unlock()
	p.tileLock.Unlock()
}

func TestIntervalSuppliesContinueForFifteenSimulatedMinutesWithoutEntry(t *testing.T) {
	world, areas := frontierFixture(t)
	stage := createStageFromArea(areas["willow-reach:0-0"])
	// Only supplies: deterministic wall clock injection exercises real interval,
	// collection, expiry and rotation code without waiting fifteen minutes.
	rules := []WorldSpawnRule{}
	for _, r := range world.config.manifest.Spawns {
		if r.Kind != "npc" {
			rules = append(rules, r)
		}
	}
	world.config.manifest.Spawns = rules
	p := fixturePlayer(world, stage, "blue", 24, 17)
	collected := 0
	now := time.Unix(1800000000, 0)
	for second := 0; second < 900; second++ {
		processPopulationStage(world, stage, now.Add(time.Duration(second)*time.Second), true)
		for _, r := range rules {
			if r.Stage != stage.name {
				continue
			}
			state := stage.ruleState[r.ID]
			if state == nil {
				continue
			}
			if len(state.pickups) > r.MaxAlive {
				t.Fatal("pickup capacity exceeded")
			}
			for tile := range state.pickups {
				if tile.collectItemsForPlayer(p) {
					collected++
				}
			}
		}
		// Drain the test player's UI queue just as a connected sender does.
		for len(p.updates) > 0 {
			<-p.updates
		}
	}
	if collected < 100 {
		t.Fatalf("supplies did not sustain play: %d collections", collected)
	}
	if p.health.Load() != 125 || p.killstreak.Load() != 4 {
		t.Fatal("supplies changed health/streak")
	}
}

func TestSpawnExcludesPlayersBlocksPortalsAndRespectsRegion(t *testing.T) {
	stage := makeSpectatorTestStage(12, 12)
	world := &World{}
	p := fixturePlayer(world, stage, "blue", 5, 5)
	stage.tiles[3][3].interactable = &Interactable{}
	stage.tiles[3][4].teleport = &Teleport{}
	rule := WorldSpawnRule{Kind: "npc", RandomPosition: true, Region: &SpawnRegion{X: 3, Y: 3, Width: 5, Height: 5}, MinPlayerDistance: 3}
	for _, tile := range populationCandidates(stage, rule) {
		if tile.x < 3 || tile.x > 7 || tile.y < 3 || tile.y > 7 || tile == p.tile || tile.interactable != nil || tile.teleport != nil || absInt(tile.x-5)+absInt(tile.y-5) < 3 {
			t.Fatal("unsafe candidate", tile.y, tile.x)
		}
	}
	stage.safe = true
	if len(populationCandidates(stage, rule)) != 0 {
		t.Fatal("NPC spawn accepted safe stage")
	}
}

func TestPuzzleSoloDuoVacancyAndResetPreserveCombatState(t *testing.T) {
	world, areas := frontierFixture(t)
	solo := createStageFromArea(areas["willow-solo:0-0"])
	p := fixturePlayer(world, solo, "blue", 12, 8)
	updatePuzzle(world, solo)
	if walkable(solo.tiles[6][8]) {
		t.Fatal("closed puzzle gate is walkable")
	}
	moveFixturePlayer(p, 12, 5)
	moveNorth(p)
	moveNorth(p)
	updatePuzzle(world, solo)
	if !walkable(solo.tiles[6][8]) {
		t.Fatal("block on socket did not open gate")
	}
	moveFixturePlayer(p, 4, 8)
	updatePuzzle(world, solo)
	if !solo.puzzleState.completed {
		t.Fatal("solo completion not recognized")
	}
	requestPuzzleReset(p)
	if solo.puzzleState.completed || solo.tiles[11][5].interactable == nil || walkable(solo.tiles[6][8]) {
		t.Fatal("reset did not restore initial state")
	}
	if p.health.Load() != 125 || p.killstreak.Load() != 4 || p.actions.spaceStack.count() != 0 {
		t.Fatal("puzzle changed combat state")
	}
	duo := createStageFromArea(areas["willow-duo:0-0"])
	a := fixturePlayer(world, duo, "a", 9, 5)
	b := fixturePlayer(world, duo, "b", 9, 11)
	updatePuzzle(world, duo)
	if !duo.puzzleState.completed || !walkable(duo.tiles[6][8]) {
		t.Fatal("two players should complete without a third player")
	}
	requestPuzzleReset(a)
	if !duo.puzzleState.completed {
		t.Fatal("one player reset another participant")
	}
	requestPuzzleReset(b)
	if duo.puzzleState.completed {
		t.Fatal("unanimous reset failed")
	}
	moveFixturePlayer(b, 12, 8)
	updatePuzzle(world, duo)
	if walkable(duo.tiles[6][8]) {
		t.Fatal("gate stayed open after plate released")
	}
	duo.playerMap = map[string]*Player{}
	updatePuzzle(world, duo)
	if duo.puzzleState.completed || len(duo.puzzleState.votes) > 0 {
		t.Fatal("empty chamber kept session state")
	}
}

func TestSafeRoomsPreventBothDamageAndPowerConsumption(t *testing.T) {
	stage := makeSpectatorTestStage(8, 8)
	stage.safe = true
	world := &World{App: App{config: &Configuration{}}}
	a := fixturePlayer(world, stage, "blue", 2, 2)
	b := fixturePlayer(world, stage, "pink", 2, 3)
	a.actions.spaceStack.push(&PowerUp{areaOfInfluence: grid3x3})
	a.activatePower()
	if a.actions.spaceStack.count() != 1 {
		t.Fatal("safe room consumed power")
	}
	b.takeDamageFrom(a, 50)
	if b.health.Load() != 125 || b.killstreak.Load() != 4 {
		t.Fatal("safe room allowed damage")
	}
}

func TestInvalidPopulationAndPuzzleRulesAreRejected(t *testing.T) {
	world, areas := frontierFixture(t)
	world.config.manifest.Spawns[0].Region = &SpawnRegion{X: 999, Y: 0, Width: 8, Height: 8}
	if validateWorldRules(world.config.manifest, areas) == nil {
		t.Fatal("invalid spawn region accepted")
	}
	world.config.manifest.Spawns = nil
	world.config.manifest.Puzzles[0].Stage = "willow-reach:0-0"
	if validateWorldRules(world.config.manifest, areas) == nil {
		t.Fatal("unsafe puzzle accepted")
	}
	if validateClassTokens("scenery-unbounded") == nil {
		t.Fatal("unknown scenery accepted")
	}
}

func TestPuzzleTurnsDoNotRotateASolvedBlock(t *testing.T) {
	world, areas := frontierFixture(t)
	stage := createStageFromArea(areas["willow-solo:0-0"])
	p := fixturePlayer(world, stage, "blue", 12, 5)
	p.handlePress(&PlayerSocketEvent{Name: "w"}, "")
	p.handlePress(&PlayerSocketEvent{Name: "w"}, "w")
	updatePuzzle(world, stage)
	p.handlePress(&PlayerSocketEvent{Name: "d"}, "w")
	if stage.tiles[9][5].interactable == nil {
		t.Fatal("turning moved the solved block off its socket")
	}
	for _, key := range []string{"d", "d", "w", "w", "w", "w", "w", "w"} {
		p.handlePress(&PlayerSocketEvent{Name: key}, "d")
		updatePuzzle(world, stage)
	}
	if !stage.puzzleState.completed {
		t.Fatal("ordinary movement could not pass the open gate and reach the goal")
	}
}

func TestPickupExpiryDoesNotDeleteUnrelatedDroppedMoney(t *testing.T) {
	stage := makeSpectatorTestStage(8, 8)
	rule := WorldSpawnRule{ID: "supply", Stage: stage.name, Kind: "powerup", Trigger: "interval", Count: 1, MaxAlive: 1, CooldownSeconds: 60, LifetimeSeconds: 10, Y: 2, X: 2}
	world := &World{App: App{config: &Configuration{manifest: WorldManifest{Spawns: []WorldSpawnRule{rule}}}}}
	now := time.Now()
	processPopulationStage(world, stage, now, true)
	tile := stage.tiles[2][2]
	tile.itemMutex.Lock()
	tile.money = 17
	tile.itemMutex.Unlock()
	processPopulationStage(world, stage, now.Add(11*time.Second), true)
	if tile.powerUp != nil || tile.money != 17 {
		t.Fatal("pickup expiry erased unrelated drops")
	}
}

func TestPuzzleBoostCannotJumpClosedGateOrConsumeCharges(t *testing.T) {
	world, areas := frontierFixture(t)
	stage := createStageFromArea(areas["willow-solo:0-0"])
	p := fixturePlayer(world, stage, "blue", 7, 8)
	p.addBoosts(3)
	updatePuzzle(world, stage)
	p.moveBoost(-1, 0)
	if p.getTileSync() != stage.tiles[7][8] || p.getBoostCountSync() != 3 {
		t.Fatal("boost bypassed a puzzle gate or consumed a charge")
	}
}

func TestBriefArrivalProtectionEndsOnAttack(t *testing.T) {
	stage := makeSpectatorTestStage(8, 8)
	world := &World{App: App{config: &Configuration{}}}
	a := fixturePlayer(world, stage, "blue", 3, 3)
	b := fixturePlayer(world, stage, "pink", 3, 4)
	b.arrivalProtection.Store(time.Now().Add(time.Second).UnixNano())
	b.takeDamageFrom(a, 25)
	if b.health.Load() != 125 {
		t.Fatal("arrival protection did not block damage")
	}
	b.actions.spaceStack.push(&PowerUp{areaOfInfluence: grid3x3})
	b.activatePower()
	if b.arrivalProtection.Load() != 0 {
		t.Fatal("attacking did not release protection")
	}
	b.takeDamageFrom(a, 25)
	if b.health.Load() != 100 {
		t.Fatal("ordinary rival damage failed after protection ended")
	}
}

func TestRecurringClockStopsOnCancellationAndSkipsEmptyStages(t *testing.T) {
	world, areas := frontierFixture(t)
	world.worldStages["willow-reach:0-0"] = createStageFromArea(areas["willow-reach:0-0"])
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { runWorldPopulation(world, ctx); close(done) }()
	time.Sleep(550 * time.Millisecond)
	if len(world.worldStages["willow-reach:0-0"].ruleState) > 0 {
		t.Fatal("empty stage spawned supplies")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("population clock did not stop")
	}
}

func TestPublishedFrontierPreservesRecurringRulesAndPuzzleContract(t *testing.T) {
	world, _ := frontierFixture(t)
	collection, err := loadSeedCollection("../../tools/main/data/collections", "bloop")
	if err != nil {
		t.Fatal(err)
	}
	palette, err := os.ReadFile("../../tools/main/data/colors/colors.json")
	if err != nil {
		t.Fatal(err)
	}
	resources := []WorldResource{{Kind: "collection", Key: "source", Content: mustJSON(t, collection)}, {Kind: "manifest", Key: "world", Content: mustJSON(t, world.config.manifest)}, {Kind: "palette", Key: "colors", Content: palette}}
	release, err := compileWorldResources(resources, true)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if err = extractReleaseArchive(release.Artifact, directory); err != nil {
		t.Fatal(err)
	}
	loaded := loadWorldManifest(directory)
	if len(loaded.Puzzles) != 2 || len(loaded.Spawns) != 21 || !loaded.Puzzles[1].PlayersOnly {
		t.Fatal("published release lost chamber or population data")
	}
	for _, r := range loaded.Spawns {
		if r.Trigger != "interval" {
			t.Fatal("recurring trigger lost")
		}
		if r.Kind == "npc" && (r.Region == nil || r.MinPlayerDistance < 5) {
			t.Fatal("spawn safety fields lost")
		}
	}
}

func TestRivalDefeatRespawnsWithoutHealingAttackerOrHurtingAlly(t *testing.T) {
	fixture, areas := frontierFixture(t)
	config := &Configuration{manifest: fixture.config.manifest}
	config.manifest.Spawns = nil
	world := createGameWorld(nil, config)
	world.stopPopulation()
	world.areas = nil
	for _, area := range areas {
		world.areas = append(world.areas, area)
	}
	join := func(name, team string, health int64, x int) *Player {
		p := world.join(createLoginRequest(PlayerRecord{Username: name, Team: team, Health: health, StageName: "willow-reach:0-0", Y: 24, X: x}), &MockConn{})
		if p == nil {
			t.Fatal("could not join review player")
		}
		return p
	}
	a := join("Blue", "sky-blue", 100, 17)
	b := join("Pink", "fuchsia", 50, 18)
	ally := join("Ally", "sky-blue", 50, 19)
	defer func() {
		for _, p := range []*Player{a, b, ally} {
			initiateLogout(p)
		}
	}()
	if ally.takeDamageFrom(a, 50) || ally.health.Load() != 50 {
		t.Fatal("friendly damage accepted")
	}
	if !b.takeDamageFrom(a, 50) {
		t.Fatal("rival defeat not registered")
	}
	if a.killCount.Load() != 1 || a.killCountNpc.Load() != 0 || a.killstreak.Load() != 1 || a.health.Load() != 100 {
		t.Fatal("incorrect attribution or kill healing")
	}
	if b.deathCount.Load() != 1 || b.health.Load() != 150 || b.killstreak.Load() != 0 || b.getTileSync().stage.name != "willow-refuge:0-1" {
		t.Fatal("death did not return to own refuge with a fresh life")
	}
}
