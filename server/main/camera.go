package main

import (
	"fmt"
	"sync"
)

// Odd grid size allows centering player with padding - Has problems with smaller grid
const VIEW_HEIGHT = 16
const VIEW_WIDTH = 16

type Camera struct {
	height, width, padding int
	positionLock           sync.Mutex
	topLeft                *Tile
	outgoing               chan<- []byte // Send only: is == player.updates
}

func (camera *Camera) setView(posY, posX int, stage *Stage) []*Tile {
	camera.positionLock.Lock()
	defer camera.positionLock.Unlock()
	if camera.topLeft != nil {
		// should be impossible? if needed handle by removing from previous zone
		fmt.Println("ERROR: Camera topLeft not nil in setView")
		return make([]*Tile, 0)
	}

	y, x := topLeft(len(stage.tiles), len(stage.tiles[0]), camera.height, camera.width, posY, posX)
	region := getRegion(stage.tiles, Rect{y, y + camera.height - 1, x, x + camera.width - 1})

	camera.outgoing <- []byte(fmt.Sprintf(`[~ id="set" y="%d" x="%d" class=""]`, y, x))
	for _, tile := range region {
		camera.outgoing <- []byte(swapsForTileNoHighlight(tile))
	}

	newTopLeft := region[0]
	newTopLeft.primaryZone.addCamera(camera)
	camera.topLeft = newTopLeft
	return region
}

// Awkward?
func (player *Player) tryTrack() {
	newTiles := player.camera.track(player)
	player.updates <- []byte(highlightBoxesForPlayer(player, newTiles))
}

func (camera *Camera) track(character Character) []*Tile {
	focus := character.getTileSync()
	stageH, stageW := len(focus.stage.tiles), len(focus.stage.tiles[0])
	camera.positionLock.Lock()
	defer camera.positionLock.Unlock()
	if camera.topLeft == nil {
		// This has Occurred. Unsure why.
		fmt.Println("WARN: Camera topLeft is nil in track")
		return nil
	}

	newY := axisAdjust(focus.y, camera.topLeft.y, camera.height, stageH, camera.padding)
	newX := axisAdjust(focus.x, camera.topLeft.x, camera.width, stageW, camera.padding)
	dy := camera.topLeft.y - newY
	dx := camera.topLeft.x - newX
	if dx == 0 && dy == 0 {
		return nil
	}

	camera.outgoing <- []byte(fmt.Sprintf(`[~ id="shift" y="%d" x="%d" class=""]`, dy, dx))

	return updateTiles(camera, newY, newX)
}

func updateTiles(camera *Camera, newY, newX int) []*Tile {
	oldTopLeft := camera.topLeft
	newTopLeft := oldTopLeft.stage.tiles[newY][newX]

	if oldTopLeft.primaryZone != newTopLeft.primaryZone {
		if !oldTopLeft.primaryZone.tryRemoveCamera(camera) {
			// Can this ever happen? Does it add security.
			fmt.Println("ERROR: Camera not found section, cannot add to new section")
			return nil
		}
		newTopLeft.primaryZone.addCamera(camera)
	}

	camera.topLeft = newTopLeft

	stage := oldTopLeft.stage
	stageH, stageW := len(oldTopLeft.stage.tiles), len(oldTopLeft.stage.tiles[0])

	oldY0, oldY1 := oldTopLeft.y, oldTopLeft.y+camera.height-1
	oldX0, oldX1 := oldTopLeft.x, oldTopLeft.x+camera.width-1
	newY0, newY1 := newY, newY+camera.height-1
	newX0, newX1 := newX, newX+camera.width-1

	// Tiles that came into view.
	newTiles := make([]*Tile, 0)
	for y := newY0; y <= newY1; y++ {
		if y < 0 || y >= stageH {
			continue
		}
		for x := newX0; x <= newX1; x++ {
			if x < 0 || x >= stageW {
				continue
			}
			if y < oldY0 || y > oldY1 || x < oldX0 || x > oldX1 {
				newTiles = append(newTiles, stage.tiles[y][x])
				camera.outgoing <- []byte(swapsForTileNoHighlight(stage.tiles[y][x]))
			}
		}
	}
	return newTiles
}

func (camera *Camera) drop() {
	camera.positionLock.Lock()
	defer camera.positionLock.Unlock()
	if camera.topLeft == nil {
		// This has happenned. Not sure why.
		fmt.Println("WARN: Camera topLeft is nil in drop")
		return
	}

	camera.topLeft.primaryZone.camerasLock.Lock()
	defer camera.topLeft.primaryZone.camerasLock.Unlock()
	delete(camera.topLeft.primaryZone.activeCameras, camera)

	camera.topLeft = nil
}

func topLeft(gridHeight, gridWidth, viewHeight, viewWidth, y, x int) (row, col int) {
	// clamp the requested window
	if viewHeight > gridHeight {
		viewHeight = gridHeight
	}
	if viewWidth > gridWidth {
		viewWidth = gridWidth
	}

	// Ideal top‑left: put (r,c) at ⌊n/2⌋, ⌊m/2⌋ inside the window.
	row = y - viewHeight/2
	col = x - viewWidth/2

	// Clamp to the valid range 0 … rows‑n and 0 … cols‑m.
	if row < 0 {
		row = 0
	}
	if col < 0 {
		col = 0
	}
	if row > gridHeight-viewHeight {
		row = gridHeight - viewHeight
	}
	if col > gridWidth-viewWidth {
		col = gridWidth - viewWidth
	}
	return
}

func axisAdjust(pos, oldBoundary, viewLength, gridLength, padding int) int {
	// Nothing to do if the view already covers the whole axis.
	if viewLength >= gridLength {
		return 0
	}

	lo := oldBoundary + padding                  // nearest allowed position inside view
	hi := oldBoundary + viewLength - padding - 1 // farthest allowed position inside view

	newBoundary := oldBoundary
	switch {
	case pos < lo: // too close to the top/left edge
		newBoundary -= lo - pos // shift up/left just enough
	case pos > hi: // too close to the bottom/right edge
		newBoundary += pos - hi // shift down/right just enough
	}

	// Clamp to grid.
	if newBoundary < 0 {
		newBoundary = 0
	}
	max := gridLength - viewLength
	if newBoundary > max {
		newBoundary = max
	}
	return newBoundary
}
