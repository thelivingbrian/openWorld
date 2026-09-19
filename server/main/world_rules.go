package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Programs are bounded, declarative action sequences. Each NPC has its own
// program counter; no user code executes in the game server.
type NPCDefinition struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Team       string    `json:"team"`
	Appearance string    `json:"appearance"`
	Health     int       `json:"health"`
	Money      int       `json:"money"`
	IntervalMS int       `json:"intervalMs"`
	Lifetime   int       `json:"lifetimeSeconds"`
	Program    []NPCStep `json:"program"`
}

type NPCStep struct {
	Action    string `json:"action"`
	Ticks     int    `json:"ticks"`
	Condition string `json:"condition"`
	Radius    int    `json:"radius,omitempty"`
	Damage    int    `json:"damage,omitempty"`
}

type WorldSpawnRule struct {
	ID              string `json:"id"`
	Stage           string `json:"stage"`
	Kind            string `json:"kind"`
	NPC             string `json:"npc,omitempty"`
	Trigger         string `json:"trigger"`
	Count           int    `json:"count"`
	MaxAlive        int    `json:"maxAlive"`
	CooldownSeconds int    `json:"cooldownSeconds"`
	RandomPosition  bool   `json:"randomPosition"`
	Y               int    `json:"y"`
	X               int    `json:"x"`
	Amount          int    `json:"amount,omitempty"`
}

type WorldAchievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Metric      string `json:"metric"`
	Target      int64  `json:"target"`
	Stage       string `json:"stage,omitempty"`
}

func validateWorldRules(manifest WorldManifest, areas map[string]Area) error {
	if len(manifest.NPCs) > 64 || len(manifest.Spawns) > 256 || len(manifest.Achievements) > 128 {
		return fmt.Errorf("world may contain at most 64 NPC types, 256 spawn rules and 128 achievements")
	}
	checkID := func(id string, seen map[string]bool) error {
		if !paletteNamePattern.MatchString(id) || seen[id] {
			return fmt.Errorf("invalid or duplicate ID %q (use lowercase letters, digits and hyphens)", id)
		}
		seen[id] = true
		return nil
	}
	npcs := map[string]bool{}
	for _, npc := range manifest.NPCs {
		if err := checkID(npc.ID, npcs); err != nil {
			return err
		}
		if strings.TrimSpace(npc.Name) == "" || len(npc.Name) > 80 || len(npc.Team) > 80 || npc.Team == "" {
			return fmt.Errorf("NPC %q needs a name and team of at most 80 characters", npc.ID)
		}
		if err := validateClassTokens(npc.Appearance); err != nil {
			return err
		}
		if npc.Health < 1 || npc.Health > 10000 || npc.Money < 0 || npc.Money > 100000 || npc.IntervalMS < 100 || npc.IntervalMS > 10000 || npc.Lifetime < 1 || npc.Lifetime > 3600 {
			return fmt.Errorf("NPC %q: health 1–10000, money 0–100000, interval 100–10000ms and lifetime 1–3600s required", npc.ID)
		}
		if len(npc.Program) == 0 || len(npc.Program) > 64 {
			return fmt.Errorf("NPC %q needs 1–64 program steps", npc.ID)
		}
		for _, step := range npc.Program {
			switch step.Action {
			case "wait", "north", "south", "east", "west", "wander", "chase", "attack":
			default:
				return fmt.Errorf("NPC %q: unknown action %q", npc.ID, step.Action)
			}
			if step.Condition != "always" && step.Condition != "hurt" && step.Condition != "player-nearby" {
				return fmt.Errorf("NPC %q: unknown condition %q", npc.ID, step.Condition)
			}
			if step.Ticks < 1 || step.Ticks > 100 || step.Radius < 0 || step.Radius > 8 || step.Damage < 0 || step.Damage > 1000 {
				return fmt.Errorf("NPC %q: invalid step limits", npc.ID)
			}
			if step.Action == "attack" && (step.Radius < 1 || step.Damage < 1) {
				return fmt.Errorf("NPC %q: attack requires radius and damage", npc.ID)
			}
		}
	}
	spawns := map[string]bool{}
	totalNPCs := 0
	for _, rule := range manifest.Spawns {
		if err := checkID(rule.ID, spawns); err != nil {
			return err
		}
		area, ok := areas[rule.Stage]
		if !ok {
			return fmt.Errorf("spawn %q: unknown stage %q", rule.ID, rule.Stage)
		}
		if rule.Kind != "npc" && rule.Kind != "money" && rule.Kind != "boost" && rule.Kind != "powerup" {
			return fmt.Errorf("spawn %q: unknown kind", rule.ID)
		}
		if rule.Kind == "npc" && !npcs[rule.NPC] {
			return fmt.Errorf("spawn %q: unknown NPC %q", rule.ID, rule.NPC)
		}
		if rule.Trigger != "once" && rule.Trigger != "entry" {
			return fmt.Errorf("spawn %q: trigger must be once or entry", rule.ID)
		}
		if rule.Count < 1 || rule.Count > 20 || rule.MaxAlive < 1 || rule.MaxAlive > 50 || rule.CooldownSeconds < 1 || rule.CooldownSeconds > 86400 || rule.Amount < 0 || rule.Amount > 10000 {
			return fmt.Errorf("spawn %q: invalid count, capacity, cooldown or amount", rule.ID)
		}
		if rule.Kind == "npc" {
			totalNPCs += rule.MaxAlive
		}
		if !rule.RandomPosition {
			if err := validateWorldLocation(WorldLocation{Stage: rule.Stage, Y: rule.Y, X: rule.X}, areas); err != nil {
				return fmt.Errorf("spawn %q: %w", rule.ID, err)
			}
		} else {
			walkable := false
			for _, row := range area.Tiles {
				for _, tile := range row {
					walkable = walkable || tile.Walkable
				}
			}
			if !walkable {
				return fmt.Errorf("spawn %q: stage has no walkable tiles", rule.ID)
			}
		}
	}
	if totalNPCs > 500 {
		return fmt.Errorf("combined NPC spawn capacity exceeds 500")
	}
	achievements := map[string]bool{}
	for _, achievement := range manifest.Achievements {
		if err := checkID(achievement.ID, achievements); err != nil {
			return err
		}
		if strings.TrimSpace(achievement.Name) == "" || len(achievement.Name) > 80 || len(achievement.Description) > 500 {
			return fmt.Errorf("achievement %q: name required (80 characters), description at most 500 characters", achievement.ID)
		}
		switch achievement.Metric {
		case "visit-stage":
			if _, ok := areas[achievement.Stage]; !ok {
				return fmt.Errorf("achievement %q: unknown stage", achievement.ID)
			}
		case "peakWealth", "peakKillStreak", "killCount", "killCountNpc", "goalsScored", "deathCount":
			if achievement.Target < 1 || achievement.Target > 1000000000 {
				return fmt.Errorf("achievement %q: target must be 1–1000000000", achievement.ID)
			}
		default:
			return fmt.Errorf("achievement %q: unknown metric", achievement.ID)
		}
	}
	return nil
}

func validateWorldLocation(location WorldLocation, areas map[string]Area) error {
	area, ok := areas[location.Stage]
	if !ok || location.Y < 0 || location.Y >= len(area.Tiles) || location.X < 0 || location.X >= len(area.Tiles[location.Y]) {
		return fmt.Errorf("invalid location %s (%d, %d)", location.Stage, location.X, location.Y)
	}
	if !area.Tiles[location.Y][location.X].Walkable {
		return fmt.Errorf("location %s (%d, %d) is not walkable", location.Stage, location.X, location.Y)
	}
	return nil
}

type spawnRuleState struct {
	last time.Time
	npcs []*NonPlayer
}

func spawnWorldRules(player *Player, stage *Stage) {
	if player.world == nil || player.world.config == nil {
		return
	}
	manifest := &player.world.config.manifest
	stage.ruleMu.Lock()
	defer stage.ruleMu.Unlock()
	if stage.ruleState == nil {
		stage.ruleState = map[string]*spawnRuleState{}
	}
	for _, rule := range manifest.Spawns {
		if rule.Stage != stage.name {
			continue
		}
		state := stage.ruleState[rule.ID]
		if state == nil {
			state = &spawnRuleState{}
			stage.ruleState[rule.ID] = state
		}
		if !state.last.IsZero() && (rule.Trigger == "once" || time.Since(state.last) < time.Duration(rule.CooldownSeconds)*time.Second) {
			continue
		}
		alive := state.npcs[:0]
		for _, npc := range state.npcs {
			if npc.health.Load() > 0 {
				alive = append(alive, npc)
			}
		}
		state.npcs = alive
		tiles := walkableTiles(stage.tiles)
		if len(tiles) == 0 {
			continue
		}
		spawned := false
		for i := 0; i < rule.Count; i++ {
			tile := tiles[rand.Intn(len(tiles))]
			if !rule.RandomPosition {
				if !validCoordinate(rule.Y, rule.X, stage) {
					break
				}
				tile = stage.tiles[rule.Y][rule.X]
			}
			switch rule.Kind {
			case "npc":
				if len(state.npcs) >= rule.MaxAlive {
					continue
				}
				for _, definition := range manifest.NPCs {
					if definition.ID == rule.NPC {
						if npc := spawnProgrammedNPC(player.world, tile, definition); npc != nil {
							state.npcs = append(state.npcs, npc)
							spawned = true
						}
						break
					}
				}
			case "money":
				tile.addMoneyAndNotifyAll(rule.Amount)
				spawned = true
			case "boost":
				tile.addBoostsAndNotifyAll()
				spawned = true
			case "powerup":
				tile.addPowerUpAndNotifyAll(shapesNpc[0])
				spawned = true
			}
		}
		if spawned {
			state.last = time.Now()
		}
	}
}

func spawnProgrammedNPC(world *World, tile *Tile, definition NPCDefinition) *NonPlayer {
	// Also guard local manifests, which can be authored outside the publisher.
	if definition.IntervalMS < 100 || definition.Lifetime < 1 || len(definition.Program) == 0 {
		return nil
	}
	for {
		count := world.programmedNPCCount.Load()
		if count >= 500 {
			return nil
		}
		if world.programmedNPCCount.CompareAndSwap(count, count+1) {
			break
		}
	}
	npc, ctx := createNewNPC(world, definition.Team)
	npc.icon, npc.iconLow = definition.Appearance, definition.Appearance
	npc.health.Store(int64(definition.Health))
	npc.money.Store(int64(definition.Money))
	addNPCAndNotifyOthers(npc, tile)
	go runNPCProgram(ctx, npc, definition)
	return npc
}

func runNPCProgram(ctx context.Context, npc *NonPlayer, definition NPCDefinition) {
	defer npc.world.programmedNPCCount.Add(-1)
	ticker := time.NewTicker(time.Duration(definition.IntervalMS) * time.Millisecond)
	expires := time.NewTimer(time.Duration(definition.Lifetime) * time.Second)
	defer ticker.Stop()
	defer expires.Stop()
	defer npc.terminate()
	defer removeNpcFromTile(npc)
	defer npc.health.Store(0)
	index, ticks := 0, 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-expires.C:
			return
		case <-ticker.C:
			if ctx.Err() != nil || npc.health.Load() <= 0 {
				return
			}
			step := definition.Program[index]
			runNPCStep(npc, definition.Health, step)
			ticks++
			if ticks >= step.Ticks {
				ticks = 0
				index = (index + 1) % len(definition.Program)
			}
		}
	}
}

func runNPCStep(npc *NonPlayer, initialHealth int, step NPCStep) {
	if step.Condition == "hurt" && npc.health.Load()*2 >= int64(initialHealth) {
		return
	}
	origin := npc.getTileSync()
	if origin == nil {
		return
	}
	target := nearestNPCPlayer(npc, origin, max(1, step.Radius))
	if step.Condition == "player-nearby" && target == nil {
		return
	}
	switch step.Action {
	case "north":
		moveNorth(npc)
	case "south":
		moveSouth(npc)
	case "east":
		moveEast(npc)
	case "west":
		moveWest(npc)
	case "wander":
		moveRandomly(npc)
	case "chase":
		if target == nil {
			return
		}
		if target.y < origin.y {
			moveNorth(npc)
		} else if target.y > origin.y {
			moveSouth(npc)
		} else if target.x < origin.x {
			moveWest(npc)
		} else if target.x > origin.x {
			moveEast(npc)
		}
	case "attack":
		tiles := getOrderedRegion(origin.stage, origin.y-step.Radius, origin.x-step.Radius, 2*step.Radius+1, 2*step.Radius+1)
		damageAndIndicate(tiles, npc, step.Damage)
	}
}

func nearestNPCPlayer(npc *NonPlayer, origin *Tile, radius int) *Tile {
	origin.stage.playerMutex.RLock()
	players := make([]*Player, 0, len(origin.stage.playerMap))
	for _, player := range origin.stage.playerMap {
		players = append(players, player)
	}
	origin.stage.playerMutex.RUnlock()
	var nearest *Tile
	distance := radius + 1
	for _, player := range players {
		if player.getTeamNameSync() == npc.getTeamNameSync() {
			continue
		}
		tile := player.getTileSync()
		if tile == nil || tile.stage != origin.stage {
			continue
		}
		dy, dx := tile.y-origin.y, tile.x-origin.x
		if dy < 0 {
			dy = -dy
		}
		if dx < 0 {
			dx = -dx
		}
		if dy+dx < distance {
			nearest, distance = tile, dy+dx
		}
	}
	return nearest
}

func (player *Player) checkWorldAchievements(stage string) {
	if player.world == nil || player.world.config == nil {
		return
	}
	for _, achievement := range player.world.config.manifest.Achievements {
		var value int64
		switch achievement.Metric {
		case "visit-stage":
			if stage == achievement.Stage {
				value = achievement.Target
			}
		case "peakWealth":
			value = player.peakWealth.Load()
		case "peakKillStreak":
			value = player.peakKillStreak.Load()
		case "killCount":
			value = player.killCount.Load()
		case "killCountNpc":
			value = player.killCountNpc.Load()
		case "goalsScored":
			value = player.goalsScored.Load()
		case "deathCount":
			value = player.deathCount.Load()
		}
		if achievement.Metric == "visit-stage" && stage != achievement.Stage {
			continue
		}
		if value < achievement.Target {
			continue
		}
		key := "world-" + achievement.ID
		acc := player.accomplishments.addByName(key)
		if acc != nil && player.world.db != nil {
			player.world.db.addAccomplishmentToPlayer(player.world.config.worldID, player.username, key, *acc)
		}
	}
}
