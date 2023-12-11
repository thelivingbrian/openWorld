package main

import (
	"strconv"
)

type Player struct {
	id        string
	stage     *Stage
	stageName string
	x         int
	y         int
}

type Tile struct {
	color string
}

type Stage struct {
	tiles [][]Tile
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

func (stage *Stage) printStageFor(player *Player) string {
	var output string = `
	<div class="grid" id="screen">
		<input hx-post="/w" hx-trigger="keyup[key=='w'] from:body" hx-target="#screen" hx-swap="outerHTML" type="hidden" name="token" value="` + player.id + `" />
		<input hx-post="/s" hx-trigger="keyup[key=='s'] from:body" hx-target="#screen" hx-swap="outerHTML" type="hidden" name="token" value="` + player.id + `" />
		<input hx-post="/a" hx-trigger="keyup[key=='a'] from:body" hx-target="#screen" hx-swap="outerHTML" type="hidden" name="token" value="` + player.id + `" />
		<input hx-post="/d" hx-trigger="keyup[key=='d'] from:body" hx-target="#screen" hx-swap="outerHTML" type="hidden" name="token" value="` + player.id + `" />
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
