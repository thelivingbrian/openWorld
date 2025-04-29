package main

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestGridActions(t *testing.T) {
	c := populateFromJson()
	col := c.Collections["bloop"]
	var bp *Blueprint

	t.Run("Test make empty grid for TestGridActions", func(t *testing.T) {
		bp = &Blueprint{Tiles: MakeGrid(7, 8, "-")}
		snaps.MatchJSON(t, bp)
	})

	// act
	t.Run("Test 'between' by creating a boundary.", func(t *testing.T) {
		click1 := makeClick(1, 6, "between", "6").withSelected(1, 1)
		click2 := makeClick(5, 6, "between", "6").withSelected(1, 6)
		click3 := makeClick(5, 1, "between", "6").withSelected(5, 6)
		click4 := makeClick(1, 1, "between", "6").withSelected(5, 1)
		click5 := makeClick(3, 4, "between", "6").withSelected(1, 4)

		col.gridClickAction(click1, bp)
		col.gridClickAction(click2, bp)
		col.gridClickAction(click3, bp)
		col.gridClickAction(click4, bp)
		col.gridClickAction(click5, bp)

		snaps.MatchSnapshot(t, bp)
	})

	t.Run("Test 'fill' against boundary.", func(t *testing.T) {
		click1 := makeClick(2, 2, "fill", "1") // fill inner region
		click2 := makeClick(3, 3, "fill", "2") // fill inner again - different
		click3 := makeClick(0, 0, "fill", "4") // fill outer region
		click4 := makeClick(1, 1, "fill", "3") // fill boundary
		click5 := makeClick(1, 1, "fill", "4") // fill boundary to match outer
		click6 := makeClick(2, 2, "fill", "4") // entire grid should now match

		col.gridClickAction(click1, bp)
		snaps.MatchSnapshot(t, bp)
		col.gridClickAction(click2, bp)
		snaps.MatchSnapshot(t, bp)
		col.gridClickAction(click3, bp)
		snaps.MatchSnapshot(t, bp)
		col.gridClickAction(click4, bp)
		snaps.MatchSnapshot(t, bp)
		col.gridClickAction(click5, bp)
		snaps.MatchSnapshot(t, bp)
		col.gridClickAction(click6, bp)
		snaps.MatchSnapshot(t, bp)
	})
}

//////////////////////////////////////////////////////
// Helpers

func makeClick(y, x int, tool, assetId string) *GridClickDetails {
	return &GridClickDetails{Y: y, X: x, Tool: tool, SelectedAssetId: assetId}
}

func (click *GridClickDetails) withSelected(selectedY, selectedX int) *GridClickDetails {
	new := *click
	new.haveASelection = true
	new.selectedY, new.selectedX = selectedY, selectedX
	return &new
}

func MakeGrid(h, w int, prototypeId string) [][]TileData {
	grid := make([][]TileData, h)
	for y := range grid {
		grid[y] = make([]TileData, w)
		for x := range grid[y] {
			grid[y][x].PrototypeId = prototypeId
		}
	}
	return grid
}
