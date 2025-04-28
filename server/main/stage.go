package main

import (
	"log"
	"sync"
)

type Stage struct {
	tiles              [][]*Tile
	playerMap          map[string]*Player // Only used for updates
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
			outputStage.tiles[y][x] = newTile(area.Tiles[y][x], y, x)
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
	defer stage.playerMutex.Unlock()
	stage.playerMap[player.id] = player
}

func (stage *Stage) removeLockedPlayerById(id string) {
	stage.playerMutex.Lock()
	defer stage.playerMutex.Unlock()
	delete(stage.playerMap, id)
}

func placePlayerOnStageAt(p *Player, stage *Stage, y, x int) {
	if !validCoordinate(y, x, stage) {
		// Extreme - Should be impossible - log error and return?
		log.Fatal("Fatal: Invalid coords to place on stage.")
	}

	p.setStage(stage)
	stage.addLockedPlayer(p)
	stage.tiles[y][x].addPlayerAndNotifyOthers(p)
	spawnItemsFor(p, stage)
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

func (stage *Stage) updateAllWithSound(soundName string) {
	stage.updateAll(soundTriggerByName(soundName))
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

func getOrderedRegion(s *Stage, y, x, height, width int) []*Tile {
	if s == nil || len(s.tiles) == 0 || height <= 0 || width <= 0 {
		return nil
	}

	maxY := len(s.tiles) - 1
	maxX := len(s.tiles[0]) - 1

	y0 := max(0, y)
	x0 := max(0, x)
	y1 := min(maxY, y+height-1)
	x1 := min(maxX, x+width-1)

	capHint := (y1 - y0 + 1) * (x1 - x0 + 1)
	region := make([]*Tile, 0, capHint)

	// Tiles returned in row-major (y-then-x) order for safe locking
	for row := y0; row <= y1; row++ {
		tilesRow := s.tiles[row]
		for col := x0; col <= x1; col++ {
			if t := tilesRow[col]; t != nil {
				region = append(region, t)
			}
		}
	}
	return region
}
