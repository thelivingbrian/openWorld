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

	tileHtml := htmlFromTileGrid(player.stage.tiles, player.y, player.x, player.actions.spaceHighlights)

	err := parsedScreenTemplate.Execute(&buf, tileHtml)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func htmlFromTileGrid(tiles [][]*Tile, py, px int, highlights map[*Tile]bool) [][]string {
	output := make([][]string, len(tiles))
	for y := range output {
		output[y] = make([]string, len(tiles[y]))
		for x := range output[y] {
			highlightColor := ""
			_, found := highlights[tiles[y][x]]
			if found {
				highlightColor = spaceHighlighter()
			}
			output[y][x] = htmlForTile(tiles[y][x], highlightColor)
		}
	}
	return output
}

func spaceHighlighter() string {
	// constant instead.
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

/*
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

			<div id="modal_background">

			</div>
			` + divInput() + `
			<div id="script"></div>
		</div>
	</div>`
}
*/

func chooseYourColor() string {
	return `
	<div id="page" hx-swap-oob="true">
	
		<div id="main_view">
			
			<div id="info" hx-swap-oob="true">
				 <form hx-post="/new" hx-target="#bottom_text">
					<b>New Player</b>
					
					<div class="form-group color-selection">
						<label id="color-window-0">
							<input type="radio" name="player-team" value="fuchsia" checked />
							<div id="exampleSquare-0">
								<div class="grid-square-example fuchsia"></div>
							</div>
						</label>

						<label id="color-window-1">
							<input type="radio" name="player-team" value="sky-blue" />
							<div id="exampleSquare-1">
								<div class="grid-square-example sky-blue"></div>
							</div>
						</label>	
					
					</div>

					<div class="form-group">
						<label class="left-float">Username:</label>
						<input type="text" name="player-name" />
					</div>

					<div class="form-group" style="justify-content: center;">
						<input type="submit" value="Go">
					</div>
				</form>
			</div>
			<div id="bottom_text">
			</div>
		</div>
	
	</div>
	`
}

func divBottomInvalid(s string) string {
	return `
	<div id="bottom_text" hx-swap-oob="true">
		<p style='color:red'>	
			` + s + `  
		</p>
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
	<div id="x0-0" class="container">
	</div>
	<div id="x0-1" class="container hidden">
	</div>
`
}

func divInputShift() string {
	return `
	<div id="x0-0" class="container hidden">
	</div>
	<div id="x0-1" class="container">
	</div>
`
}

func divInputDisabled() string {
	return `
	<div id="input">

	</div>
`
}

func playerInformation(player *Player) string {
	hearts := getHeartsFromHealth(player.getHealthSync())
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

func htmlForTile(tile *Tile, highlight string) string {
	svgtag := svgFromTile(tile)
	// grab tile y and x only once here or in parent method?
	return fmt.Sprintf(tile.htmlTemplate, playerBox(tile), interactableBox(tile), svgtag, emptyWeatherBox(tile.y, tile.x), oobHighlightBox(tile, highlight))
}

func playerBoxSpecifc(y, x int, icon string) string {
	return fmt.Sprintf(`<div id="p%d-%d" class="box zp %s"></div>`, y, x, icon)
}

func playerBox(tile *Tile) string {
	playerIndicator := ""
	if p := tile.getAPlayer(); p != nil {
		playerIndicator = p.getIconSync()
	}
	return fmt.Sprintf(`<div id="p%d-%d" class="box zp %s"></div>`, tile.y, tile.x, playerIndicator)
}

func interactableBox(tile *Tile) string {
	indicator := ""
	//mutex
	if tile.interactable != nil {
		indicator = tile.interactable.cssClass
	}
	return fmt.Sprintf(`<div id="i%d-%d" class="box zi %s"></div>`, tile.y, tile.x, indicator)
}

func emptyWeatherBox(y, x int) string {
	//  blue trsp20 for gloom
	return fmt.Sprintf(`<div id="w%d-%d" class="box zw"></div>`, y, x)
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
			highlights += oobHighlightBox(tile, spaceHighlighter())
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

func weatherBox(tile *Tile, cssClass string) string {
	template := `<div id="w%d-%d" class="box zw %s"></div>`
	return fmt.Sprintf(template, tile.y, tile.x, cssClass)
}

func svgFromTile(tile *Tile) string {
	svgtag := `<div id="%s" class="box zs">`
	if tile.powerUp != nil || tile.money != 0 || tile.boosts != 0 {
		svgtag += `<svg width="22" height="22">`
		if tile.powerUp != nil {
			svgtag += `<circle class="svgRed" cx="7" cy="7" r="7" />`
		}
		if tile.money != 0 {
			svgtag += `<circle class="svgGreen" cx="7" cy="14" r="7" />`
		}
		if tile.boosts != 0 {
			svgtag += `<circle class="svgBlue" cx="14" cy="14" r="7" />`
		}
		svgtag += `</svg>`
	}
	svgtag += "</div>"
	sId := fmt.Sprintf("s%d-%d", tile.y, tile.x)
	return fmt.Sprintf(svgtag, sId)
}
