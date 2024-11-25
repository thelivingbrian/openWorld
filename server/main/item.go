package main

import (
	"math/rand"
	"strconv"
	"strings"
)

// Both func of player?
type SpawnAction struct {
	Should func(*Player, *Stage) bool // "and" method
	Action func(*Stage)
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
	"none":           []SpawnAction{SpawnAction{}},                                                                                                      // Same as Should: Always, Action: doNothing
	"":               []SpawnAction{SpawnAction{Should: checkDistanceFromEdge(8, 8), Action: basicSpawn}, SpawnAction{Should: nil, Action: basicSpawn}}, // CheckDistanceFromEdge(8, 8)
	"tutorial-boost": []SpawnAction{SpawnAction{Should: always, Action: tutorialBoost}},
	"tutorial-power": []SpawnAction{SpawnAction{Should: always, Action: tutorialPower}},
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

func checkTeamName(teamname string) func(*Player, *Stage) bool {
	return func(p *Player, _ *Stage) bool {
		// getTeamName Method
		p.viewLock.Lock()
		defer p.viewLock.Unlock()
		return p.team == teamname
	}
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

/////////////////////////////////////////////
//  Actions

func doNothing(*Stage) {

}

func tutorialBoost(stage *Stage) {
	stage.tiles[8][8].addBoostsAndNotifyAll()
}

func tutorialPower(stage *Stage) {
	stage.tiles[12][12].addPowerUpAndNotifyAll(grid5x5)
}

func basicSpawn(stage *Stage) {
	// Very basic spawn algorithm
	// Will spawn on convered tiles with higher freq. than uncovered
	// Will spawn boost and powers with equal probability
	shapes := [][][2]int{grid9x9, grid3x3, grid5x5, grid7x7, grid9x9, jumpCross(), cross(), x()}

	randn := rand.Intn(30)
	spawnCovered := randn%3 == 0
	spawnUncovered := randn%7 == 0
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
	//fmt.Println(randn)
}

func sortWalkableTiles(tiles [][]*Tile) ([]*Tile, []*Tile) {
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
