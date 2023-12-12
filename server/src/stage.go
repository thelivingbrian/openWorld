package main

import (
	"strconv"
)

type Tile struct {
	color string
}

type Stage struct {
	tiles   [][]Tile
	Players []*Player
}

func printRow(row []Tile, y int) string {
	var output string = `<div class="grid-row">`
	for x, tile := range row {
		var yStr = strconv.Itoa(y)
		var xStr = strconv.Itoa(x)
		output += `<div class="grid-square ` + tile.color + `" id="c` + yStr + `-` + xStr + `"></div>`
	}
	output += `</div>`
	return output
}

func printPageHeaderFor(player *Player) string {
	return `
	<div id="page">
    <div id="controls">      
		<input hx-post="/w" hx-trigger="keyup[key=='w'] from:body" type="hidden" name="token" value="` + player.id + `" />
		<input hx-post="/s" hx-trigger="keyup[key=='s'] from:body" type="hidden" name="token" value="` + player.id + `" />
		<input hx-post="/a" hx-trigger="keyup[key=='a'] from:body" type="hidden" name="token" value="` + player.id + `" />
		<input hx-post="/d" hx-trigger="keyup[key=='d'] from:body" type="hidden" name="token" value="` + player.id + `" />	
		<input id="tick" hx-post="/screen" hx-trigger="every 20ms" hx-target="#tick" hx-swap="innerHTML" type="hidden" name="token" value="` + player.id + `" />
	</div>
    <div id="screen" class="grid">
			
	</div>
	</div>`
}

func (stage *Stage) printStageFor(player *Player) string {
	// <input id="tick" hx-post="/screen" hx-trigger="load" hx-target="#tick" hx-swap="outerHTML" type="hidden" name="token" value="` + player.id + `" />
	var output string = `
	<div id="screen" class="grid" hx-swap-oob="true">	
	`
	for y, row := range stage.tiles {
		output += printRow(row, y)
	}
	output += `</div>`
	return output
}

func (stage *Stage) placeOnStage(p *Player) {
	x := p.x
	y := p.y
	stage.tiles[y][x] = Tile{"fusia"}
	stage.Players = append(stage.Players, p)
}

func (stage *Stage) markAllDirty() {
	for _, player := range stage.Players {
		player.viewIsDirty = true
	}
}

func walkable(tile *Tile) bool {
	return tile.color != "half-gray"
}

func moveNorth(stage *Stage, p *Player) {
	x := p.x
	y := p.y
	nextTile := &stage.tiles[y-1][x]
	if walkable(nextTile) {
		stage.tiles[y][x] = Tile{""}
		*nextTile = Tile{"fusia"}
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
		stage.tiles[y][x] = Tile{""}
		*nextTile = Tile{"fusia"}
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
		stage.tiles[y][x] = Tile{""}
		*nextTile = Tile{"fusia"}
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
		stage.tiles[y][x] = Tile{""}
		*nextTile = Tile{"fusia"}
		p.x = x - 1
		stage.markAllDirty()
	} else {
		//nop
	}
}

func getBigEmptyStage() Stage {
	return Stage{
		tiles: [][]Tile{
			{Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}},
			{Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{"half-gray"}},
			{Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{"half-gray"}},
			{Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{"half-gray"}},
			{Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}},
		},
	}
}

func getStageByName(name string) Stage {
	if name == "greenX" {
		return Stage{
			tiles: [][]Tile{
				{Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}},
				{Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{""}, Tile{"green"}, Tile{"half-gray"}},
				{Tile{"half-gray"}, Tile{""}, Tile{"green"}, Tile{"green"}, Tile{""}, Tile{"half-gray"}},
				{Tile{"half-gray"}, Tile{"green"}, Tile{""}, Tile{""}, Tile{""}, Tile{"half-gray"}},
				{Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}},
			},
		}
	}
	if name == "big" {
		return Stage{
			tiles: [][]Tile{
				{Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}},
				{Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{"half-gray"}},
				{Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{"half-gray"}},
				{Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{"half-gray"}},
				{Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{"half-gray"}},
				{Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{"half-gray"}},
				{Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{"half-gray"}},
				{Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{"half-gray"}},
				{Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{"half-gray"}},
				{Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}},
			},
		}
	}
	return getBigEmptyStage()
}
