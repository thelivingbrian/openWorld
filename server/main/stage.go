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

func (world *World) getStageByName(name string) (stage *Stage, new bool) {
	new = false

	world.wStageMutex.Lock() // New method
	existingStage, stageExists := world.worldStages[name]
	world.wStageMutex.Unlock()

	if !stageExists {
		new = true
		existingStage, stageExists = createStageByName(name)
		if !stageExists {
			log.Fatal("Unable to create stage")
		}
		world.wStageMutex.Lock()
		world.worldStages[name] = existingStage
		world.wStageMutex.Unlock()
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

		// Change this
		mat := outputStage.tiles[transport.SourceY][transport.SourceX].material
		mat.CssColor = "pink"
		outputStage.tiles[transport.SourceY][transport.SourceX].htmlTemplate = makeTileTemplate(mat, transport.SourceY, transport.SourceX)

	}

	return &outputStage, true
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
	if update.player.conn != nil {
		update.player.conn.WriteMessage(websocket.TextMessage, update.update)
	} else {
		fmt.Println("WARN: Attempted to serve update to expired connection.")
	}
}

// Enqueue updates

func (stage *Stage) updateAllWithHud(tiles []*Tile) {
	stage.playerMutex.Lock()
	defer stage.playerMutex.Unlock()
	for _, player := range stage.playerMap {
		oobUpdateWithHud(player, tiles)
	}
}

func (stage *Stage) updateAllWithHudExcept(ignore *Player, tiles []*Tile) {
	stage.playerMutex.Lock()
	defer stage.playerMutex.Unlock()
	for _, player := range stage.playerMap {
		if player == ignore {
			continue
		}
		oobUpdateWithHud(player, tiles)
	}
}

func updateOneAfterMovement(player *Player, tiles []*Tile, previous *Tile) {
	playerIcon := fmt.Sprintf(`<div class="box zp fusia" id="p%d-%d" hx-swap-oob="true"></div>`, player.y, player.x)
	previousBox := playerBox(previous)

	player.stage.updates <- Update{player, []byte(highlightBoxesForPlayer(player, tiles) + previousBox + playerIcon)}
}

func oobUpdateWithHud(player *Player, tiles []*Tile) {
	// Is this getting blocked? where does this return to
	player.stage.updates <- Update{player, []byte(highlightBoxesForPlayer(player, tiles))}
}

// Wrong file
func highlightBoxesForPlayer(player *Player, tiles []*Tile) string {
	highlights := ""
	// Create slice of proper size? Currently has many null entries

	// Still risk here of concurrent read/write?
	for _, tile := range tiles {
		if tile == nil {
			continue
		}
		if tile.stage != player.stage {
			continue
		}
		_, impactsHud := player.actions.shiftHighlights[tile]
		if impactsHud && player.actions.boostCounter > 0 {
			highlights += oobHighlightBox(tile, shiftHighlighter(tile))
			continue
		}
		_, impactsHud = player.actions.spaceHighlights[tile]
		if impactsHud {
			highlights += oobHighlightBox(tile, spaceHighlighter(tile)) //oobColoredTile(tile, spaceHighlighter(tile))
			continue
		}

		// Empty Highlight?
		highlights += oobHighlightBox(tile, "")
	}

	// Maybe should just be the actual color, or do similar check to see if it was impacted
	// To work this way needs to come without classic swap
	//playerIcon := fmt.Sprintf(`<div class="box zp fusia" id="p%d-%d" hx-swap-oob="true"></div>`, player.y, player.x)

	return highlights // + playerIcon
}

func (stage *Stage) updateAll(update string) {
	stage.playerMutex.Lock()
	defer stage.playerMutex.Unlock()
	for _, player := range stage.playerMap {
		updateOne(update, player)
	}
}

func (stage *Stage) updateAllExcept(update string, ignore *Player) {
	stage.playerMutex.Lock()
	defer stage.playerMutex.Unlock()
	for _, player := range stage.playerMap {
		if player == ignore {
			continue
		}
		updateOne(update, player)
	}
}

func updateOne(update string, player *Player) {
	player.stage.updates <- Update{player, []byte(update)}
}

func updateScreenFromScratch(player *Player) {
	player.stage.updates <- Update{player, htmlFromPlayer(player)}
}
