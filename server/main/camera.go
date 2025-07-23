package main

import (
	"fmt"
	"sync"
)

type Camera struct {
	height, width, padding int
	positionLock           sync.Mutex
	topLeft                *Tile
	outgoing               chan []byte // is == player.updates
}

func (camera *Camera) setView(posY, posX int, stage *Stage) []*Tile {
	y, x := topLeft(len(stage.tiles), len(stage.tiles[0]), camera.height, camera.width, posY, posX)
	region := getRegion(stage.tiles, Rect{y, y + camera.height, x, x + camera.width})

	camera.outgoing <- []byte(fmt.Sprintf(`[~ id="set" y="%d" x="%d" class=""]`, y, x))
	for _, tile := range region {
		//addCamera(tile, camera) // Causes send on closed for some reason
		camera.outgoing <- []byte(swapsForTileNoHighlight(tile))
	}
	newTopLeft := region[0]

	camera.positionLock.Lock()
	defer camera.positionLock.Unlock()
	if camera.topLeft != nil {
		// should be impossible? if needed handle by removing from previous zone
		fmt.Println("ERROR: Camera topLeft not nil in setView")
		return make([]*Tile, 0)
	}

	newTopLeft.primaryZone.addCamera(camera)
	camera.topLeft = newTopLeft
	return region
}

// Awkward?
func (player *Player) tryTrack() {
	player.camera.track(player)
}

func (camera *Camera) track(character Character) {
	focus := character.getTileSync()
	stageH, stageW := len(focus.stage.tiles), len(focus.stage.tiles[0])
	camera.positionLock.Lock() // :( ?
	defer camera.positionLock.Unlock()

	newY := axisAdjust(focus.y, camera.topLeft.y, camera.height, stageH, camera.padding)
	newX := axisAdjust(focus.x, camera.topLeft.x, camera.width, stageW, camera.padding)
	dy := camera.topLeft.y - newY
	dx := camera.topLeft.x - newX
	if dx == 0 && dy == 0 {
		return
	}

	camera.outgoing <- []byte(fmt.Sprintf(`[~ id="shift" y="%d" x="%d" class=""]`, dy, dx))

	updateTiles(camera, newY, newX)
}

func updateTiles(camera *Camera, newY, newX int) {
	updateTilesA(camera, newY, newX)
}

func updateTilesA(camera *Camera, newY, newX int) {
	oldTopLeft := camera.topLeft
	camera.topLeft = oldTopLeft.stage.tiles[newY][newX]
	stage := oldTopLeft.stage
	stageH, stageW := len(oldTopLeft.stage.tiles), len(oldTopLeft.stage.tiles[0])

	oldY0, oldY1 := oldTopLeft.y, oldTopLeft.y+camera.height-1
	oldX0, oldX1 := oldTopLeft.x, oldTopLeft.x+camera.width-1
	newY0, newY1 := newY, newY+camera.height-1
	newX0, newX1 := newX, newX+camera.width-1

	// Tiles that dropped out of view.
	for y := oldY0; y <= oldY1; y++ {
		if y < 0 || y >= stageH {
			continue
		}
		for x := oldX0; x <= oldX1; x++ {
			if x < 0 || x >= stageW {
				continue
			}
			if y < newY0 || y > newY1 || x < newX0 || x > newX1 {
				removeCamera(stage.tiles[y][x], camera)
			}
		}
	}

	// Tiles that came in to view.
	for y := newY0; y <= newY1; y++ {
		if y < 0 || y >= stageH {
			continue
		}
		for x := newX0; x <= newX1; x++ {
			if x < 0 || x >= stageW {
				continue
			}
			if y < oldY0 || y > oldY1 || x < oldX0 || x > oldX1 {
				attachCamera(stage.tiles[y][x], camera)
			}
		}
	}
}

func updateTilesC(camera *Camera, newY, newX int) {
	oldTopLeft := camera.topLeft
	newTopLeft := oldTopLeft.stage.tiles[newY][newX]

	if oldTopLeft.primaryZone != newTopLeft.primaryZone {
		if !oldTopLeft.primaryZone.tryRemoveCamera(camera) {
			// Can this ever happen? Does it add security.
			fmt.Println("ERROR: Camera not found section, cannot add to new section")
			return
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
	for y := newY0; y <= newY1; y++ {
		if y < 0 || y >= stageH {
			continue
		}
		for x := newX0; x <= newX1; x++ {
			if x < 0 || x >= stageW {
				continue
			}
			if y < oldY0 || y > oldY1 || x < oldX0 || x > oldX1 {
				camera.outgoing <- []byte(swapsForTileNoHighlight(stage.tiles[y][x]))
			}
		}
	}
}

func (camera *Camera) drop() {
	camera.positionLock.Lock()
	defer camera.positionLock.Unlock()

	region := getRegion(camera.topLeft.stage.tiles, Rect{camera.topLeft.y, camera.topLeft.y + camera.height, camera.topLeft.x, camera.topLeft.x + camera.width})
	for _, tile := range region {
		removeCamera(tile, camera)
	}
	camera.topLeft = nil
}

func (camera *Camera) drop2() {
	camera.positionLock.Lock()
	defer camera.positionLock.Unlock()

	camera.topLeft.primaryZone.camerasLock.Lock()
	defer camera.topLeft.primaryZone.camerasLock.Unlock()
	delete(camera.topLeft.primaryZone.activeCameras, camera)

	camera.topLeft = nil
}

///////////////
// A: fine - but option C is better

func attachCamera(tile *Tile, cam *Camera) {
	addCamera(tile, cam)
	cam.outgoing <- []byte(swapsForTileNoHighlight(tile))
}

func addCamera(tile *Tile, cam *Camera) {
	tile.camerasLock.Lock()
	defer tile.camerasLock.Unlock()
	tile.cameras[cam] = struct{}{}
}

func removeCamera(tile *Tile, cam *Camera) {
	tile.camerasLock.Lock()
	defer tile.camerasLock.Unlock()
	delete(tile.cameras, cam)
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
