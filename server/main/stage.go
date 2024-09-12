package main

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type Stage struct {
	tiles       [][]*Tile          // [][]**Tile would be weird and open up FP over mutation (also lookup is less fragile)
	playerMap   map[string]*Player // Player Map to Bson map to save whole stage in one command
	playerMutex sync.Mutex
	updates     chan Update
	name        string
	north       string
	south       string
	east        string
	west        string
	mapId       string
}

// benchmark this please
func (world *World) getNamedStageOrDefault(name string) *Stage {
	stage := world.getStageByName(name)
	if stage != nil {
		return stage
	}

	stage = world.loadStageByName(name)
	if stage == nil {
		fmt.Println("INVALID STAGE: Area with name " + name + " does not exist.")
		stage = world.loadStageByName("clinic")
		if stage == nil {
			panic("Unable to load default stage")
		}
	}

	return stage
}

func (world *World) getStageByName(name string) *Stage {
	world.wStageMutex.Lock()
	defer world.wStageMutex.Unlock()
	return world.worldStages[name]
}

func (world *World) loadStageByName(name string) *Stage {
	stage := createStageByName(name)
	if stage != nil {
		world.wStageMutex.Lock()
		world.worldStages[name] = stage
		world.wStageMutex.Unlock()
		go stage.sendUpdates()
	}
	return stage
}

func createStageByName(s string) *Stage {
	updatesForStage := make(chan Update)
	area, success := areaFromName(s)
	if !success {
		return nil
	}
	outputStage := Stage{make([][]*Tile, len(area.Tiles)), make(map[string]*Player), sync.Mutex{}, updatesForStage, s, area.North, area.South, area.East, area.West, area.MapId}

	fmt.Println("Creating stage: " + area.Name)

	for y := range outputStage.tiles {
		outputStage.tiles[y] = make([]*Tile, len(area.Tiles[y]))
		for x := range outputStage.tiles[y] {
			outputStage.tiles[y][x] = newTile(materials[area.Tiles[y][x]], y, x, area.DefaultTileColor)
			outputStage.tiles[y][x].stage = &outputStage
			if area.Interactables != nil && y < len(area.Interactables) && x < len(area.Interactables[y]) {
				description := area.Interactables[y][x]
				if description != nil {
					outputStage.tiles[y][x].interactable = &Interactable{pushable: description.Pushable, cssClass: description.CssClass}
				}
			}
		}
	}
	for _, transport := range area.Transports {
		outputStage.tiles[transport.SourceY][transport.SourceX].teleport = &Teleport{transport.DestStage, transport.DestY, transport.DestX}

		// Change this
		mat := outputStage.tiles[transport.SourceY][transport.SourceX].material
		mat.CssColor = "pink"
		outputStage.tiles[transport.SourceY][transport.SourceX].htmlTemplate = makeTileTemplate(mat, transport.SourceY, transport.SourceX)

	}

	return &outputStage
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
	playerIcon := fmt.Sprintf(`<div id="u%d-%d" class="box zu fusia r0"></div>`, player.y, player.x)
	previousBoxes := ""
	if previous.stage == player.stage {
		previousBoxes += fmt.Sprintf(`<div id="u%d-%d" class="box zu"></div>`, previous.y, previous.x)
		previousBoxes += playerBox(previous) // This box may be including the user as well so it needs an update
	}

	player.stage.updates <- Update{player, []byte(highlightBoxesForPlayer(player, tiles) + previousBoxes + playerIcon)}
}

func oobUpdateWithHud(player *Player, tiles []*Tile) {
	// Is this getting blocked? where does this return to
	player.stage.updates <- Update{player, []byte(highlightBoxesForPlayer(player, tiles))}
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
