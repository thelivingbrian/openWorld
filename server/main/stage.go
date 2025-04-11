package main

import (
	"log"
	"sync"
)

type Stage struct {
	tiles              [][]*Tile          // [][]**Tile would be weird and open up FP over mutation (also lookup is less fragile)
	playerMap          map[string]*Player // Player Map to Bson map to save whole stage in one command
	playerMutex        sync.RWMutex
	name               string
	north              string
	south              string
	east               string
	west               string
	mapId              string
	spawn              []SpawnAction
	broadcastGroupName string
	weather            string
}

////////////////////////////////////////////////////
// Get / Create and Load Stage

func createStageFromArea(area Area) *Stage {
	spawnAction := spawnActions[area.SpawnStrategy]
	outputStage := Stage{
		tiles:              make([][]*Tile, len(area.Tiles)),
		playerMap:          make(map[string]*Player),
		playerMutex:        sync.RWMutex{},
		name:               area.Name,
		north:              area.North,
		south:              area.South,
		east:               area.East,
		west:               area.West,
		mapId:              area.MapId,
		spawn:              spawnAction,
		broadcastGroupName: area.BroadcastGroup,
		weather:            area.Weather,
	}
	for y := range outputStage.tiles {
		outputStage.tiles[y] = make([]*Tile, len(area.Tiles[y]))
		for x := range outputStage.tiles[y] {
			outputStage.tiles[y][x] = newTile(area.Tiles[y][x], y, x, area.DefaultTileColor)
			outputStage.tiles[y][x].stage = &outputStage
			if area.Interactables != nil && y < len(area.Interactables) && x < len(area.Interactables[y]) {
				description := area.Interactables[y][x]
				if description != nil {
					reaction := interactableReactions[description.Reactions]
					outputStage.tiles[y][x].interactable = &Interactable{name: description.Name, cssClass: description.CssClass, pushable: description.Pushable, walkable: description.Walkable, fragile: description.Fragile, reactions: reaction}
				}
			}
		}
	}
	for _, transport := range area.Transports {
		outputStage.tiles[transport.SourceY][transport.SourceX].teleport = &Teleport{transport.DestStage, transport.DestY, transport.DestX, area.Name, transport.Confirmation, transport.RejectInteractable}

		// Change this
		mat := outputStage.tiles[transport.SourceY][transport.SourceX].material
		mat.Floor1Css = "pink"
		mat.Floor2Css = ""
		outputStage.tiles[transport.SourceY][transport.SourceX].quickSwapTemplate = makeQuickSwapTemplate(mat, transport.SourceY, transport.SourceX)

	}

	return &outputStage
}

////////////////////////////////////////////////////
// Add / Remove Player

func (stage *Stage) addLockedPlayer(player *Player) {
	stage.playerMutex.Lock()
	stage.playerMap[player.id] = player
	stage.playerMutex.Unlock()
}

func (stage *Stage) removeLockedPlayerById(id string) {
	stage.playerMutex.Lock()
	delete(stage.playerMap, id)
	stage.playerMutex.Unlock()
}

func placePlayerOnStageAt(p *Player, stage *Stage, y, x int) {
	if !validCoordinate(y, x, stage) {
		// Extreme - Can this happen ? - Yes log in
		log.Fatal("Fatal: Invalid coords to place on stage.")
	}

	p.setStage(stage)
	spawnItemsFor(p, stage)
	stage.addLockedPlayer(p)
	stage.tiles[y][x].addPlayerAndNotifyOthers(p)
	p.setSpaceHighlights()
	updateEntireExistingScreen(p)
}

///////////////////////////////////////////////////
// Spawn Items

func spawnItemsFor(p *Player, stage *Stage) {
	for i := range stage.spawn {
		stage.spawn[i].activateFor(p, stage)
	}
}

//////////////////////////////////////////////////
// Send Updates

func (stage *Stage) updateAll(update string) {
	stage.updateAllExcept(update, nil)
}

func (stage *Stage) updateAllExcept(update string, ignore *Player) {
	updateAsBytes := []byte(update)
	stage.playerMutex.RLock()
	defer stage.playerMutex.RUnlock()
	for _, player := range stage.playerMap {
		if player == ignore {
			continue
		}
		player.updates <- updateAsBytes
	}
}

/////////////////////////////////////////////////////////////
// Utilities

func validCoordinate(y int, x int, stage *Stage) bool {
	if stage == nil || stage.tiles == nil {
		return false
	}
	if y < 0 || y >= len(stage.tiles) {
		return false
	}
	if x < 0 || x >= len(stage.tiles[y]) {
		return false
	}
	return true
}

func validityByAxis(y int, x int, tiles [][]*Tile) (bool, bool) {
	invalidY, invalidX := false, false
	if y < 0 || y >= len(tiles) {
		invalidY = true
	}
	if x < 0 || x >= len(tiles[0]) { // Not the best, assumes rectangular grid
		invalidX = true
	}
	return invalidY, invalidX
}
