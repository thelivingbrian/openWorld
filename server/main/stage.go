package main

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type Stage struct {
	tiles       [][]*Tile
	playerMap   map[string]*Player
	playerMutex sync.Mutex
	updates     chan Update
	name        string
}

func getStageByName(name string) *Stage {
	stageMutex.Lock()
	existingStage, stageExists := stageMap[name]
	if !stageExists {
		existingStage = createStageAndHandleUpdates(name)
	}
	stageMutex.Unlock()
	return existingStage
}

func createStageAndHandleUpdates(name string) *Stage {
	fmt.Println("New Stage " + name)
	stage := createStageByName(name)
	stageMap[name] = stage
	go stage.sendUpdates()
	return stage

}

func createStageByName(s string) *Stage {
	area := areaFromName(s)
	tiles := make([][]*Tile, len(area.Tiles))
	for y := range tiles {
		tiles[y] = make([]*Tile, len(area.Tiles[y]))
		for x := range tiles[y] {
			tiles[y][x] = newTile(materials[area.Tiles[y][x]], y, x)
		}
	}
	for _, transport := range area.Transports {
		tiles[transport.SourceY][transport.SourceX].teleport = &Teleport{transport.DestStage, transport.DestY, transport.DestX}
	}
	updates := make(chan Update)
	//go sendUpdates(updates)
	return &Stage{tiles, make(map[string]*Player), sync.Mutex{}, updates, s}
}

func getClinic() *Stage {
	return getStageByName("clinic")
}

func (stage *Stage) sendUpdates() {
	for {
		update, ok := <-stage.updates
		if !ok {
			fmt.Println("hi")
			return
		}
		//fmt.Println("yo")
		//fmt.Println(update.player.stageName)
		sendUpdate(websocket.TextMessage, update)
	}
}

func sendUpdate(messageType int, update Update) {
	update.player.conn.WriteMessage(messageType, update.update)
}

func (stage *Stage) updateAll(update string) {
	for _, player := range playerMap {

		oobUpdateWithHud(player, update)
		//stage.updates <- Update{player, []byte(update)}
	}
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
		updateScreenWithStarter(player, screenHtml)
	}
}

func fullUpdate(stage *Stage) {
	for _, player := range stage.playerMap {
		updateScreenFromScratch(player)
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

					updateScreenWithStarter(player, "") // Player is no longer on screen, Should this be a different method or should update happen elsewhere?
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
