package main

import (
	"math/rand"
	"strconv"
	"time"
)

type SpawnAction struct {
	Should func(*Player, *Stage) bool
	Action func(*Player)
}

func (s *SpawnAction) activateFor(player *Player, stage *Stage) {
	if s.Should == nil || s.Should(player, stage) {
		if s.Action == nil {
			return
		}
		s.Action(player)
	}
}

var spawnActions = map[string][]SpawnAction{
	"none": {}, // Same as Should: Always, Action: doNothing
	"": {
		{Should: always, Action: onCurrentStage(basicSpawnNoRing)},
	},
	"basic-ring": {
		{Should: always, Action: basicSpawnWithRingAndNPCs},
	},
	"basic-weak": {
		{Should: always, Action: onCurrentStage(basicSpawnWeak)},
	},
	"tutorial-boost": {
		{Should: always, Action: onCurrentStage(tutorialBoost())},
	},
	"tutorial-1-skip": {
		{Should: both(checkYCoord(3), checkXCoord(3)), Action: openNamedMenuAfterDelay("skip", 0)},
	},
	"tutorial-1-menu": {
		{Should: checkYCoord(0), Action: openNamedMenuAfterDelay("pause", 1600)},
	},
	"tutorial-1-boost": {
		{Should: checkXCoord(0), Action: tutorial1Boost},
	},
	"tutorial-1-ring": {
		{Should: always, Action: tutorial1Ring},
	},
	"tutorial-1-npc": {
		{Should: always, Action: tutorial1Npc},
	},
	"tutorial-power": {
		{Should: always, Action: onCurrentStage(tutorialPower)},
	},
	"tutorial-2": {
		{Should: oneOutOf(4), Action: onCurrentStage(spawnBoosts)},
	},
	"tutorial-2-boost": {
		{Should: always, Action: onCurrentStage(tutorial2Boost())},
	},
}

func onCurrentStage(f func(*Stage)) func(*Player) {
	return func(p *Player) {
		tile := p.getTileSync()
		f(tile.stage)
	}
}

/////////////////////////////////////////////
// Gates

func always(*Player, *Stage) bool {
	return true
}

func both(f1, f2 func(*Player, *Stage) bool) func(*Player, *Stage) bool {
	return func(p *Player, s *Stage) bool {
		return f1(p, s) && f2(p, s)
	}
}

func oneOutOf(n int) func(*Player, *Stage) bool {
	return func(_ *Player, stage *Stage) bool {
		r := rand.Intn(n)
		return r == 0
	}
}

func max(vals ...int) int {
	if len(vals) == 0 {
		panic("max requires at least one argument")
	}
	maxVal := vals[0]
	for _, v := range vals[1:] {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}

func min(vals ...int) int {
	if len(vals) == 0 {
		panic("min requires at least one argument")
	}
	minVal := vals[0]
	for _, v := range vals[1:] {
		if v < minVal {
			minVal = v
		}
	}
	return minVal
}

func checkYCoord(check int) func(p *Player, s *Stage) bool {
	return func(p *Player, s *Stage) bool {
		return p.getTileSync().y == check
	}
}

func checkXCoord(check int) func(p *Player, s *Stage) bool {
	return func(p *Player, s *Stage) bool {
		return p.getTileSync().x == check
	}
}

/////////////////////////////////////////////
//  Actions

func doNothing(*Stage) {

}

func addBoostsAt(y, x int) func(stage *Stage) {
	return func(stage *Stage) {
		if y >= len(stage.tiles) || x >= len(stage.tiles[y]) {
			return
		}
		stage.tiles[y][x].addBoostsAndNotifyAll()
	}
}

func tutorialBoost() func(stage *Stage) {
	return addBoostsAt(8, 8)
}

func tutorial1Boost(player *Player) {
	stage := player.getTileSync().stage
	stage.tiles[7][4].addBoostsAndNotifyAll()
}

func tutorial1Npc(player *Player) {
	stage := player.getTileSync().stage
	tiles0 := getRegion(stage.tiles, Rect{2, 5, 5, 6})
	tiles1 := getRegion(stage.tiles, Rect{6, 7, 2, 4})
	tiles := append(tiles0, tiles1...)
	tile, ok := pickOne(tiles)
	if !ok {
		return
	}
	randStr := strconv.Itoa(rand.Intn(16))
	spawnNewNPCDoingAction(player, randStr, 110, 60, moveAgressiveRand(shortShapes), tile)
}

type Rect struct {
	MinY, MaxY int
	MinX, MaxX int
}

func getRegion[T any](grid [][]T, r Rect) []T {
	var out []T
	for y := r.MinY; y <= r.MaxY; y++ {
		if y < 0 || y >= len(grid) {
			continue
		}
		row := grid[y]
		minX := clamp(r.MinX, 0, len(row)-1)
		maxX := clamp(r.MaxX, 0, len(row)-1)
		if maxX < minX {
			continue
		}
		out = append(out, row[minX:maxX+1]...)
	}
	return out
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ok is false when the slice is empty.
func pickOne[T any](items []T) (item T, ok bool) {
	if len(items) == 0 {
		return
	}
	item = items[rand.Intn(len(items))]
	ok = true
	return
}

func tutorial1Ring(player *Player) {
	stage := player.getTileSync().stage
	tiles := []*Tile{
		stage.tiles[2][2],
		stage.tiles[13][2],
		stage.tiles[2][13],
	}
	n := rand.Intn(len(tiles))
	n2 := mod(n+1, len(tiles))
	tile := tiles[n]
	tile2 := tiles[n2]
	ring := Interactable{
		name:     "ring-big",
		cssClass: "gold-b thick r1",
		pushable: true,
		fragile:  true,
	}
	copy := ring
	trySetInteractable(tile, &ring)
	trySetInteractable(tile2, &copy)
}

func openNamedMenuAfterDelay(name string, delay int) func(*Player) {
	return func(p *Player) {
		go func() {
			time.Sleep(time.Millisecond * time.Duration(delay))
			ownLock := p.tangibilityLock.TryLock()
			if !ownLock {
				return
			}
			defer p.tangibilityLock.Unlock()
			if !p.tangible {
				return
			}
			p.addMoneyAndUpdate(0) // Set Peak Wealth
			turnMenuOnByName(p, name)
		}()
	}
}

func tutorial2Boost() func(stage *Stage) {
	return addBoostsAt(10, 11)
}

func tutorialPower(stage *Stage) {
	stage.tiles[12][12].addPowerUpAndNotifyAll(grid5x5)
}

func spawnBoosts(stage *Stage) {
	_, uncoveredTiles := sortWalkableTiles(stage.tiles)
	tile := uncoveredTiles[rand.Intn(len(uncoveredTiles))]
	tile.addBoostsAndNotifyAll()
}

var shortShapes = [][][2]int{
	grid3x3,
	grid5x5,
	jumpCross(),
	x(),
}

var weakShapes = [][][2]int{
	grid3x3, grid3x3,
	cross(),
	jumpCross(), jumpCross(),
	x(),
}

var standardShapes = [][][2]int{
	diagonalBlock(true, 2), diagonalBlock(false, 2),
	diagonalBlock(true, 3), diagonalBlock(false, 3),
	grid3x3, grid3x3,
	grid5x5, grid5x5, grid5x5,
	grid7x7, grid7x7,
	grid9x9,
	jumpCross(),
	longCross(5),
	longCross(3),
	x(),
}

func spawnPowerupShort(stage *Stage) {
	spawnPowerupFromSet(stage, shortShapes)
}

func spawnPowerup(stage *Stage) {
	spawnPowerupFromSet(stage, standardShapes)
}

func spawnPowerupFromSet(stage *Stage, shapes [][][2]int) {
	index := rand.Intn(len(shapes))
	tiles, uncoveredTiles := sortWalkableTiles(stage.tiles)
	tiles = append(tiles, uncoveredTiles...)
	tile := tiles[rand.Intn(len(tiles))]
	tile.addPowerUpAndNotifyAll(shapes[index])
}

func basicSpawnOld(stage *Stage) {
	// Very basic spawn algorithm
	// Will spawn on convered tiles with higher freq. than uncovered
	// Will spawn boost and powers with equal probability
	shapes := [][][2]int{grid9x9, grid5x5} //, grid3x3, grid5x5, grid7x7, grid9x9, jumpCross(), cross(), x()}

	randn := rand.Intn(30)
	spawnCovered := false //randn%3 == 0
	spawnUncovered := randn%3 == 0
	heads := randn%2 == 0

	coveredTiles, uncoveredTiles := sortWalkableTiles(stage.tiles)

	if spawnCovered && len(coveredTiles) != 0 {
		randomIndex := rand.Intn(len(coveredTiles))
		if heads {
			randomIndex2 := rand.Intn(len(shapes))
			coveredTiles[randomIndex].addPowerUpAndNotifyAll(shapes[randomIndex2])
		} else {
			coveredTiles[randomIndex].addBoostsAndNotifyAll()
		}
	}

	if spawnUncovered {
		randomIndex := rand.Intn(len(uncoveredTiles))
		if heads {
			randomIndex2 := rand.Intn(len(shapes))
			uncoveredTiles[randomIndex].addPowerUpAndNotifyAll(shapes[randomIndex2])
		} else {
			uncoveredTiles[randomIndex].addBoostsAndNotifyAll()
		}
	}
}

func basicSpawnWithRingAndNPCs(p *Player) {
	determination := rand.Intn(1000)
	if determination < 350 {
		// Do nothing
		return
	}
	stage := p.getTileSync().stage
	if determination < 665 {
		spawnBoosts(stage)
	} else if determination < 925 {
		spawnPowerup(stage)
	} else if determination < 975 {
		tryPlaceInteractableOnStage(stage, createRing())
	}

	determination2 := rand.Intn(16)
	lifeInSeconds := 450
	if determination2%8 == 0 {
		spawnPowerup(stage)
		spawnNewNPCDoingAction(p, "npc", 105, lifeInSeconds, moveRandomlyAndActivatePower, nil)
	}
	if determination2 == 0 {
		spawnPowerup(stage)
		npc := spawnNewNPCDoingAction(p, "npc", 95, lifeInSeconds, moveAgressiveRand(shapesNpc), nil)
		npc.money.Add(int64(125))
	}

}

func basicSpawnNoRing(stage *Stage) {
	determination := rand.Intn(1000)
	if determination < 250 {
		// Do nothing
	} else if determination < 700 {
		spawnBoosts(stage)
	} else {
		spawnPowerup(stage)
	}
}

func basicSpawnWeak(stage *Stage) {
	determination := rand.Intn(1000)
	if determination < 400 {
		// Do nothing
	} else if determination < 750 {
		spawnBoosts(stage)
	} else {
		spawnPowerupFromSet(stage, weakShapes)
	}
}

func sortWalkableTiles(tiles [][]*Tile) (covered []*Tile, uncovered []*Tile) {
	var outCovered, outUncovered []*Tile
	for y := range tiles {
		for x := range tiles[y] {
			if tiles[y][x].material.Walkable {
				if tiles[y][x].material.Ceiling1Css != "" || tiles[y][x].material.Ceiling2Css != "" {
					outCovered = append(outCovered, tiles[y][x])
					continue
				}
				outUncovered = append(outUncovered, tiles[y][x])
			}
		}
	}
	return outCovered, outUncovered
}

func walkableTiles(tiles [][]*Tile) []*Tile {
	var out []*Tile
	for y := range tiles {
		for x := range tiles[y] {
			if tiles[y][x].material.Walkable {
				out = append(out, tiles[y][x])
			}
		}
	}
	return out
}
