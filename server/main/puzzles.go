package main

import (
	"fmt"
	"strings"
)

type PuzzlePoint struct {
	Y int `json:"y"`
	X int `json:"x"`
}
type WorldPuzzle struct {
	ID          string        `json:"id"`
	Stage       string        `json:"stage"`
	Plates      []PuzzlePoint `json:"plates"`
	Goal        PuzzlePoint   `json:"goal"`
	Gates       []PuzzlePoint `json:"gates"`
	MinPlayers  int           `json:"minPlayers"`
	PlayersOnly bool          `json:"playersOnly,omitempty"`
}
type puzzleRuntime struct {
	original    map[*Tile]*Interactable
	votes       map[string]bool
	completed   bool
	active      bool
	wasOccupied bool
}

func validatePuzzles(manifest WorldManifest, areas map[string]Area) error {
	if len(manifest.Puzzles) > 32 {
		return fmt.Errorf("at most 32 puzzle chambers")
	}
	seen := map[string]bool{}
	ids := map[string]bool{}
	for _, p := range manifest.Puzzles {
		if !paletteNamePattern.MatchString(p.ID) || ids[p.ID] || seen[p.Stage] || p.MinPlayers < 1 || p.MinPlayers > 8 || len(p.Plates) < 1 || len(p.Plates) > 8 || len(p.Gates) > 16 {
			return fmt.Errorf("invalid puzzle %q", p.ID)
		}
		seen[p.Stage] = true
		ids[p.ID] = true
		area, ok := areas[p.Stage]
		if !ok || !area.Safe {
			return fmt.Errorf("puzzle %q requires a safe stage", p.ID)
		}
		for _, point := range append(append(append([]PuzzlePoint{}, p.Plates...), p.Gates...), p.Goal) {
			if point.Y < 0 || point.Y >= len(area.Tiles) || point.X < 0 || point.X >= len(area.Tiles[point.Y]) {
				return fmt.Errorf("puzzle %q point out of bounds", p.ID)
			}
		}
		for _, point := range p.Gates {
			if area.Tiles[point.Y][point.X].Walkable {
				return fmt.Errorf("puzzle gate must be blocked in the blueprint")
			}
		}
		for _, point := range append(append([]PuzzlePoint{}, p.Plates...), p.Goal) {
			if !area.Tiles[point.Y][point.X].Walkable {
				return fmt.Errorf("puzzle plates/goal must be walkable")
			}
		}
	}
	return nil
}

func puzzleFor(world *World, stage *Stage) *WorldPuzzle {
	if world == nil || world.config == nil {
		return nil
	}
	for i := range world.config.manifest.Puzzles {
		if world.config.manifest.Puzzles[i].Stage == stage.name {
			return &world.config.manifest.Puzzles[i]
		}
	}
	return nil
}

// Called outside movement/tile locks. Lock every interactable in canonical
// order so blocks and gates are evaluated/restored as a consistent snapshot.
func updatePuzzle(world *World, stage *Stage) {
	spec := puzzleFor(world, stage)
	if spec == nil {
		return
	}
	stage.puzzleMu.Lock()
	defer stage.puzzleMu.Unlock()
	state := stage.puzzleState
	if state == nil {
		state = &puzzleRuntime{original: map[*Tile]*Interactable{}, votes: map[string]bool{}}
		stage.puzzleState = state
	}
	tiles := []*Tile{}
	for _, row := range stage.tiles {
		for _, t := range row {
			if !t.interactableMutex.TryLock() {
				for _, held := range tiles {
					held.interactableMutex.Unlock()
				}
				return
			}
			tiles = append(tiles, t)
		}
	}
	defer func() {
		for _, t := range tiles {
			t.interactableMutex.Unlock()
		}
	}()
	if state.original == nil {
		state.original = map[*Tile]*Interactable{}
	}
	if !state.active {
		for _, t := range tiles {
			if t.interactable != nil {
				copy := *t.interactable
				state.original[t] = &copy
			}
		}
		state.active = true
	}
	players := stagePlayers(stage)
	current := map[string]*Tile{}
	for _, p := range players {
		if t := p.getTileSync(); t != nil && t.stage == stage {
			current[p.id] = t
		}
	}
	for id := range state.votes {
		if current[id] == nil {
			delete(state.votes, id)
		}
	}
	reset := (len(current) == 0 && state.wasOccupied) || (len(current) > 0 && len(state.votes) == len(current))
	state.wasOccupied = len(current) > 0
	if reset {
		// A block may have been pushed onto a player's tile since the vote. Wait
		// for that tile to clear rather than trap anyone when restoring the block.
		for t := range state.original {
			if hasCharacters(t) {
				return
			}
		}
		for _, t := range tiles {
			t.interactable = nil
			if original := state.original[t]; original != nil {
				copy := *original
				t.interactable = &copy
			}
			t.updateAll(interactableBoxLocked(t))
		}
		state.votes = map[string]bool{}
		state.completed = false
	}
	open := !reset && len(current) >= spec.MinPlayers
	for _, point := range spec.Plates {
		t := stage.tiles[point.Y][point.X]
		occupied := false
		for _, location := range current {
			if location == t {
				occupied = true
			}
		}
		if !spec.PlayersOnly && t.interactable != nil {
			occupied = true
		}
		open = open && occupied
		color := "lavender-b thick r0"
		if occupied {
			color = "turquoise-b thick r0"
		}
		t.updateAll(swapToken(t.y, t.x, "Lf1", "", color))
	}
	for _, point := range spec.Gates {
		t := stage.tiles[point.Y][point.X]
		mode := int32(1)
		if open || hasCharacters(t) {
			mode = 2
		}
		if t.gateMode.Swap(mode) != mode {
			t.updateAll(gateSwap(t))
		}
	}
	if open && !state.completed {
		goal := stage.tiles[spec.Goal.Y][spec.Goal.X]
		for _, location := range current {
			if location == goal {
				state.completed = true
				for _, p := range players {
					if p.tangibilityLock.TryLock() {
						if p.tangible {
							p.updateBottomText("Relay complete · Nicely done! No health or combat supplies changed.")
						}
						p.tangibilityLock.Unlock()
					}
				}
				goal.updateAll(soundTriggerByName("clink"))
				break
			}
		}
	}
}

func interactableBoxLocked(t *Tile) string {
	classes := ""
	if t.interactable != nil {
		classes = t.interactable.cssClass
	}
	return swapToken(t.y, t.x, "Li1", "", classes)
}
func gateSwap(t *Tile) string {
	if t.gateMode.Load() == 0 {
		return ""
	}
	classes := "lavender r0 turquoise-b med"
	if t.gateMode.Load() == 2 {
		classes = "turquoise-b thin r0"
	}
	return swapToken(t.y, t.x, "Lf1", "", classes)
}
func stagePlayers(stage *Stage) []*Player {
	stage.playerMutex.RLock()
	defer stage.playerMutex.RUnlock()
	players := make([]*Player, 0, len(stage.playerMap))
	for _, p := range stage.playerMap {
		players = append(players, p)
	}
	return players
}
func requestPuzzleReset(p *Player) {
	tile := p.getTileSync()
	if tile == nil || puzzleFor(p.world, tile.stage) == nil {
		return
	}
	stage := tile.stage
	stage.puzzleMu.Lock()
	if stage.puzzleState != nil {
		stage.puzzleState.votes[p.id] = true
	}
	stage.puzzleMu.Unlock()
	p.updateBottomText("Reset requested · Everyone in the chamber must choose Reset, or leave.")
	updatePuzzle(p.world, stage)
}
func inPuzzle(p *Player) bool {
	t := p.getTileSync()
	return t != nil && puzzleFor(p.world, t.stage) != nil
}
func openHelp(p *Player) {
	menu := Menu{Name: "help", InfoHtml: `<h2>A little danger. A little wonder.</h2><p>Move with WASD / arrows or the direction pad. Collect red powers, then use Power / Space. Blue boosts let you jump two tiles with Boost / Shift + direction.</p><p>Orange wild bloops and rival teams can hurt you. Health returns only after death. Death is a quick fresh start; your streak resets.</p><p>Safe rooms preserve health and supplies. Dark relays are optional: one is solo, one needs two players. Use ordinary single-tile steps inside puzzle rooms; Boost does not consume charges there. Reset a chamber through the menu when everyone agrees.</p>`, Links: []MenuLink{{Text: "Back", eventHandler: openPauseMenu}}}
	sendMenu(p, menu)
}
func zoneLabel(p *Player) string {
	t := p.getTileSync()
	if t == nil {
		return ""
	}
	if inPuzzle(p) {
		return "PUZZLE"
	}
	if t.stage.safe {
		return "SAFE"
	}
	return "FRONTIER"
}
func teamLabel(team string) string {
	if strings.Contains(team, "fuchsia") {
		return "◆ Fuchsia"
	}
	return "● Sky Blue"
}
