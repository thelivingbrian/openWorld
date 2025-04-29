package main

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestGridActions(t *testing.T) {
	c := populateFromJson()
	col := c.Collections["bloop"]

	// Arrange
	//collection := MockAssetLocator{}
	details := &GridClickDetails{Y: 5, X: 5, Tool: "fill"}
	grid := AddOutlineToGrid(MakeGrid(12, 12, "empty"), 2, 2, 4, 4, "wall")
	bp := &Blueprint{Tiles: grid}
	snaps.MatchJSON(t, bp)

	// act
	col.gridClickAction(details, bp)

	//Assert
	snaps.MatchJSON(t, bp)

	t.Run("should make an int snapshot", func(t *testing.T) {
		col.gridClickAction(&GridClickDetails{Y: 5, X: 5, Tool: "fill", SelectedAssetId: "2"}, bp)
		snaps.MatchSnapshot(t, bp)
	})
}

//////////////////////////////////////////////////////
// Helpers

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

func AddOutlineToGrid(grid [][]TileData,
	top, left, height, width int, outlineId string) [][]TileData {

	H := len(grid)
	if H == 0 {
		return nil
	}
	W := len(grid[0])

	// --- deep-copy the grid -------------------------------------------------
	newGrid := make([][]TileData, H)
	for y := 0; y < H; y++ {
		newGrid[y] = make([]TileData, W)
		copy(newGrid[y], grid[y]) // copies the row’s TileData values
	}

	// --- clamp rectangle to grid bounds ------------------------------------
	bottom := min(top+height, H)
	right := min(left+width, W)

	// --- paint the outline on the copy -------------------------------------
	for y := max(top, 0); y < bottom; y++ {
		for x := max(left, 0); x < right; x++ {
			if y == top || y == bottom-1 || x == left || x == right-1 {
				newGrid[y][x].PrototypeId = outlineId
			}
		}
	}
	return newGrid
}
