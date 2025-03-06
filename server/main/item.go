package main

import (
	"math/rand"
	"strconv"
	"strings"
)

type SpawnAction struct {
	Should func(*Player, *Stage) bool
	Action func(*Stage) // func of player?
}

func (s *SpawnAction) activateFor(player *Player, stage *Stage) {
	if s.Should == nil || s.Should(player, stage) {
		if s.Action == nil {
			return
		}
		s.Action(stage)
	}
}

var spawnActions = map[string][]SpawnAction{
	"none": []SpawnAction{}, // Same as Should: Always, Action: doNothing
	"": []SpawnAction{
		SpawnAction{Should: always, Action: basicSpawnNoRing},
	},
	"basic-ring": []SpawnAction{
		SpawnAction{Should: always, Action: basicSpawnWithRing},
	},
	"tutorial-boost":   []SpawnAction{SpawnAction{Should: always, Action: tutorialBoost()}},
	"tutorial-power":   []SpawnAction{SpawnAction{Should: always, Action: tutorialPower}},
	"tutorial-2":       []SpawnAction{SpawnAction{Should: oneOutOf(4), Action: spawnBoosts}},
	"tutorial-2-boost": []SpawnAction{SpawnAction{Should: always, Action: tutorial2Boost()}},
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

func checkDistanceFromEdge(gridHeight, gridWidth int) func(*Player, *Stage) bool {
	return func(_ *Player, stage *Stage) bool {
		maxDistance := min((gridHeight-1)/2, (gridWidth-1)/2)
		currentDistance := distanceFromEdgeOfSpace(stage, gridHeight, gridWidth)

		// Faster than equivalent:  1.0 / math.Pow(4.0, float64(maxDistance-currentDistance))
		denominator := 1 << (2 * (maxDistance - currentDistance))
		probability := 1.0 / float64(denominator)

		r := rand.Float64()
		return r < probability
	}
}

func oneOutOf(n int) func(*Player, *Stage) bool {
	return func(_ *Player, stage *Stage) bool {
		r := rand.Intn(n)
		return r == 0
	}
}

func checkTeamName(teamname string) func(*Player, *Stage) bool {
	return func(p *Player, _ *Stage) bool {
		return teamname == p.getTeamNameSync()
	}
}

func excludeInfirmary(p *Player, s *Stage) bool {
	return !strings.HasPrefix(s.name, "infirmary")
}

func distanceFromEdgeOfSpace(stage *Stage, gridHeight, gridWidth int) int {
	if stage == nil {
		return -1
	}
	arr := strings.Split(stage.name, ":")
	if len(arr) != 2 {
		return -1
	}
	coords := strings.Split(arr[1], "-")
	if len(coords) != 2 {
		return -1
	}
	y, err := strconv.Atoi(coords[0])
	if err != nil {
		return -1
	}
	x, err := strconv.Atoi(coords[1])
	if err != nil {
		return -1
	}
	return min(y, x, gridHeight-1-y, gridWidth-1-x)
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
	shapes := [][][2]int{grid3x3, grid3x3, grid5x5, grid5x5, grid5x5, grid7x7, grid7x7, grid9x9, jumpCross(), longCross(5), longCross(3), longCross(3), cross(), x()}
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

func basicSpawnWithRing(stage *Stage) {
	determination := rand.Intn(1000)
	if determination < 400 {
		// Do nothing
	} else if determination < 750 {
		spawnBoosts(stage)
	} else if determination < 990 {
		spawnPowerup(stage)
	} else {
		tryPlaceInteractableOnStage(stage, createRing())
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
