package main

import (
	"sync"
)

type Stage struct {
	tiles       [][]Tile
	players     []*Player // Should this also be 2d array? (Probably no, mutex lock for each movement, although delete mildly faster. Current delete from stage already O(1))
	playerMutex sync.Mutex
	name        string
}

func (stage *Stage) placeOnStage(p *Player) {
	x := p.x
	y := p.y
	stage.tiles[y][x].playerMap[p.id] = p
	stage.players = append(stage.players, p)
	stage.markAllDirty()

}

func (stage *Stage) markAllDirty() {
	for _, player := range stage.players {
		player.viewIsDirty = true
	}
}

func moveNorth(stage *Stage, p *Player) {
	x := p.x
	y := p.y
	nextTile := &stage.tiles[y-1][x]
	if walkable(nextTile) {
		currentTile := &stage.tiles[y][x]
		currentTile.removePlayer(p.id)

		nextTile.addPlayer(p)

		p.y = y - 1
		stage.markAllDirty()
	} else {
		//nop
	}
}

func moveSouth(stage *Stage, p *Player) {
	x := p.x
	y := p.y

	nextTile := &stage.tiles[y+1][x]
	if walkable(nextTile) {
		currentTile := &stage.tiles[y][x]
		currentTile.removePlayer(p.id)

		nextTile.addPlayer(p)

		p.y = y + 1
		stage.markAllDirty()
	} else {
		//nop
	}
}

func moveEast(stage *Stage, p *Player) {
	x := p.x
	y := p.y
	nextTile := &stage.tiles[y][x+1]
	if walkable(nextTile) {
		currentTile := &stage.tiles[y][x]
		currentTile.removePlayer(p.id)

		nextTile.addPlayer(p)

		p.x = x + 1
		stage.markAllDirty()
	} else {
		//nop
	}
}

func moveWest(stage *Stage, p *Player) {
	x := p.x
	y := p.y
	nextTile := &stage.tiles[y][x-1]
	if walkable(nextTile) {
		currentTile := &stage.tiles[y][x]
		currentTile.removePlayer(p.id)

		nextTile.addPlayer(p)

		p.x = x - 1
		stage.markAllDirty()
	} else {
		//nop
	}
}

func (stage *Stage) damageAt(coords [][2]int) {
	for _, pair := range coords {
		for i, player := range stage.players {
			if pair[0] == player.y && pair[1] == player.x {
				player.health += -50
				player.viewIsDirty = true
				if !player.isAlive() {
					deadPlayerTile := &stage.tiles[pair[0]][pair[1]]
					deadPlayerTile.playerMutex.Lock() // break into function, no high level mutexing
					delete(deadPlayerTile.playerMap, player.id)
					deadPlayerTile.playerMutex.Unlock()

					stage.removePlayerAtIndex(i)
				}
			}
		}
	}
}

func (stage *Stage) removePlayerAtIndex(i int) {
	playerMutex.Lock()
	highestIndex := len(stage.players) - 1
	if highestIndex > 0 {
		stage.players[i] = stage.players[highestIndex]
		stage.players = stage.players[:highestIndex]
	} else {
		delete(stageMap, stage.name)
	}
	playerMutex.Unlock()
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

func createBigEmptyStage() Stage {
	return Stage{
		tiles: [][]Tile{
			{newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0)},
			{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
			{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
			{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
			{newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0)},
		},
	}
}

func createStageByName(name string) Stage {
	if name == "greenX" {
		return Stage{
			tiles: [][]Tile{
				{newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(52), newTile(0)},
				{newTile(0), newTile(51), newTile(52), newTile(52), newTile(51), newTile(0)},
				{newTile(0), newTile(52), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0)},
			},
			name: "greenX",
		}
	}
	if name == "big" {
		return Stage{
			tiles: [][]Tile{
				{newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(0), newTile(51), newTile(51), newTile(0), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(0), newTile(51), newTile(51), newTile(0), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0)},
			},
			name: "big",
		}
	}
	if name == "clinic" {
		return Stage{
			tiles: [][]Tile{
				{newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0), newTile(51), newTile(51), newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0), newTile(51), newTile(51), newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0), newTile(0), newTile(0), newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0), newTile(51), newTile(51), newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0), newTile(51), newTile(51), newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0)},
			},
			name: "clinic",
		}
	}
	return createBigEmptyStage()
}
