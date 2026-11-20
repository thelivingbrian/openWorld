package main

import (
	"strings"
	"testing"
)

func TestWillowReachCompiledRoutesAndCottageReturns(t *testing.T) {
	c := populateFromJson()
	col := c.Collections["bloop"]
	region := col.Spaces["willow-reach"]
	if region == nil || len(region.Areas) != 1 {
		t.Fatal("missing editable review region")
	}
	out := col.areaOutputFromDescription(region.Areas[0], "")
	if len(out.Tiles) != 48 || len(out.Tiles[0]) != 48 {
		t.Fatal("review region must remain 48 × 48")
	}
	reachable := willowReachable(out.Tiles, [2]int{23, 17})
	for _, point := range [][2]int{{19, 13}, {13, 23}, {35, 39}, {25, 38}, {7, 39}, {35, 11}, {37, 11}, {44, 17}} {
		if !reachable[point] {
			t.Errorf("unreachable landmark %v", point)
		}
	}
	if !strings.Contains(out.Tiles[23][17].DisplayText, "WILLOW REACH") {
		t.Error("arrival lost its introduction")
	}
	seed := out.Interactables[37][11]
	if seed == nil || !seed.Pushable || seed.Walkable {
		t.Fatal("garden seed must be a solid pushable interaction")
	}
	cottages := 0
	for _, door := range out.Transports {
		if !reachable[[2]int{door.SourceY, door.SourceX}] {
			t.Error("unreachable transport", door)
		}
		if door.DestStage == "infirmary:0-0" {
			legacy := getAreaByName(col.Spaces["infirmary"].Areas, door.DestStage)
			tiles, err := col.compileMaterialsFromBlueprint(legacy.Blueprint)
			if err != nil {
				t.Fatal(err)
			}
			if !door.Confirmation || !tiles[door.DestY][door.DestX].Walkable {
				t.Error("old-world connection must confirm and land on walkable ground")
			}
			continue
		}
		cottages++
		if out.Tiles[door.SourceY][door.SourceX].Floor2Css != "scenery-cottage" {
			t.Error("door lost its cottage anchor")
		}
		for dy := -1; dy <= 0; dy++ {
			for dx := -2; dx <= 2; dx++ {
				if out.Tiles[door.SourceY+dy][door.SourceX+dx].Walkable != (dy == 0 && dx == 0) {
					t.Errorf("incorrect cottage footprint at %d,%d offset %d,%d", door.SourceY, door.SourceX, dy, dx)
				}
			}
		}
		if !out.Tiles[door.SourceY-2][door.SourceX].Walkable {
			t.Error("roof extent must allow walking behind the cottage")
		}
		interior := getAreaByName(col.Spaces["willow-cottage"].Areas, door.DestStage)
		if interior == nil {
			t.Fatal("missing cottage interior", door.DestStage)
		}
		room := col.areaOutputFromDescription(*interior, "")
		if len(room.Transports) != 1 {
			t.Fatal("cottage needs one explicit return door")
		}
		exit := room.Transports[0]
		if !willowReachable(room.Tiles, [2]int{door.DestY, door.DestX})[[2]int{exit.SourceY, exit.SourceX}] {
			t.Error("interior exit unreachable")
		}
		if exit.DestStage != out.Name || exit.DestY != door.SourceY+1 || exit.DestX != door.SourceX {
			t.Errorf("return must land outside its own door: %+v", exit)
		}
		if !out.Tiles[exit.DestY][exit.DestX].Walkable {
			t.Error("return destination blocked")
		}
		for _, other := range out.Transports {
			if other.SourceY == exit.DestY && other.SourceX == exit.DestX {
				t.Error("return would immediately teleport again")
			}
		}
	}
	if cottages != 3 {
		t.Errorf("expected three usable cottages, got %d", cottages)
	}
}

func willowReachable(tiles [][]Material, start [2]int) map[[2]int]bool {
	seen := map[[2]int]bool{}
	if !tiles[start[0]][start[1]].Walkable {
		return seen
	}
	seen[start] = true
	queue := [][2]int{start}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, delta := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			next := [2]int{current[0] + delta[0], current[1] + delta[1]}
			y, x := next[0], next[1]
			if y >= 0 && y < len(tiles) && x >= 0 && x < len(tiles[y]) && tiles[y][x].Walkable && !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
	return seen
}

func TestWillowNativeFragmentsAndPaletteCompile(t *testing.T) {
	c := populateFromJson()
	col := c.Collections["bloop"]
	palette := map[string]bool{}
	for _, color := range c.colors {
		palette[color.CssClassName] = true
	}
	for _, p := range col.PrototypeSets["willow-reach"] {
		for _, layer := range []string{p.CssColor, p.MapColor, p.Floor1Css, p.Floor2Css, p.Ceiling1Css, p.Ceiling2Css} {
			for _, token := range strings.Fields(layer) {
				if strings.HasPrefix(token, "willow-") && !palette[strings.TrimSuffix(token, "-b")] {
					t.Errorf("prototype %s references missing color %s", p.ID, token)
				}
			}
		}
	}
	fragments := col.Fragments["willow-reach"]
	if len(fragments) < 10 {
		t.Fatal("missing reusable scenery fragments")
	}
	for _, fragment := range fragments {
		t.Run(fragment.Name, func(t *testing.T) {
			tiles, err := col.compileMaterialsFromBlueprint(fragment.Blueprint)
			if err != nil {
				t.Fatal(err)
			}
			if fragment.ID == "willow-fragment-cottage" {
				if len(tiles) != 2 || len(tiles[0]) != 5 || !tiles[1][2].Walkable || tiles[1][2].Floor2Css != "scenery-cottage" {
					t.Fatal("cottage fragment lost its footprint or door anchor")
				}
			}
			if fragment.ID == "willow-fragment-oak" || fragment.ID == "willow-fragment-pine" || fragment.ID == "willow-fragment-birch" {
				if len(tiles) != 1 || len(tiles[0]) != 1 || tiles[0][0].Walkable || !strings.HasPrefix(tiles[0][0].Floor2Css, "scenery-") {
					t.Fatal("tree fragment must preserve its single blocked anchor and scenery token")
				}
			}
		})
	}
}
