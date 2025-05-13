package main

import (
	"math/rand"
)

type SpawnAction struct {
	Should func(*Player, *Stage) bool
	Action func(*Player) // func of player?
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
		{Should: always, Action: basicSpawnWithRing},
	},
	"tutorial-boost": {
		{Should: always, Action: onCurrentStage(tutorialBoost())},
	},
	"tutorial-1-boost": {
		{Should: always, Action: onCurrentStage(tutorial1Boost())},
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

func tutorial1Boost() func(stage *Stage) {
	return addBoostsAt(7, 4)
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

func spawnPowerup(stage *Stage) {
	shapes := [][][2]int{
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

func basicSpawnWithRing(p *Player) {
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
	if determination2%8 == 0 {
		spawnPowerup(stage)
		spawnNewNPCDoingAction(p, 105, moveRandomlyAndActivatePower, false)
	}
	if determination2 == 0 {
		spawnPowerup(stage)
		npc := spawnNewNPCDoingAction(p, 95, moveAggressively, false)
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
