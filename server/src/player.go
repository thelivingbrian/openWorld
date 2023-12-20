package main

import (
	"strconv"

	"github.com/gorilla/websocket"
)

type Player struct {
	id        string
	stage     *Stage
	stageName string
	conn      *websocket.Conn
	x         int
	y         int
	actions   *Actions
	health    int
}

type Actions struct {
	space bool
}

func (player *Player) isAlive() bool {
	return player.health > 0
}

func printPageHeaderFor(player *Player) string {
	return `
	<div id="page">
		<div id="controls">      
			<input hx-post="/w" hx-trigger="keydown[key=='w'] from:body" type="hidden" name="token" value="` + player.id + `" />
			<input hx-post="/s" hx-trigger="keydown[key=='s'] from:body" type="hidden" name="token" value="` + player.id + `" />
			<input hx-post="/a" hx-trigger="keydown[key=='a'] from:body" type="hidden" name="token" value="` + player.id + `" />
			<input hx-post="/d" hx-trigger="keydown[key=='d'] from:body" type="hidden" name="token" value="` + player.id + `" />
			<input id="spaceOn" hx-post="/spaceOn" hx-trigger="keydown[key==' '] from:body once" type="hidden" name="token" value="` + player.id + `" />
			<input hx-post="/spaceOff" hx-trigger="keyup[key==' '] from:body" hx-target="#spaceOn" hx-swap="outerHTML" type="hidden" name="token" value="` + player.id + `" />	
			<input id="tick" hx-ext="ws" ws-connect="/screen" ws-send hx-swap="innerHTML" hx-trigger="load once"type="hidden" name="token" value="` + player.id + `" />
		</div>
		<div id="screen" class="grid">
				
		</div>
		<div id="chat" hx-ext="ws" ws-connect="/chat">
			<form id="form" ws-send hx-swap="outerHTML" hx-target="#msg">
				<input type="hidden" name="token" value="` + player.id + `">
				<input id="msg" type="text" name="chat_message" value="">
			</form>
			<div id="chat_room">
				
			</div>
		</div>
	</div>`
}

func placeOnStage(p *Player) {
	x := p.x
	y := p.y
	p.stage.tiles[y][x].playerMap[p.id] = p
	p.stage.playerMap[p.id] = p
	p.stage.markAllDirty()
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

func spaceHighlighter(tile *Tile) string {
	if walkable(tile) {
		return "green"
	} else {
		return "red"
	}
}

func applyHighlights(player *Player, tileColorsPtr *[][]string, relativeCoords [][2]int, highligher func(*Tile) string) {
	absCoordinatePairs := applyRelativeDistance(player.y, player.x, relativeCoords)
	for _, pair := range absCoordinatePairs {
		if pair[0] >= 0 &&
			pair[1] >= 0 &&
			pair[0] < len(player.stage.tiles) &&
			pair[1] < len(player.stage.tiles[0]) {
			tileColors := *tileColorsPtr
			tileColors[pair[0]][pair[1]] = highligher(&player.stage.tiles[pair[0]][pair[1]])
		}
	}
}

func livingView(player *Player) string {
	output := ""

	// Get default colors
	var tileColors [][]string = make([][]string, len(player.stage.tiles))
	for i, row := range player.stage.tiles {
		tileColors[i] = colorArray(row)
	}

	// Add player
	tileColors[player.y][player.x] = "fusia"

	// Add Space
	if player.actions.space {
		applyHighlights(player, &tileColors, cross(), spaceHighlighter)
	}

	output += htmlFromColorMatrix(tileColors)

	return output
}

func updateScreen(player *Player) {
	var output string = `
	<div id="screen" class="grid" hx-swap-oob="true">	
	`

	if player.health > 0 {
		output += livingView(player)
	} else {
		output += `<h2>You Died.</h2>`
		clinic := getClinic()
		player.health = 100
		player.stage = clinic
		player.x = 2
		player.y = 2
		placeOnStage(player)
		output += livingView(player)
	}

	output += `</div>`

	updates <- Update{player, output}
}
