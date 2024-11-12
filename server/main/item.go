package main

import (
	"math/rand"
)

type SpawnAction struct {
	Should func(*Stage) bool
	Action func(*Stage)
}

func (s *SpawnAction) activateFor(stage *Stage) {
	if s.Should(stage) {
		s.Action(stage)
	}
}

var actionMap = map[string]*SpawnAction{
	"none":           &SpawnAction{Should: Always, Action: doNothing},
	"":               &SpawnAction{Should: Always, Action: basicSpawn},
	"tutorial-boost": &SpawnAction{Should: Always, Action: tutorialBoost},
	"tutorial-power": &SpawnAction{Should: Always, Action: tutorialPower},
}

/*
var spawnActions = map[string]func(*Stage){
	"":               basicSpawn,
	"none":           nil,
	"tutorial-boost": tutorialBoost,
	"tutorial-power": tutorialPower,
}
*/

/////////////////////////////////////////////
// Gates

func Always(stage *Stage) bool {
	return true
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
