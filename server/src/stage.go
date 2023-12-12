package main

import (
	"sync"
)

type Tile struct {
	material    int
	playerMap   map[string]*Player
	playerMutex sync.Mutex // Currently unused. Probably important?
	// Items and coords?
}

type Stage struct {
	tiles   [][]Tile
	players []*Player
}

func colorOf(tile *Tile) string {
	if len(tile.playerMap) > 0 {
		return "blue"
	}
	if tile.material == 0 {
		return "half-gray"
	}
	return ""
}

func colorArray(row []Tile) []string {
	var output []string = make([]string, len(row))
	for i := range row {
		output[i] = colorOf(&row[i])
	}
	return output
}

func newTile(mat int) Tile {
	return Tile{mat, make(map[string]*Player), sync.Mutex{}}
}

func (stage *Stage) placeOnStage(p *Player) {
	x := p.x
	y := p.y
	stage.tiles[y][x].playerMap[p.id] = p
	stage.players = append(stage.players, p)
}

func (stage *Stage) markAllDirty() {
	for _, player := range stage.players {
		player.viewIsDirty = true
	}
}

func walkable(tile *Tile) bool {
	return tile.material > 50
}

func moveNorth(stage *Stage, p *Player) {
	x := p.x
	y := p.y
	nextTile := &stage.tiles[y-1][x]
	if walkable(nextTile) {
		delete(stage.tiles[y][x].playerMap, p.id)
		nextTile.playerMap[p.id] = p
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
		delete(stage.tiles[y][x].playerMap, p.id)
		nextTile.playerMap[p.id] = p
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
		delete(stage.tiles[y][x].playerMap, p.id)
		nextTile.playerMap[p.id] = p
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
		delete(stage.tiles[y][x].playerMap, p.id)
		nextTile.playerMap[p.id] = p
		p.x = x - 1
		stage.markAllDirty()
	} else {
		//nop
	}
}

func getBigEmptyStage() Stage {
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

func getStageByName(name string) Stage {
	if name == "greenX" {
		return Stage{
			tiles: [][]Tile{
				{newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0)},
				{newTile(0), newTile(51), newTile(51), newTile(51), newTile(52), newTile(0)},
				{newTile(0), newTile(51), newTile(52), newTile(52), newTile(51), newTile(0)},
				{newTile(0), newTile(52), newTile(51), newTile(51), newTile(51), newTile(0)},
				{newTile(0), newTile(0), newTile(0), newTile(0), newTile(0), newTile(0)},
			},
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
		}
	}
	return getBigEmptyStage()
}
