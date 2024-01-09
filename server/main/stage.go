package main

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type Stage struct {
	tiles       [][]*Tile // [][]**Tile would be weird and open up FP over mutation (also lookup is less fragile)
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
	updatesForStage := make(chan Update)
	area := areaFromName(s)
	outputStage := Stage{make([][]*Tile, len(area.Tiles)), make(map[string]*Player), sync.Mutex{}, updatesForStage, s}

	for y := range outputStage.tiles {
		outputStage.tiles[y] = make([]*Tile, len(area.Tiles[y]))
		for x := range outputStage.tiles[y] {
			outputStage.tiles[y][x] = newTile(materials[area.Tiles[y][x]], y, x)
			outputStage.tiles[y][x].stage = &outputStage
		}
	}
	for _, transport := range area.Transports {
		outputStage.tiles[transport.SourceY][transport.SourceX].teleport = &Teleport{transport.DestStage, transport.DestY, transport.DestX}
		outputStage.tiles[transport.SourceY][transport.SourceX].originalCssClass = "pink"
		outputStage.tiles[transport.SourceY][transport.SourceX].currentCssClass = "pink"

	}

	return &outputStage
}

func getClinic() *Stage {
	return getStageByName("clinic")
}

func (stage *Stage) sendUpdates() {
	for {
		update, ok := <-stage.updates
		if !ok {
			fmt.Println("Stage update channel closed")
			return
		}

		sendUpdate(websocket.TextMessage, update)
	}
}

func sendUpdate(messageType int, update Update) {
	update.player.conn.WriteMessage(messageType, update.update)
}

func (stage *Stage) updateAll(update string) {
	for _, player := range stage.playerMap {
		oobUpdateWithHud(player, update)
	}
}

func (stage *Stage) updateAllExcept(update string, ignore *Player) {
	for _, player := range stage.playerMap {
		if player == ignore {
			continue
		}
		oobUpdateWithHud(player, update)
	}
}

func updateOne(update string, player *Player) {
	oobUpdateWithHud(player, update)
}

func (stage *Stage) markAllDirty() { // This may become prohibitively slow upon players spawning, and full screen probably only needed for spawned player
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

func (stage *Stage) removePlayerById(id string) {
	stage.playerMutex.Lock()
	delete(stage.playerMap, id)
	stage.playerMutex.Unlock()
}

func (stage *Stage) addPlayer(player *Player) {
	stage.playerMutex.Lock()
	stage.playerMap[player.id] = player
	stage.playerMutex.Unlock()
}
