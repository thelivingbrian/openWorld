package main

import (
	"math/rand"
)

func (stage *Stage) spawnItems() {
	// Very basic spawn algorithm
	// Will spawn on convered tiles with higher freq. than uncovered
	// Will spawn boost and powers with equal probability
	shapes := [][][2]int{grid3x3, grid5x5, grid7x7, grid9x9, jumpCross(), cross(), x()}

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
			// This was mathematically impossible
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
