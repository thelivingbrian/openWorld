package main

import (
	"strconv"
)

type Player struct {
	id          string
	stage       *Stage
	stageName   string
	viewIsDirty bool
	x           int
	y           int
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

func htmlFromColorMatrix(matrix [][]string) string {
	output := ""
	for y := range matrix {
		output += `<div class="grid-row">`
		for x := range matrix[y] {
			var yStr = strconv.Itoa(y)
			var xStr = strconv.Itoa(x)
			output += `<div class="grid-square ` + matrix[y][x] + `" id="c` + yStr + `-` + xStr + `"></div>`
		}
		output += `</div>`
	}
	return output
}

func printStageFor(player *Player) string {
	var output string = `
	<div id="screen" class="grid" hx-swap-oob="true">	
	`

	var tileColors [][]string = make([][]string, len(player.stage.tiles))
	for i, row := range player.stage.tiles {
		tileColors[i] = colorArray(row)
	}

	tileColors[player.y][player.x] = "fusia"
	output += htmlFromColorMatrix(tileColors)

	output += `</div>`
	return output
}
