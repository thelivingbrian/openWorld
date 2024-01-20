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
		applyHighlights(player, tileColors, player.actions.spaceShape, spaceHighlighter) // (Is this actually possible?)
	}
}

func hudAsOutOfBound(player *Player) string {
	highlights := ""
	if player.actions.space {
		// Any risk here of concurrent read/write?
		for tile := range player.actions.spaceHighlights {
			highlights += oobColoredTile(tile, spaceHighlighter(tile))
		}
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
	if len(tile.playerMap) > 0 {
		return "dark-blue"
	} else if walkable(tile) {
		return "green"
	} else {
		return "red"
	}
}

func printPageFor(player *Player) string {
	return `
	<div id="page" hx-swap-oob="true">
		<div id="controls" hx-ext="ws" ws-connect="/screen">
			<input id="token" type="hidden" name="token" value="` + player.id + `" />
			<input id="w" type="hidden" ws-send hx-trigger="keydown[key=='w'||key=='W'||key=='ArrowUp'] from:body" hx-include="#token" name="keypress" value="W" />
			<input id="a" type="hidden" ws-send hx-trigger="keydown[key=='a'||key=='A'||key=='ArrowLeft'] from:body" hx-include="#token" name="keypress" value="A" />
			<input id="s" type="hidden" ws-send hx-trigger="keydown[key=='s'||key=='S'||key=='ArrowDown'] from:body" hx-include="#token" name="keypress" value="S" />
			<input id="d" type="hidden" ws-send hx-trigger="keydown[key=='d'||key=='D'||key=='ArrowRight'] from:body" hx-include="#token" name="keypress" value="D" />
			<input id="space-on" type="hidden" ws-send hx-trigger="keydown[key==' '] from:body once" hx-include="#token" name="keypress" value="Space-On" />
			<input id="space-off" type="hidden" ws-send hx-trigger="keyup[key==' '] from:body" hx-include="#token" name="keypress" value="Space-Off" />
			<input hx-post="/clear" hx-target="#screen" hx-swap="outerHTML" hx-trigger="keydown[key=='0'] from:body" type="hidden" name="token" value="` + player.id + `" />
			<input id="tick" ws-send hx-trigger="load once" type="hidden" name="token" value="` + player.id + `" />
		</div>
		<div id="info">
			<b>` + playerInformation(player) + `</b>
		</div>
		<div id="screen" class="grid">
				
		</div>
	</div>`
}

func divPlayerInformation(player *Player) string {
	return `
	<div id="info" hx-swap-oob="true">
		<b>` + playerInformation(player) + `</b>
	</div>`
}

func playerInformation(player *Player) string {
	return fmt.Sprintf("%s | Health %d Money %d", player.username, player.health, player.money)
}

func htmlFromTile(tile *Tile) string {
	return fmt.Sprintf(`<div class="grid-square %s" id="c%d-%d" hx-swap-oob="true"  ></div>`, tile.currentCssClass, tile.y, tile.x)
}

func oobColoredTile(tile *Tile, cssClass string) string {
	return fmt.Sprintf(`<div class="grid-square %s" id="c%d-%d" hx-swap-oob="true"></div>`, cssClass, tile.y, tile.x)
}
