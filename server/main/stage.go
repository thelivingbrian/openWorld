package main

import (
	"fmt"
	"sync"
)

type Stage struct {
	tiles       [][]*Tile
	playerMap   map[string]*Player
	playerMutex sync.Mutex
	name        string
}

func (stage *Stage) markAllDirty() {
	if len(stage.playerMap) > 4 {
		startingScreenUpdate(stage)
	} else {
		fullUpdate(stage)
	}
}

func startingScreenUpdate(stage *Stage) {
	screenHtml := htmlFromStage(stage)
	for _, player := range stage.playerMap {
		updateScreenWithStarter(player, screenHtml, updates)
	}
}

func fullUpdate(stage *Stage) {
	for _, player := range stage.playerMap {
		updateScreenFromScratch(player, updates)
	}
}

func (stage *Stage) damageAt(coords [][2]int) {
	for _, pair := range coords {
		if validCoordinate(pair[0], pair[1], stage.tiles) {
			for _, player := range stage.tiles[pair[0]][pair[1]].playerMap {
				player.health += -50
				if player.isDead() {
					fmt.Println(player.id + " has died")

					deadPlayerTile := stage.tiles[pair[0]][pair[1]]
					deadPlayerTile.removePlayer(player.id)
					removePlayerById(stage, player.id) // Is stage player map used (maybe for player count only?)
					stage.markAllDirty()

					updateScreenWithStarter(player, "", updates) // Player is no longer on screen, Should this be a different method or should update happen elsewhere?
				}

			}
		}
	}
}

func removePlayerById(stage *Stage, id string) {
	stage.playerMutex.Lock()
	delete(stage.playerMap, id)
	stage.playerMutex.Unlock()
}

func getStageByName(name string) *Stage {
	stageMutex.Lock()
	existingStage, stageExists := stageMap[name]
	if !stageExists {
		newStage := createStageByName(name)
		stagePtr := &newStage
		stageMap[name] = stagePtr
		existingStage = stagePtr
	}
	stageMutex.Unlock()
	return existingStage
}

func getClinic() *Stage {
	return getStageByName("clinic")
}
