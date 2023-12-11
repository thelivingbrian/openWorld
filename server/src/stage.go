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
		//fmt.Printf("%s\t", tile.color)
	}
	output += `</div>`
	return output
}

func (stage *Stage) printStageFor(player *Player) string {
	var output string = `
	<div class="grid" id="screen">
		<form hx-post="/userInput" hx-trigger="keyup[key=='w'] from:body" hx-target="#screen" hx-swap="outerHTML">
			<input type="hidden" name="token" value="` + player.id + `" />
			<button>Click Me!</button>
		</form>`
	for y, row := range stage.tiles {
		//fmt.Printf("Row %d: ", y)
		output += printRow(row, y)
		//fmt.Println()
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
	return true
}

func moveNorth(stage *Stage, p *Player) {
	x := p.x
	y := p.y
	currentTile := &stage.tiles[y][x]
	if walkable(currentTile) {
		stage.tiles[y][x] = Tile{""}
		stage.tiles[y-1][x] = Tile{"fusia"}
		p.y = y - 1
		p.x = x
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
	return getBigEmptyStage()
}
