package main

import (
	"context"
	"fmt"
	"html"
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
	Label             string       `json:"label,omitempty"`
	ID                string       `json:"id"`
	Stage             string       `json:"stage"`
	Kind              string       `json:"kind"`
	NPC               string       `json:"npc,omitempty"`
	Trigger           string       `json:"trigger"`
	Count             int          `json:"count"`
	MaxAlive          int          `json:"maxAlive"`
	CooldownSeconds   int          `json:"cooldownSeconds"`
	RandomPosition    bool         `json:"randomPosition"`
	Y                 int          `json:"y"`
	X                 int          `json:"x"`
	Amount            int          `json:"amount,omitempty"`
	Region            *SpawnRegion `json:"region,omitempty"`
	MinPlayerDistance int          `json:"minPlayerDistance,omitempty"`
	LifetimeSeconds   int          `json:"lifetimeSeconds,omitempty"`
	RotationGroup     string       `json:"rotationGroup,omitempty"`
	ActiveSeconds     int          `json:"activeSeconds,omitempty"`
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
	rotationPeriods := map[string]int{}
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
		if rule.Trigger != "once" && rule.Trigger != "entry" && rule.Trigger != "interval" {
			return fmt.Errorf("spawn %q: trigger must be once, entry or interval", rule.ID)
		}
		if rule.Count < 1 || rule.Count > 20 || rule.MaxAlive < 1 || rule.MaxAlive > 50 || rule.CooldownSeconds < 1 || rule.CooldownSeconds > 86400 || rule.Amount < 0 || rule.Amount > 10000 {
			return fmt.Errorf("spawn %q: invalid count, capacity, cooldown or amount", rule.ID)
		}
		if rule.Kind == "npc" {
			totalNPCs += rule.MaxAlive
		}
		if len(rule.Label) > 80 {
			return fmt.Errorf("spawn label at most 80 characters")
		}
		if rule.MinPlayerDistance < 0 || rule.MinPlayerDistance > 32 || rule.LifetimeSeconds < 0 || rule.LifetimeSeconds > 3600 {
			return fmt.Errorf("spawn %q: invalid distance or lifetime", rule.ID)
		}
		if rule.RotationGroup != "" && (!paletteNamePattern.MatchString(rule.RotationGroup) || rule.ActiveSeconds < 10 || rule.ActiveSeconds > 3600 || rule.Trigger != "interval") {
			return fmt.Errorf("spawn %q: rotation requires interval trigger and active seconds 10–3600", rule.ID)
		}
		if rule.RotationGroup != "" {
			key := rule.Stage + "/" + rule.RotationGroup
			if duration, ok := rotationPeriods[key]; ok && duration != rule.ActiveSeconds {
				return fmt.Errorf("rotation group must use one active duration")
			}
			rotationPeriods[key] = rule.ActiveSeconds
		}
		if r := rule.Region; r != nil {
			if r.X < 0 || r.Y < 0 || r.Width < 1 || r.Height < 1 || r.Y+r.Height > len(area.Tiles) || r.X+r.Width > len(area.Tiles[0]) {
				return fmt.Errorf("spawn %q: invalid region", rule.ID)
			}
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
	return validatePuzzles(manifest, areas)
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

type SpawnRegion struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type spawnRuleState struct {
	last    time.Time
	npcs    []*NonPlayer
	pickups map[*Tile]time.Time
	active  bool
}

func spawnWorldRules(player *Player, stage *Stage) {
	if player.world == nil || player.world.config == nil {
		return
	}
	processPopulationStage(player.world, stage, time.Now(), false)
}

// A single world-owned clock services all occupied stages. It never depends on
// players crossing boundaries and never creates a goroutine per entry/rule.
func runWorldPopulation(world *World, ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			world.wPlayerMutex.Lock()
			players := make([]*Player, 0, len(world.worldPlayers))
			for _, p := range world.worldPlayers {
				players = append(players, p)
			}
			world.wPlayerMutex.Unlock()
			stages := map[*Stage]bool{}
			for _, p := range players {
				if tile := p.getTileSync(); tile != nil {
					stages[tile.stage] = true
				}
			}
			for stage := range stages {
				processPopulationStage(world, stage, now, true)
				updatePuzzle(world, stage)
			}
			world.wStageMutex.Lock()
			emptyPuzzles := []*Stage{}
			for _, s := range world.worldStages {
				if !stages[s] && puzzleFor(world, s) != nil {
					emptyPuzzles = append(emptyPuzzles, s)
				}
			}
			world.wStageMutex.Unlock()
			for _, s := range emptyPuzzles {
				updatePuzzle(world, s)
			}
		}
	}
}

func processPopulationStage(world *World, stage *Stage, now time.Time, interval bool) {
	stage.ruleMu.Lock()
	defer stage.ruleMu.Unlock()
	if stage.ruleState == nil {
		stage.ruleState = map[string]*spawnRuleState{}
	}
	manifest := &world.config.manifest
	for _, rule := range manifest.Spawns {
		if rule.Stage != stage.name || (interval && rule.Trigger != "interval") {
			continue
		}
		state := stage.ruleState[rule.ID]
		if state == nil {
			state = &spawnRuleState{pickups: map[*Tile]time.Time{}}
			stage.ruleState[rule.ID] = state
		}
		active := rotationRuleActive(manifest.Spawns, rule, now)
		if active && (!state.active || !interval) && rule.RotationGroup != "" && rule.Kind == "powerup" {
			stage.updateAll(soundTriggerByName("clink"))
			stage.updateAll(fmt.Sprintf(`<div id="frontier-event" role="status">Supply signal · %s</div>`, html.EscapeString(supplyLabel(rule))))
		}
		state.active = active
		alive := state.npcs[:0]
		for _, npc := range state.npcs {
			if npc.health.Load() > 0 {
				alive = append(alive, npc)
			}
		}
		state.npcs = alive
		for tile, born := range state.pickups {
			tile.itemMutex.Lock()
			owns := tile.spawnOwner == rule.ID
			present := (rule.Kind == "powerup" && tile.powerUp != nil) || (rule.Kind == "boost" && tile.boosts > 0) || (rule.Kind == "money" && tile.money > 0)
			expired := rule.LifetimeSeconds > 0 && now.Sub(born) >= time.Duration(rule.LifetimeSeconds)*time.Second
			if owns && present && (expired || !active) {
				switch rule.Kind {
				case "powerup":
					tile.powerUp = nil
				case "boost":
					tile.boosts = 0
				case "money":
					tile.money = 0
				}
				tile.spawnOwner = ""
			}
			tile.itemMutex.Unlock()
			if !owns || !present || expired || !active {
				delete(state.pickups, tile)
				if owns {
					tile.updateAll(svgFromTile(tile))
				}
			}
		}
		if !active || (!state.last.IsZero() && (rule.Trigger == "once" || now.Sub(state.last) < time.Duration(rule.CooldownSeconds)*time.Second)) {
			continue
		}
		state.last = now // failed attempts also back off; occupied sites cannot spin
		candidates := populationCandidates(stage, rule)
		for i := 0; i < rule.Count && len(candidates) > 0; i++ {
			index := rand.Intn(len(candidates))
			tile := candidates[index]
			candidates = append(candidates[:index], candidates[index+1:]...)
			if rule.Kind == "npc" {
				if len(state.npcs) >= rule.MaxAlive {
					break
				}
				for _, definition := range manifest.NPCs {
					if definition.ID == rule.NPC {
						if npc := spawnProgrammedNPC(world, tile, definition); npc != nil {
							state.npcs = append(state.npcs, npc)
						}
						break
					}
				}
			} else {
				if len(state.pickups) >= rule.MaxAlive {
					break
				}
				if insertRulePickup(tile, rule) {
					state.pickups[tile] = now
				}
			}
		}
	}
}

func rotationRuleActive(rules []WorldSpawnRule, rule WorldSpawnRule, now time.Time) bool {
	if rule.RotationGroup == "" {
		return true
	}
	group := []string{}
	for _, r := range rules {
		if r.Stage == rule.Stage && r.RotationGroup == rule.RotationGroup {
			group = append(group, r.ID)
		}
	}
	if len(group) == 0 || rule.ActiveSeconds < 1 {
		return false
	}
	return group[(now.Unix()/int64(rule.ActiveSeconds))%int64(len(group))] == rule.ID
}

func populationCandidates(stage *Stage, rule WorldSpawnRule) []*Tile {
	stage.playerMutex.RLock()
	players := make([]*Player, 0, len(stage.playerMap))
	for _, p := range stage.playerMap {
		players = append(players, p)
	}
	stage.playerMutex.RUnlock()
	locations := []*Tile{}
	for _, p := range players {
		if t := p.getTileSync(); t != nil && t.stage == stage {
			locations = append(locations, t)
		}
	}
	out := []*Tile{}
	for _, row := range stage.tiles {
		for _, tile := range row {
			if !tile.material.Walkable || tile.teleport != nil || (rule.Kind == "npc" && stage.safe) {
				continue
			}
			if !rule.RandomPosition && (tile.y != rule.Y || tile.x != rule.X) {
				continue
			}
			if r := rule.Region; r != nil && (tile.x < r.X || tile.y < r.Y || tile.x >= r.X+r.Width || tile.y >= r.Y+r.Height) {
				continue
			}
			if hasCharacters(tile) {
				continue
			}
			tile.interactableMutex.Lock()
			blocked := tile.interactable != nil
			tile.interactableMutex.Unlock()
			if blocked {
				continue
			}
			distant := true
			for _, p := range locations {
				if absInt(p.x-tile.x)+absInt(p.y-tile.y) < rule.MinPlayerDistance {
					distant = false
					break
				}
			}
			if distant {
				out = append(out, tile)
			}
		}
	}
	if len(locations) > 0 && rule.RandomPosition {
		reachable := map[*Tile]bool{}
		queue := append([]*Tile{}, locations...)
		for _, t := range queue {
			reachable[t] = true
		}
		for head := 0; head < len(queue); head++ {
			t := queue[head]
			for _, off := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
				y, x := t.y+off[0], t.x+off[1]
				if !validCoordinate(y, x, stage) {
					continue
				}
				next := stage.tiles[y][x]
				if !reachable[next] && next.material.Walkable {
					reachable[next] = true
					queue = append(queue, next)
				}
			}
		}
		connected := out[:0]
		for _, t := range out {
			if reachable[t] {
				connected = append(connected, t)
			}
		}
		out = connected
	}

	if rule.Kind == "npc" && len(locations) > 0 {
		nearby := []*Tile{}
		for _, t := range out {
			for _, p := range locations {
				if absInt(p.y-t.y)+absInt(p.x-t.x) <= max(6, rule.MinPlayerDistance+1) {
					nearby = append(nearby, t)
					break
				}
			}
		}
		if len(nearby) > 0 {
			return nearby
		}
	}
	return out
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func insertRulePickup(tile *Tile, rule WorldSpawnRule) bool {
	tile.itemMutex.Lock()
	if tile.powerUp != nil || tile.boosts > 0 || tile.money > 0 {
		tile.itemMutex.Unlock()
		return false
	}
	tile.spawnOwner = rule.ID
	switch rule.Kind {
	case "powerup":
		tile.powerUp = &PowerUp{areaOfInfluence: grid3x3}
	case "boost":
		tile.boosts = rule.Amount
		if tile.boosts < 1 {
			tile.boosts = 5
		}
	case "money":
		tile.money = rule.Amount
	}
	tile.itemMutex.Unlock()
	tile.updateAll(svgFromTile(tile))
	return true
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
			if ctx.Err() != nil || npc.health.Load() <= 0 || (npc.world.populationContext != nil && npc.world.populationContext.Err() != nil) {
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
		if tile == nil || tile.stage != origin.stage || safeFromDamage(tile) {
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

// Local review telemetry: no production route exposes this endpoint.
func populationStatus(world *World) map[string]any {
	world.wPlayerMutex.Lock()
	players := make([]*Player, 0, len(world.worldPlayers))
	for _, p := range world.worldPlayers {
		players = append(players, p)
	}
	world.wPlayerMutex.Unlock()
	playerRows := []map[string]any{}
	for _, p := range players {
		if t := p.getTileSync(); t != nil {
			playerRows = append(playerRows, map[string]any{"name": p.username, "team": p.getTeamNameSync(), "stage": t.stage.name, "y": t.y, "x": t.x, "health": p.health.Load(), "streak": p.killstreak.Load(), "powers": p.actions.spaceStack.count(), "boosts": p.getBoostCountSync()})
		}
	}
	world.wStageMutex.Lock()
	stages := []*Stage{}
	for _, s := range world.worldStages {
		stages = append(stages, s)
	}
	world.wStageMutex.Unlock()
	rules := []map[string]any{}
	for _, s := range stages {
		s.ruleMu.Lock()
		for _, r := range world.config.manifest.Spawns {
			if r.Stage != s.name {
				continue
			}
			if state := s.ruleState[r.ID]; state != nil {
				rules = append(rules, map[string]any{"id": r.ID, "stage": s.name, "active": state.active, "npcs": len(state.npcs), "pickups": len(state.pickups), "capacity": r.MaxAlive, "nextSeconds": max(0, int(time.Until(state.last.Add(time.Duration(r.CooldownSeconds)*time.Second)).Seconds()))})
			}
		}
		s.ruleMu.Unlock()
	}
	return map[string]any{"players": playerRows, "rules": rules}
}

func supplyLabel(rule WorldSpawnRule) string {
	if rule.Label != "" {
		return rule.Label
	}
	return strings.ReplaceAll(strings.TrimSuffix(strings.TrimSuffix(rule.ID, "-power"), "-boost"), "-", " ") + " · Powers + Boosts"
}
