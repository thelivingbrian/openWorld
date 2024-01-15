package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Stage struct {
	tiles       [][]*Tile          // [][]**Tile would be weird and open up FP over mutation (also lookup is less fragile)
	playerMap   map[string]*Player // Player Map to Bson map to save whole stage in one command
	playerMutex sync.Mutex
	updates     chan Update
	name        string
}

func getStageByName(name string) (stage *Stage, new bool) {
	new = false

	stageMutex.Lock() // Inject this
	existingStage, stageExists := stageMap[name]
	stageMutex.Unlock()

	if !stageExists {
		new = true
		existingStage, stageExists = createStageByName(name)
		if !stageExists {
			log.Fatal("Unable to create stage")
		}
		stageMap[name] = existingStage
	}

	return existingStage, new
}

func createStageByName(s string) (*Stage, bool) {
	updatesForStage := make(chan Update)
	area, success := areaFromName(s)
	if !success {
		return nil, false
	}
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

	return &outputStage, true
}

/*
func (stage *Stage) markAllDirty() { // This may become prohibitively slow upon players spawning, and full screen probably only needed for spawned player
	stage.playerMutex.Lock()
	currentPlayerCount := len(stage.playerMap)
	stage.playerMutex.Unlock()

	if currentPlayerCount > 4 {
		startingScreenUpdate(stage)
	} else {
		fullUpdate(stage)
	}
}
*/

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

////////////////////////////////////////////////////////////
//   Updates

func (stage *Stage) sendUpdates() {
	for {
		update, ok := <-stage.updates
		if !ok {
			fmt.Println("Stage update channel closed")
			return
		}

		sendUpdate(update)
	}
}

func sendUpdate(update Update) {
	update.player.connLock.Lock()
	defer update.player.connLock.Unlock()
	update.player.conn.WriteMessage(websocket.TextMessage, update.update)
}

func (stage *Stage) updateAll(update string) {
	stage.playerMutex.Lock()
	defer stage.playerMutex.Unlock()
	for _, player := range stage.playerMap {
		oobUpdateWithHud(player, update)
	}
}

func (stage *Stage) updateAllExcept(update string, ignore *Player) {
	stage.playerMutex.Lock()
	defer stage.playerMutex.Unlock()
	for _, player := range stage.playerMap {
		if player == ignore {
			continue
		}
		oobUpdateWithHud(player, update)
	}
}

func updateOneWithHud(update string, player *Player) {
	oobUpdateWithHud(player, update)
}

func updateOne(update string, player *Player) {
	player.stage.updates <- Update{player, []byte(update)}
}

/*
func startingScreenUpdate(stage *Stage) {
	stage.playerMutex.Lock()
	defer stage.playerMutex.Unlock()
	screenHtml := htmlFromStage(stage)
	for _, player := range stage.playerMap {
		updateScreenWithStarter(player, screenHtml)
	}
}

func fullUpdate(stage *Stage) {
	stage.playerMutex.Lock()
	defer stage.playerMutex.Unlock()

	for _, player := range stage.playerMap { // This throws if playermap is empty?!
		updateScreenFromScratch(player)
	}
}
*/

func updateScreenWithStarter(player *Player, html string) {
	if player.isDead() {
		respawn(player)
		return
	}
	html += hudAsOutOfBound(player)
	player.stage.updates <- Update{player, []byte(html)}
}

func updateScreenFromScratch(player *Player) {
	if player.isDead() {
		respawn(player)
		return
	}
	player.stage.updates <- Update{player, htmlFromPlayer(player)}
}

func oobUpdateWithHud(player *Player, update string) {
	player.stage.updates <- Update{player, []byte(update + hudAsOutOfBound(player))}
}
