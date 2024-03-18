package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"text/template"
)

const screenTemplate = `
<div id="screen" class="grid">
	{{range $y, $row := .}}
	<div class="grid-row">
		{{range $x, $html := $row}}
		{{$html}}
		{{end}}
	</div>
	{{end}}
</div>`

var parsedScreenTemplate = template.Must(template.New("playerScreen").Parse(screenTemplate))

func htmlFromPlayer(player *Player) []byte {
	var buf bytes.Buffer

	tileHtml := htmlFromTileGrid(player.stage.tiles, player.y, player.x)

	err := parsedScreenTemplate.Execute(&buf, tileHtml)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func htmlFromTileGrid(tiles [][]*Tile, py, px int) [][]string {
	output := make([][]string, len(tiles))
	for y := range output {
		output[y] = make([]string, len(tiles[y]))
		for x := range output[y] {
			if x == px && y == py {
				output[y][x] = htmlForPlayerTile(tiles[y][x])
				continue
			}
			output[y][x] = htmlForTile(tiles[y][x])
		}
	}
	return output
}

func spaceHighlighter(tile *Tile) string {
	if len(tile.playerMap) > 0 {
		return "half-trsp dark-blue"
	} else if walkable(tile) {
		return "half-trsp salmon"
	} else {
		return "half-trsp salmon"
	}
}

func shiftHighlighter(tile *Tile) string {
	if walkable(tile) {
		return "red trsp20"
	}
	return ""
}

func randomFieryColor() string {
	randN := rand.Intn(4)
	if randN < 1 {
		return "yellow"
	}
	if randN < 2 {
		return "orange"
	}
	return "red"
}

func printPageFor(player *Player) string {
	return `
	<div id="page" hx-swap-oob="true">
		<div id="controls" hx-ext="ws" ws-connect="/screen">
			<input id="token" type="hidden" name="token" value="` + player.id + `" />
			<input id="w" type="hidden" ws-send hx-trigger="keydown[key=='w'||key=='ArrowUp'] from:body" hx-include="#token" name="keypress" value="w" />
			<input id="a" type="hidden" ws-send hx-trigger="keydown[key=='a'||key=='ArrowLeft'] from:body" hx-include="#token" name="keypress" value="a" />
			<input id="s" type="hidden" ws-send hx-trigger="keydown[key=='s'||key=='ArrowDown'] from:body" hx-include="#token" name="keypress" value="s" />
			<input id="d" type="hidden" ws-send hx-trigger="keydown[key=='d'||key=='ArrowRight'] from:body" hx-include="#token" name="keypress" value="d" />
			<input id="w" type="hidden" ws-send hx-trigger="keydown[key=='W'] from:body" hx-include="#token" name="keypress" value="W" />
			<input id="a" type="hidden" ws-send hx-trigger="keydown[key=='A'] from:body" hx-include="#token" name="keypress" value="A" />
			<input id="s" type="hidden" ws-send hx-trigger="keydown[key=='S'] from:body" hx-include="#token" name="keypress" value="S" />
			<input id="d" type="hidden" ws-send hx-trigger="keydown[key=='D'] from:body" hx-include="#token" name="keypress" value="D" />
			<input id="f" type="hidden" ws-send hx-trigger="keydown[key=='f'] from:body" hx-include="#token" name="keypress" value="f" />
			<input id="g" type="hidden" ws-send hx-trigger="keydown[key=='g'] from:body" hx-include="#token" name="keypress" value="g" />
			<input id="space-on" type="hidden" ws-send hx-trigger="keydown[key==' '] from:body once" hx-include="#token" name="keypress" value="Space-On" />
			<input id="space-off" type="hidden" ws-send hx-trigger="keyup[key==' '] from:body" hx-include="#token" name="keypress" value="Space-Off" />
			<input hx-post="/clear" hx-target="#screen" hx-swap="outerHTML" hx-trigger="keydown[key=='0'] from:body" type="hidden" name="token" value="` + player.id + `" />
			<input id="tick" ws-send hx-trigger="load once" type="hidden" name="token" value="` + player.id + `" />
			<div id="script"></div>
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
	hearts := getHeartsFromHealth(player.health)
	return fmt.Sprintf(`%s %s<br /><span class="red">Streak %d</span> | <span class="blue">^ %d</span>  | <span class="dark-green">$ %d</span>`, player.username, hearts, 0, player.actions.boostCounter, player.money)
}

func getHeartsFromHealth(i int) string {
	i = (i - (i % 50)) / 50
	if i == 0 {
		return ""
	}
	if i == 1 {
		return "❤️"
	}
	if i == 2 {
		return "❤️❤️"
	}
	if i == 3 {
		return "❤️❤️❤️"
	}
	return fmt.Sprintf("❤️x%d", i)
}

func htmlForTile(tile *Tile) string {
	svgtag := svgFromTile(tile)
	return fmt.Sprintf(tile.htmlTemplate, playerBox(tile), svgtag)
}

func playerBox(tile *Tile) string {
	playerIndicator := ""
	if p := tile.getAPlayer(); p != nil {
		playerIndicator = cssClassFromHealth(p)
	}
	return fmt.Sprintf(`<div id="p%d-%d" class="box zp %s" id=""></div>`, tile.y, tile.x, playerIndicator)
}

func htmlForPlayerTile(tile *Tile) string {
	svgtag := svgFromTile(tile)
	return fmt.Sprintf(tile.htmlTemplate, fusiaBox(tile.y, tile.x), svgtag)
}

func fusiaBox(y, x int) string {
	return fmt.Sprintf(`<div id="p%d-%d" class="box zp fusia r0" id=""></div>`, y, x)
}

// Create slice of proper size? Currently has many null entries
func highlightBoxesForPlayer(player *Player, tiles []*Tile) string {
	highlights := ""

	// Still risk here of concurrent read/write?
	for _, tile := range tiles {
		if tile == nil {
			continue
		}
		if tile.stage != player.stage {
			continue
		}

		// shiftHighlights not needed, want generic highlight option
		_, impactsHud := player.actions.shiftHighlights[tile]
		if impactsHud && player.actions.boostCounter > 0 {
			highlights += oobHighlightBox(tile, shiftHighlighter(tile))
			//continue
		}
		_, impactsHud = player.actions.spaceHighlights[tile]
		if impactsHud {
			highlights += oobHighlightBox(tile, spaceHighlighter(tile))
			continue
		}

		highlights += oobHighlightBox(tile, "")
	}

	return highlights
}

func oobHighlightBox(tile *Tile, cssClass string) string {
	template := `<div id="t%d-%d" class="box top %s"></div>`
	return fmt.Sprintf(template, tile.y, tile.x, cssClass)
}

func svgFromTile(tile *Tile) string {
	svgtag := `<div id="%s" class="box zS">`
	if tile.powerUp != nil || tile.money != 0 || tile.boosts != 0 {
		svgtag += `<svg width="30" height="30">`
		if tile.powerUp != nil {
			svgtag += `<circle class="svgRed" cx="10" cy="10" r="10" />`
		}
		if tile.money != 0 {
			svgtag += `<circle class="svgGreen" cx="10" cy="20" r="10" />`
		}
		if tile.boosts != 0 {
			svgtag += `<circle class="svgBlue" cx="20" cy="20" r="10" />`
		}
		svgtag += `</svg>`
	}
	svgtag += "</div>"
	sId := fmt.Sprintf("s%d-%d", tile.y, tile.x)
	return fmt.Sprintf(svgtag, sId)
}
