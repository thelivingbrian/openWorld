package main

import (
	"fmt"
	"sync"
)

type Stage struct {
	tiles       [][]*Tile          // [][]**Tile would be weird and open up FP over mutation (also lookup is less fragile)
	playerMap   map[string]*Player // Player Map to Bson map to save whole stage in one command
	playerMutex sync.RWMutex
	name        string
	north       string
	south       string
	east        string
	west        string
	mapId       string
	spawn       []SpawnAction
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

// compare these two
func (world *World) getStageByName(name string) *Stage {
	world.wStageMutex.Lock()
	defer world.wStageMutex.Unlock()
	return world.worldStages[name]
}

func (world *World) loadStageByName(name string) *Stage {
	area, success := areaFromName(name)
	if !success {
		return nil
	}
	stage := createStageFromArea(area)
	if area.LoadStrategy == "Individual" {
		return stage
	}
	if stage != nil {
		world.wStageMutex.Lock()
		world.worldStages[name] = stage
		world.wStageMutex.Unlock()
	}
	return stage
}

func createStageFromArea(area Area) *Stage {
	spawnAction := spawnActions[area.SpawnStrategy]
	outputStage := Stage{make([][]*Tile, len(area.Tiles)), make(map[string]*Player), sync.RWMutex{}, area.Name, area.North, area.South, area.East, area.West, area.MapId, spawnAction}
	for y := range outputStage.tiles {
		outputStage.tiles[y] = make([]*Tile, len(area.Tiles[y]))
		for x := range outputStage.tiles[y] {
			outputStage.tiles[y][x] = newTile(materials[area.Tiles[y][x]], y, x, area.DefaultTileColor)
			outputStage.tiles[y][x].stage = &outputStage
			if area.Interactables != nil && y < len(area.Interactables) && x < len(area.Interactables[y]) {
				description := area.Interactables[y][x]
				if description != nil {
					reaction := interactableReactions[description.Reactions]
					outputStage.tiles[y][x].interactable = &Interactable{name: description.Name, cssClass: description.CssClass, pushable: description.Pushable, fragile: description.Fragile, reactions: reaction}
				}
			}
		}
	}
	for _, transport := range area.Transports {
		outputStage.tiles[transport.SourceY][transport.SourceX].teleport = &Teleport{transport.DestStage, transport.DestY, transport.DestX, area.Name, transport.Confirmation}

		// Change this
		mat := outputStage.tiles[transport.SourceY][transport.SourceX].material
		mat.CssColor = "pink"
		mat.Floor1Css = ""
		mat.Floor2Css = ""
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

// Enqueue updates

func updateOneAfterMovement(player *Player, tiles []*Tile, previous *Tile) {
	playerIcon := playerBoxSpecifc(player.y, player.x, player.icon)
	previousBoxes := ""
	if previous != nil && previous.stage == player.stage {
		previousBoxes += playerBox(previous)
	}

	player.updates <- []byte(highlightBoxesForPlayer(player, tiles) + previousBoxes + playerIcon)
}

func (stage *Stage) updateAll(update string) {
	stage.updateAllExcept(update, nil)
}

func (stage *Stage) updateAllExcept(update string, ignore *Player) {
	stage.playerMutex.RLock()
	defer stage.playerMutex.RUnlock()
	updateAsBytes := []byte(update)
	for _, player := range stage.playerMap {
		if player == ignore {
			continue
		}
		player.updates <- updateAsBytes
	}
}

// not related to stage?
func updateOne(update string, player *Player) {
	player.updates <- []byte(update)
}

func updateScreenFromScratch(player *Player) {
	clearChannel(player.updates)
	player.clearUpdateBuffer <- struct{}{}
	player.updates <- htmlFromPlayer(player)
}

func clearChannel(ch chan []byte) {
	for {
		select {
		case <-ch: // Read from the channel
			// Do nothing, just drain
		default: // Exit when the channel is empty
			return
		}
	}
}
