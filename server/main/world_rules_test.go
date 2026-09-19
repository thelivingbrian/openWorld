package main

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

func rulesFixture() (WorldManifest, map[string]Area) {
	manifest := WorldManifest{Name: "Rules", Entry: WorldLocation{Stage: "room"}, MaxPlayers: 20, Lifecycle: LifecyclePersistent,
		NPCs: []NPCDefinition{{ID: "guard", Name: "Guard", Team: "npc", Appearance: "red-b thick r0", Health: 100, Money: 20, IntervalMS: 10000, Lifetime: 60,
			Program: []NPCStep{{Action: "east", Ticks: 2, Condition: "always"}, {Action: "attack", Ticks: 1, Condition: "player-nearby", Radius: 2, Damage: 10}}}},
		Spawns:       []WorldSpawnRule{{ID: "guards", Stage: "room", Kind: "npc", NPC: "guard", Trigger: "entry", Count: 2, MaxAlive: 3, CooldownSeconds: 30, X: 1, Y: 1}},
		Achievements: []WorldAchievement{{ID: "explore", Name: "Explorer", Metric: "visit-stage", Stage: "room", Target: 1}, {ID: "hunter", Name: "Hunter", Metric: "killCountNpc", Target: 2}},
	}
	area := Area{Name: "room", SpawnStrategy: "none", Tiles: [][]Material{{{Walkable: true}, {Walkable: true}, {Walkable: true}}, {{Walkable: true}, {Walkable: true}, {Walkable: true}}, {{Walkable: true}, {Walkable: true}, {Walkable: true}}}}
	return manifest, map[string]Area{"room": area}
}

func TestWorldRuleValidation(t *testing.T) {
	mutations := map[string]func(*WorldManifest){
		"unknown NPC":           func(m *WorldManifest) { m.Spawns[0].NPC = "missing" },
		"unknown stage":         func(m *WorldManifest) { m.Spawns[0].Stage = "missing" },
		"out of bounds":         func(m *WorldManifest) { m.Spawns[0].X = 99 },
		"unbounded interval":    func(m *WorldManifest) { m.NPCs[0].IntervalMS = 0 },
		"empty program":         func(m *WorldManifest) { m.NPCs[0].Program = nil },
		"unknown action":        func(m *WorldManifest) { m.NPCs[0].Program[0].Action = "eval" },
		"oversized spawn":       func(m *WorldManifest) { m.Spawns[0].Count = 10000 },
		"duplicate IDs":         func(m *WorldManifest) { m.NPCs = append(m.NPCs, m.NPCs[0]) },
		"unsafe achievement ID": func(m *WorldManifest) { m.Achievements[0].ID = "$field.name" },
		"unknown metric":        func(m *WorldManifest) { m.Achievements[0].Metric = "invalid" },
	}
	manifest, areas := rulesFixture()
	if err := validateWorldRules(manifest, areas); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			manifest, areas := rulesFixture()
			mutate(&manifest)
			if err := validateWorldRules(manifest, areas); err == nil {
				t.Fatal("invalid rules were accepted")
			}
		})
	}
}

func TestWorldSpawnCapacityCooldownAndNPCProgram(t *testing.T) {
	manifest, areas := rulesFixture()
	world := &World{App: App{config: &Configuration{manifest: manifest}}}
	player := &Player{world: world}
	stage := createStageFromArea(areas["room"])
	var group sync.WaitGroup
	for i := 0; i < 20; i++ {
		group.Add(1)
		go func() { defer group.Done(); spawnWorldRules(player, stage) }()
	}
	group.Wait()
	state := stage.ruleState["guards"]
	if len(state.npcs) != 2 {
		t.Fatalf("concurrent entry spawned %d NPCs, want 2", len(state.npcs))
	}
	state.last = time.Now().Add(-time.Minute)
	spawnWorldRules(player, stage)
	if len(state.npcs) != 3 {
		t.Fatalf("capacity not respected: %d", len(state.npcs))
	}
	t.Cleanup(func() {
		for _, npc := range state.npcs {
			npc.terminate()
		}
	})
	npc := state.npcs[0]
	runNPCStep(npc, 100, NPCStep{Action: "east", Condition: "hurt"})
	if npc.getTileSync().x != 1 {
		t.Fatal("conditional action ran while healthy")
	}
	runNPCStep(npc, 100, NPCStep{Action: "east", Condition: "always"})
	if npc.getTileSync().x != 2 {
		t.Fatal("NPC program did not move east")
	}
}

func TestWorldAchievementsAreStableAndEscaped(t *testing.T) {
	manifest, _ := rulesFixture()
	manifest.Achievements[0].Name = "<script>Explorer</script>"
	player := &Player{world: &World{App: App{config: &Configuration{manifest: manifest}}}}
	player.checkWorldAchievements("elsewhere")
	if len(player.accomplishments.Accomplishments) != 0 {
		t.Fatal("awarded an unmet achievement")
	}
	player.checkWorldAchievements("room")
	first := player.accomplishments.Accomplishments["world-explore"]
	player.checkWorldAchievements("room")
	if player.accomplishments.Accomplishments["world-explore"] != first {
		t.Fatal("awarded twice")
	}
	player.incrementKillCountNpc()
	player.incrementKillCountNpc()
	if len(player.accomplishments.Accomplishments) != 2 {
		t.Fatal("NPC defeat achievement did not trigger")
	}
	output := string(createAccomplishmentsHtmlForPlayer(player))
	if strings.Contains(output, "<script>") || !strings.Contains(output, "&lt;script&gt;") {
		t.Fatal("achievement labels must be escaped")
	}
}

func TestWorldSpawnUsesTeamAndEntryLocations(t *testing.T) {
	manifest, _ := rulesFixture()
	manifest.Teams = []WorldTeam{{ID: "blue", Spawn: WorldLocation{Stage: "base", Y: 4, X: 5}}}
	player := &Player{team: "blue", world: &World{App: App{config: &Configuration{manifest: manifest}}}}
	if infirmaryStagenameForPlayer(player) != "base" {
		t.Fatal("team spawn ignored")
	}
	y, x := infirmaryCoordsForPlayer(player)
	if y != 4 || x != 5 {
		t.Fatal("team spawn coordinates ignored")
	}
	player.team = "unknown"
	if infirmaryStagenameForPlayer(player) != "room" {
		t.Fatal("entry fallback ignored")
	}
}

func TestWorldAdmissionRespectsPlayerLimitConcurrently(t *testing.T) {
	world := &World{App: App{config: &Configuration{manifest: WorldManifest{MaxPlayers: 2}}}, worldPlayers: map[string]*Player{}, teamQuantities: map[string]int{}}
	var group sync.WaitGroup
	for _, id := range []string{"a", "b", "c", "d", "e"} {
		group.Add(1)
		go func(id string) { defer group.Done(); world.admitPlayer(&Player{id: id, team: "blue"}) }(id)
	}
	group.Wait()
	if len(world.worldPlayers) != 2 {
		t.Fatalf("admitted %d players, want 2", len(world.worldPlayers))
	}
}

func TestCompiledWorldPreservesRulesAndRejectsInvalidEntry(t *testing.T) {
	resources := testReleaseResources(t)
	var manifest WorldManifest
	if err := json.Unmarshal(resources[1].Content, &manifest); err != nil {
		t.Fatal(err)
	}
	rules, _ := rulesFixture()
	manifest.NPCs = rules.NPCs
	manifest.Spawns = rules.Spawns
	manifest.Spawns[0].Stage = manifest.Entry.Stage
	manifest.Achievements = rules.Achievements
	manifest.Achievements[0].Stage = manifest.Entry.Stage
	resources[1].Content = mustJSON(t, manifest)
	compiled, err := compileWorldResources(resources, true)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if err := extractReleaseArchive(compiled.Artifact, directory); err != nil {
		t.Fatal(err)
	}
	loaded := loadWorldManifest(directory)
	if len(loaded.NPCs) != 1 || len(loaded.Spawns) != 1 || len(loaded.Achievements) != 2 {
		t.Fatal("release lost world rules")
	}
	for i := range resources {
		if resources[i].Kind == "manifest" {
			manifest.Entry.X = 999
			resources[i].Content = mustJSON(t, manifest)
		}
	}
	if _, err := compileWorldResources(resources, true); err == nil {
		t.Fatal("out-of-bounds player entry accepted")
	}
}
