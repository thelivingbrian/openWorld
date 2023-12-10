package main

import (
	"fmt"
	"strconv"
)

type Player struct {
	id    string
	stage Stage
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
		fmt.Printf("%s\t", tile.color)
	}
	output += `</div>`
	return output
}

func (stage *Stage) printStage() string {
	var output string = `<div class="grid">`
	for y, row := range stage.tiles {
		fmt.Printf("Row %d: ", y)
		output += printRow(row, y)
		fmt.Println()
	}
	output += `</div>`
	return output
}
