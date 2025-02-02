package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"regexp"
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

	currentTile := player.getTileSync()
	tileHtml := htmlFromTileGrid(player.getStageSync().tiles, currentTile.y, currentTile.x, duplicateMapOfHighlights(player))

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

func htmlForTile(tile *Tile, highlight string) string {
	svgtag := svgFromTile(tile)
	// grab tile y and x only once here or in parent method?
	// Lock interactable before getting box
	return fmt.Sprintf(tile.htmlTemplate, playerBox(tile), lockedInteractableBox(tile), svgtag, emptyWeatherBox(tile.y, tile.x), oobHighlightBox(tile, highlight))
}

////////////////////////////////////////////////////////////
// Player Information

func divPlayerInformation(player *Player) string {
	return `
	<div id="info" hx-swap-oob="true">
		<b>` + playerInformation(player) + `</b>
	</div>`
}

// needs improvement
func playerInformation(player *Player) string {
	hearts := getHeartsFromHealth(player.getHealthSync())
	return fmt.Sprintf(`%s %s<br />%s | %s | %s %s`, player.username, hearts, spanStreak(player.getKillStreakSync()), spanBoosts(player.getBoostCountSync()), spanMoney(player.getMoneySync()), spanPower(player.actions.spaceStack.count()))
}

func spanPower(quantity int) string {
	if quantity == 0 {
		return `<span id="power"></span>`
	} else {
		return fmt.Sprintf(`<span id="power"> 🗡️x%d</span>`, quantity)
	}
}

func spanStreak(quantity int) string {
	return fmt.Sprintf(`<span id="streak" class="red">Streak %d</span>`, quantity)
}
func spanBoosts(quantity int) string {
	return fmt.Sprintf(`<span id="boosts" class="blue">^ %d</span>`, quantity)
}
func spanMoney(quantity int) string {
	return fmt.Sprintf(`<span id="money" class="dark-green">$ %d</span>`, quantity)
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

/////////////////////////////////////////////
// Bottom Text

var (
	// Regular expression for *[color]
	wordRegex = regexp.MustCompile(`\*\[(.+?)\]`)

	// Regular expression for @[phrase|color]
	phraseColorRegex = regexp.MustCompile(`@\[(.+?)\|(.+?)\]`)

	// Regular expression for @[phrase|---]
	teamColorWildRegex = regexp.MustCompile(`@\[(.*?)\|---\]`)
)

func processStringForColors(input string) string {
	input = wordRegex.ReplaceAllString(input, `<strong class="$1-t">$1</strong>`)
	input = phraseColorRegex.ReplaceAllString(input, `<strong class="$2-t">$1</strong>`)
	return input
}

func divBottomInvalid(s string) string {
	return `
	<div id="bottom_text" hx-swap-oob="true">
		<p style='color:red'>	
			` + s + `  
		</p>
	</div>`
}

/////////////////////////////////////////////////////
// Boxes

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

func lockedInteractableBox(tile *Tile) string {
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

func highlightBoxesForPlayer(player *Player, tiles []*Tile) string {
	highlights := ""

	playerHighlightCopy := duplicateMapOfHighlights(player)
	for _, tile := range tiles {
		if tile == nil {
			fmt.Println(".") // seems to match number of actual highlights to add
			continue
		}
		if tile.stage != player.stage {
			continue
		}

		_, impactsHud := playerHighlightCopy[tile]
		if impactsHud {
			highlights += oobHighlightBox(tile, spaceHighlighter())
			continue
		}

		highlights += oobHighlightBox(tile, "")
	}

	return highlights
}

func duplicateMapOfHighlights(player *Player) map[*Tile]bool {
	player.actions.spaceHighlightMutex.Lock()
	defer player.actions.spaceHighlightMutex.Unlock()
	original := player.actions.spaceHighlights
	duplicate := make(map[*Tile]bool, len(original))
	for key, value := range original {
		duplicate[key] = value
	}
	return duplicate
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
	tile.powerMutex.Lock()
	defer tile.powerMutex.Unlock()
	tile.moneyMutex.Lock()
	defer tile.moneyMutex.Unlock()
	tile.boostsMutex.Lock()
	defer tile.boostsMutex.Unlock()
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

///////////////////////////////////////////
// Colors

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

///////////////////////////////////////////////////////////
// Divs

func divModalDisabled() string {
	return `
	<div id="modal_background">
		
	</div>
	`
}

func divInput() string {
	// uses htmx bypass to function
	return `
	<div id="x0-0" class="container">
	</div>
	<div id="x0-1" class="container hidden">
	</div>
`
}

func divInputShift() string {
	// uses htmx bypass to function
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
