package main

import (
	"bytes"
	"fmt"
	"text/template"
)

const screenTemplate = `
<div id="screen" class="grid">
	{{range $y, $row := .}}
	<div class="grid-row">
		{{range $x, $color := $row}}
		<div class="grid-square {{$color}}" id="c{{$y}}-{{$x}}"></div>
		{{end}}
	</div>
	{{end}}
</div>`

var parsedScreenTemplate = template.Must(template.New("playerScreen").Parse(screenTemplate))

func htmlFromPlayer(player *Player) []byte {
	var buf bytes.Buffer
	tileColors := tilesToColors(player.stage.tiles)
	playerView(player, tileColors)

	err := parsedScreenTemplate.Execute(&buf, tileColors)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func htmlFromStage(stage *Stage) string {
	var buf bytes.Buffer
	tileColors := tilesToColors(stage.tiles)

	err := parsedScreenTemplate.Execute(&buf, tileColors)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

func playerView(player *Player, tileColors [][]string) {
	tileColors[player.y][player.x] = "fusia"
	if player.actions.space {
		applyHighlights(player, tileColors, grid5x5, spaceHighlighter)
	}
}

func hudAsOutOfBound(player *Player) string {
	highlights := ""
	if player.actions.space {
		highlights += highlightsAsOob(player, grid5x5, spaceHighlighter)
	}

	playerIcon := fmt.Sprintf(`<div class="grid-square fusia" id="c%d-%d" hx-swap-oob="true"></div>`, player.y, player.x)

	return highlights + playerIcon
}

func tilesToColors(tiles [][]*Tile) [][]string {
	output := make([][]string, len(tiles))
	for y := range output {
		output[y] = make([]string, len(tiles[y]))
		for x := range output[y] {
			output[y][x] = tiles[y][x].currentCssClass
		}
	}
	return output
}

func applyHighlights(player *Player, tileColors [][]string, relativeCoords [][2]int, highligher func(*Tile) string) {
	absCoordinatePairs := applyRelativeDistance(player.y, player.x, relativeCoords)
	for _, pair := range absCoordinatePairs {
		if pair[0] >= 0 &&
			pair[1] >= 0 &&
			pair[0] < len(player.stage.tiles) &&
			pair[1] < len(player.stage.tiles[0]) {
			tileColors[pair[0]][pair[1]] = highligher(player.stage.tiles[pair[0]][pair[1]])
		}
	}
}

func highlightsAsOob(player *Player, relativeCoords [][2]int, highligher func(*Tile) string) string {
	output := ``
	absCoordinatePairs := applyRelativeDistance(player.y, player.x, relativeCoords)
	for _, pair := range absCoordinatePairs {
		if pair[0] >= 0 &&
			pair[1] >= 0 &&
			pair[0] < len(player.stage.tiles) &&
			pair[1] < len(player.stage.tiles[0]) {
			highlight := highligher(player.stage.tiles[pair[0]][pair[1]])
			output += fmt.Sprintf(`<div class="grid-square %s" id="c%d-%d" hx-swap-oob="true"></div>`, highlight, pair[0], pair[1])
		}
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

func printPageFor(player *Player) string {
	return `
	<div id="page">
		<div id="controls">      
			<input hx-post="/w" hx-trigger="keydown[key=='w'] from:body" type="hidden" name="token" value="` + player.id + `" />
			<input hx-post="/s" hx-trigger="keydown[key=='s'] from:body" type="hidden" name="token" value="` + player.id + `" />
			<input hx-post="/a" hx-trigger="keydown[key=='a'] from:body" type="hidden" name="token" value="` + player.id + `" />
			<input hx-post="/d" hx-trigger="keydown[key=='d'] from:body" type="hidden" name="token" value="` + player.id + `" />
			<input hx-post="/clear" hx-target="#screen" hx-swap="outerHTML" hx-trigger="keydown[key=='0'] from:body" type="hidden" name="token" value="` + player.id + `" />
			<input id="spaceOn" hx-post="/spaceOn" hx-trigger="keydown[key==' '] from:body once" type="hidden" name="token" value="` + player.id + `" />
			<input hx-post="/spaceOff" hx-trigger="keyup[key==' '] from:body" hx-target="#spaceOn" hx-swap="outerHTML" type="hidden" name="token" value="` + player.id + `" />	
			<input id="tick" hx-ext="ws" ws-connect="/screen" ws-send hx-trigger="load once" type="hidden" name="token" value="` + player.id + `" />
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
