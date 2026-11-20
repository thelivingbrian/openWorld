package main

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

const VIEW_HEIGHT = 16
const VIEW_WIDTH = 16
const MAX_VIEW_SIZE = 32
const SCENERY_MARGIN = 4

type Camera struct {
	height, width, padding int
	positionLock           sync.Mutex
	topLeft                *Tile
	outgoing               chan<- []byte // Send only: is == player.updates
	dropUpdatesWhenBusy    bool
	resyncRequired         *atomic.Bool
}

func (camera *Camera) send(update []byte) {
	if !camera.dropUpdatesWhenBusy {
		camera.outgoing <- update
		return
	}
	select {
	case camera.outgoing <- update:
	default:
		if camera.resyncRequired != nil {
			camera.resyncRequired.Store(true)
		}
	}
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
	region := getRegion(stage.tiles, Rect{max(0, y-SCENERY_MARGIN), min(len(stage.tiles)-1, y+camera.height-1+SCENERY_MARGIN), max(0, x-SCENERY_MARGIN), min(len(stage.tiles[0])-1, x+camera.width-1+SCENERY_MARGIN)})

	var snapshot strings.Builder
	fmt.Fprintf(&snapshot, `[~ id="view" y="%d" x="%d" class="%d,%d"]`, camera.height, camera.width, len(stage.tiles), len(stage.tiles[0]))

	fmt.Fprintf(&snapshot, `[~ id="set" y="%d" x="%d" class=""]`, y, x)
	for _, tile := range region {
		snapshot.WriteString(swapsForTileNoHighlight(tile))
	}

	camera.send([]byte(snapshot.String()))
	newTopLeft := stage.tiles[y][x]
	newTopLeft.primaryZone.addCamera(camera)
	camera.topLeft = newTopLeft
	return region
}

// Awkward?
func (player *Player) tryTrack() {
	newTiles := player.camera.track(player)
	player.updates <- highlightBoxesForPlayer(player, newTiles)
}

func (camera *Camera) track(character Character) []*Tile {
	focus := character.getTileSync()
	stageH, stageW := len(focus.stage.tiles), len(focus.stage.tiles[0])
	camera.positionLock.Lock()
	defer camera.positionLock.Unlock()
	if camera.topLeft == nil {
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

	camera.send([]byte(fmt.Sprintf(`[~ id="shift" y="%d" x="%d" class=""]`, dy, dx)))

	return updateTiles(camera, newY, newX)
}

func updateTiles(camera *Camera, newY, newX int) []*Tile {
	oldTopLeft := camera.topLeft
	newTopLeft := oldTopLeft.stage.tiles[newY][newX]

	if oldTopLeft.primaryZone != newTopLeft.primaryZone {
		if !oldTopLeft.primaryZone.tryRemoveCamera(camera) {
			// Can this ever happen?
			fmt.Println("ERROR: Camera not found section, cannot add to new section")
			return nil
		}
		newTopLeft.primaryZone.addCamera(camera)
	}

	camera.topLeft = newTopLeft

	stage := oldTopLeft.stage
	stageH, stageW := len(oldTopLeft.stage.tiles), len(oldTopLeft.stage.tiles[0])

	oldY0, oldY1 := oldTopLeft.y-SCENERY_MARGIN, oldTopLeft.y+camera.height-1+SCENERY_MARGIN
	oldX0, oldX1 := oldTopLeft.x-SCENERY_MARGIN, oldTopLeft.x+camera.width-1+SCENERY_MARGIN
	newY0, newY1 := newY-SCENERY_MARGIN, newY+camera.height-1+SCENERY_MARGIN
	newX0, newX1 := newX-SCENERY_MARGIN, newX+camera.width-1+SCENERY_MARGIN

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
				camera.send([]byte(swapsForTileNoHighlight(stage.tiles[y][x])))
			}
		}
	}
	return newTiles
}

func (camera *Camera) drop() {
	camera.positionLock.Lock()
	defer camera.positionLock.Unlock()
	if camera.topLeft == nil {
		fmt.Println("WARN: Camera topLeft is nil in drop")
		return
	}

	camera.topLeft.primaryZone.camerasLock.Lock()
	defer camera.topLeft.primaryZone.camerasLock.Unlock()
	delete(camera.topLeft.primaryZone.activeCameras, camera)

	camera.topLeft = nil
}

func topLeft(gridHeight, gridWidth, viewHeight, viewWidth, y, x int) (row, col int) {
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
	if viewLength >= gridLength {
		return 0
	}

	padding = min(padding, max(1, (viewLength-1)/3))
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

// Resize only from the player's serialized input loop. Keep a bounded viewport
// so camera-zone broadcasts and per-client work remain bounded.
func (player *Player) resizeViewport(rows, columns int) {
	if rows < 8 || rows > MAX_VIEW_SIZE || columns < 8 || columns > MAX_VIEW_SIZE {
		return
	}
	camera := player.camera
	if camera.height == rows && camera.width == columns {
		return
	}
	tile := player.getTileSync()
	camera.drop()
	camera.positionLock.Lock()
	camera.height, camera.width = rows, columns
	camera.positionLock.Unlock()
	tiles := camera.setView(tile.y, tile.x, tile.stage)
	player.updates <- highlightBoxesForPlayer(player, tiles)
}
