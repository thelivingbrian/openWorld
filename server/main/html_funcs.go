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

func spaceHighlighter(_ *Tile) string {
	/*
		This has bugs because it doesn't update on movement of the other player, only the highlight viewer
		if len(tile.playerMap) > 0 {
			return "half-trsp dark-blue"
		}
	*/
	//if walkable(tile) {
	//	return "half-trsp salmon" // vs "" to show no effect
	//}
	return "half-trsp salmon"
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
		<div id="main_view">
			` + divPlayerInformation(player) + `
			<div id="screen" class="grid">
					
			</div>
			<div id="bottom_text">
				&nbsp;&nbsp;> Press 'm' for Menu.
			</div>
		</div>
		<div id="controls" hx-ext="ws" ws-connect="/screen">
			<input id="token" type="hidden" name="token" value="` + player.id + `" />
			<input hx-post="/clear" hx-target="#screen" hx-swap="outerHTML" hx-trigger="keydown[key=='0'] from:body" type="hidden" />
			<input id="tick" ws-send hx-trigger="load once" type="hidden" name="token" value="` + player.id + `" />
			<div id="modal_background">
				
			</div>
			` + divInput() + `
			<div id="script"></div>
		</div>
	</div>`
}

func divPlayerInformation(player *Player) string {
	return `
	<div id="info" hx-swap-oob="true">
		<b>` + playerInformation(player) + `</b>
	</div>`
}

func divModalDisabled() string {
	return `
	<div id="modal_background">
		
	</div>
	`
}

func divInput() string {
	return `
	<div id="input">
		<div id="input-desktop">
			<input id="wKey" type="hidden" ws-send hx-trigger="keydown[key=='w'||key=='ArrowUp'] from:body" hx-include="#token" name="eventname" value="w" />
			<input id="aKey" type="hidden" ws-send hx-trigger="keydown[key=='a'||key=='ArrowLeft'] from:body" hx-include="#token" name="eventname" value="a" />
			<input id="sKey" type="hidden" ws-send hx-trigger="keydown[key=='s'||key=='ArrowDown'] from:body" hx-include="#token" name="eventname" value="s" />
			<input id="dKey" type="hidden" ws-send hx-trigger="keydown[key=='d'||key=='ArrowRight'] from:body" hx-include="#token" name="eventname" value="d" />
			<input id="wShift" type="hidden" ws-send hx-trigger="keydown[key=='W'] from:body" hx-include="#token" name="eventname" value="W" />
			<input id="aShift" type="hidden" ws-send hx-trigger="keydown[key=='A'] from:body" hx-include="#token" name="eventname" value="A" />
			<input id="sShift" type="hidden" ws-send hx-trigger="keydown[key=='S'] from:body" hx-include="#token" name="eventname" value="S" />
			<input id="dShift" type="hidden" ws-send hx-trigger="keydown[key=='D'] from:body" hx-include="#token" name="eventname" value="D" />
			<input id="fKey" type="hidden" ws-send hx-trigger="keydown[key=='f'] from:body" hx-include="#token" name="eventname" value="f" />
			<input id="gKey" type="hidden" ws-send hx-trigger="keydown[key=='g'] from:body" hx-include="#token" name="eventname" value="g" />
			<input id="menuOnKey" type="hidden" ws-send hx-trigger="keydown[key=='m'||key=='M'||key=='Escape'] from:body" hx-include="#token" name="eventname" value="menuOn" />
			<input id="space-onKey" type="hidden" ws-send hx-trigger="keydown[key==' '] from:body" hx-include="#token" name="eventname" value="Space-On" />
		</div>

		<input id="w" type="hidden" ws-send hx-trigger="click from:#but-w" hx-include="#token" name="eventname" value="w" />
		<input id="a" type="hidden" ws-send hx-trigger="click from:#but-a" hx-include="#token" name="eventname" value="a" />
		<input id="s" type="hidden" ws-send hx-trigger="click from:#but-s" hx-include="#token" name="eventname" value="s" />
		<input id="d" type="hidden" ws-send hx-trigger="click from:#but-d" hx-include="#token" name="eventname" value="d" />

		<input id="menuOn" type="hidden" ws-send hx-trigger="click from:#but-m" hx-include="#token" name="eventname" value="menuOn" />
		<input id="space-on" type="hidden" ws-send hx-trigger="click from:#but-space" hx-include="#token" name="eventname" value="Space-On" />
		<input id="shift-on" type="hidden" ws-send hx-trigger="click from:#but-shift-on" hx-include="#token" name="eventname" value="Shift-On" />

		<div class="container">
			<div id="dpad" class="dpad-container">
				<button id="but-w" class="button up">Up</button>
				<button id="but-a" class="button left">Left</button>
				<button class="button middle"></button>
				<button id="but-d" class="button right">Right</button>
				<button id="but-s" class="button down">Down</button>
			</div>
			<div class="center-container">
				<button id="but-m" class="half-button">menu</button>
			</div>
			<div class="a-b-container">
				<button id="but-space" class="button A">Space</button>
				<button id="but-shift-on" class="button B">Shift</button>
			</div>
		</div>
	
	</div>
`
}

func divInputShift() string {
	return `
	<div id="input">
		<div id="input-desktop">
			<input id="wKey" type="hidden" ws-send hx-trigger="keydown[key=='w'||key=='ArrowUp'] from:body" hx-include="#token" name="eventname" value="w" />
			<input id="aKey" type="hidden" ws-send hx-trigger="keydown[key=='a'||key=='ArrowLeft'] from:body" hx-include="#token" name="eventname" value="a" />
			<input id="sKey" type="hidden" ws-send hx-trigger="keydown[key=='s'||key=='ArrowDown'] from:body" hx-include="#token" name="eventname" value="s" />
			<input id="dKey" type="hidden" ws-send hx-trigger="keydown[key=='d'||key=='ArrowRight'] from:body" hx-include="#token" name="eventname" value="d" />
			<input id="wShift" type="hidden" ws-send hx-trigger="keydown[key=='W'] from:body" hx-include="#token" name="eventname" value="W" />
			<input id="aShift" type="hidden" ws-send hx-trigger="keydown[key=='A'] from:body" hx-include="#token" name="eventname" value="A" />
			<input id="sShift" type="hidden" ws-send hx-trigger="keydown[key=='S'] from:body" hx-include="#token" name="eventname" value="S" />
			<input id="dShift" type="hidden" ws-send hx-trigger="keydown[key=='D'] from:body" hx-include="#token" name="eventname" value="D" />
			<input id="fKey" type="hidden" ws-send hx-trigger="keydown[key=='f'] from:body" hx-include="#token" name="eventname" value="f" />
			<input id="gKey" type="hidden" ws-send hx-trigger="keydown[key=='g'] from:body" hx-include="#token" name="eventname" value="g" />
			<input id="menuOnKey" type="hidden" ws-send hx-trigger="keydown[key=='m'||key=='M'||key=='Escape'] from:body" hx-include="#token" name="eventname" value="menuOn" />
			<input id="space-onKey" type="hidden" ws-send hx-trigger="keydown[key==' '] from:body" hx-include="#token" name="eventname" value="Space-On" />
		</div>

		<input id="wShift" type="hidden" ws-send hx-trigger="click from:#but-w" hx-include="#token" name="eventname" value="W" />
		<input id="aShift" type="hidden" ws-send hx-trigger="click from:#but-a" hx-include="#token" name="eventname" value="A" />
		<input id="sShift" type="hidden" ws-send hx-trigger="click from:#but-s" hx-include="#token" name="eventname" value="S" />
		<input id="dShift" type="hidden" ws-send hx-trigger="click from:#but-d" hx-include="#token" name="eventname" value="D" />

		<input id="menuOn" type="hidden" ws-send hx-trigger="click from:#but-m" hx-include="#token" name="eventname" value="menuOn" />
		<input id="space-on" type="hidden" ws-send hx-trigger="click from:#but-space" hx-include="#token" name="eventname" value="Space-On" />
		<input id="shift-off" type="hidden" ws-send hx-trigger="click from:#but-shift-off" hx-include="#token" name="eventname" value="Shift-Off" />
	
		<div class="container">
			<div id="dpad" class="dpad-container">
				<button id="but-w" class="button up">UP</button>
				<button id="but-a" class="button left">LEFT</button>
				<button class="button middle"></button>
				<button id="but-d" class="button right">RIGHT</button>
				<button id="but-s" class="button down">DOWN</button>
			</div>
			<div class="center-container">
				<button id="but-m" class="half-button">menu</button>
			</div>
			<div class="a-b-container">
				<button id="but-space" class="button A">Space</button>
				<button id="but-shift-off" class="button B">Shift</button>
			</div>
		</div>
	
	</div
`
}

func divInputDisabled() string {
	return `
	<div id="input">

	</div>
`
}

func playerInformation(player *Player) string {
	// mutexing?
	hearts := getHeartsFromHealth(player.health)
	return fmt.Sprintf(`%s %s<br /><span class="red">Streak %d</span> | <span class="blue">^ %d</span>  | <span class="dark-green">$ %d</span>`, player.username, hearts, player.getKillStreakSync(), player.actions.boostCounter, player.money)
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
	// grab tile y and x only once here or in parent method?
	return fmt.Sprintf(tile.htmlTemplate, playerBox(tile), emptyUserBox(tile.y, tile.x), interactableBox(tile), svgtag)
}

func htmlForPlayerTile(tile *Tile) string {
	svgtag := svgFromTile(tile)
	return fmt.Sprintf(tile.htmlTemplate, playerBox(tile), fusiaBox(tile.y, tile.x), interactableBox(tile), svgtag)
}

func fusiaBox(y, x int) string {
	return fmt.Sprintf(`<div id="u%d-%d" class="box zu fusia r0"></div>`, y, x)
}

func playerBox(tile *Tile) string {
	playerIndicator := ""
	if p := tile.getAPlayer(); p != nil {
		playerIndicator = cssClassFromHealth(p)
	}
	return fmt.Sprintf(`<div id="p%d-%d" class="box zp %s"></div>`, tile.y, tile.x, playerIndicator)
}

func interactableBox(tile *Tile) string {
	indicator := ""
	if tile.interactable != nil {
		indicator = tile.interactable.cssClass
	}
	return fmt.Sprintf(`<div id="i%d-%d" class="box zi %s"></div>`, tile.y, tile.x, indicator)
}

func emptyUserBox(y, x int) string {
	return fmt.Sprintf(`<div id="u%d-%d" class="box zu"></div>`, y, x)
}

// Create slice of proper size? Currently has many null entries
func highlightBoxesForPlayer(player *Player, tiles []*Tile) string {
	highlights := ""

	// Still risk here of concurrent read/write?
	for _, tile := range tiles {
		if tile == nil {
			fmt.Println(".") // seems to match number of actual highlights to add
			continue
		}
		if tile.stage != player.stage {
			continue
		}

		_, impactsHud := player.actions.spaceHighlights[tile]
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
