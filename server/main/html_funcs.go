package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"regexp"
)

////////////////////////////////////////////////////////////
// Quickswaps / screen

func emptyScreenForStage(stage *Stage) []byte {
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, "player-screen", stage.tiles)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func entireScreenAsSwaps(player *Player) []byte {
	currentTile := player.getTileSync()
	return swapsForTilesWithHighlights(currentTile.stage.tiles, duplicateMapOfHighlights(player))
}

func swapsForTilesWithHighlights(tiles [][]*Tile, highlights map[*Tile]bool) []byte {
	var buf bytes.Buffer
	for y := range tiles {
		for x := range tiles[y] {
			highlightColor := ""
			_, found := highlights[tiles[y][x]]
			if found {
				highlightColor = spaceHighlighter()
			}
			tileSwaps := swapsForTile(tiles[y][x], highlightColor)
			fmt.Fprintf(&buf, tileSwaps)
		}
	}
	return buf.Bytes()
}

func swapsForTile(tile *Tile, highlight string) string {
	svgtag := svgFromTile(tile)
	return fmt.Sprintf(tile.quickSwapTemplate, playerBox(tile), interactableBox(tile), svgtag, emptyWeatherBox(tile.y, tile.x), oobHighlightBox(tile, highlight))
}

////////////////////////////////////////////////////////////
// Player Information

func divPlayerInformation(player *Player) string {
	return `
	<div id="info" hx-swap-oob="true">
		<b>` + playerInformation(player) + `</b>
	</div>`
}

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

// use through socket can prevent swap bypass
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
	return fmt.Sprintf(`[~ id="Lp1-%d-%d" class="box zp %s"]`, y, x, icon)
}

func playerBox(tile *Tile) string {
	playerIndicator := ""
	if p := tile.getAPlayer(); p != nil {
		playerIndicator = p.getIconSync()
	}
	return playerBoxSpecifc(tile.y, tile.x, playerIndicator)
}

// Interactable box
func interactableBox(tile *Tile) string {
	tile.interactableMutex.Lock()
	defer tile.interactableMutex.Unlock()
	indicator := ""
	if tile.interactable != nil {
		indicator = tile.interactable.cssClass
	}
	return fmt.Sprintf(`[~ id="Li1-%d-%d" class="box zi %s"]`, tile.y, tile.x, indicator)
}

func lockedInteractableBox(tile *Tile) string {
	indicator := ""
	if tile.interactable != nil {
		indicator = tile.interactable.cssClass
	}
	return fmt.Sprintf(`[~ id="Li1-%d-%d" class="box zi %s"]`, tile.y, tile.x, indicator)
}

func emptyWeatherBox(y, x int) string {
	//  blue trsp20 for gloom
	return fmt.Sprintf(`[~ id="Lw1-%d-%d" class="box zw"]`, y, x)
}

func highlightBoxesForPlayer(player *Player, tiles []*Tile) string {
	highlights := ""

	playerHighlightCopy := duplicateMapOfHighlights(player)
	for _, tile := range tiles {
		if tile == nil {
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
	template := `[~ id="Lt1-%d-%d" class="box top %s"]`
	return fmt.Sprintf(template, tile.y, tile.x, cssClass)
}

func weatherBox(tile *Tile, cssClass string) string {
	template := `[~ id="Lw1-%d-%d" class="box zw %s"]`
	return fmt.Sprintf(template, tile.y, tile.x, cssClass)
}

func svgFromTile(tile *Tile) string {
	tile.powerMutex.Lock()
	defer tile.powerMutex.Unlock()
	tile.moneyMutex.Lock()
	defer tile.moneyMutex.Unlock()
	tile.boostsMutex.Lock()
	defer tile.boostsMutex.Unlock()

	template := `[~ id="%s" class="%s"]`

	svgId := fmt.Sprintf("Ls1-%d-%d", tile.y, tile.x)

	classes := "box zs "
	if tile.powerUp != nil {
		classes += "svgRed "
	}
	if tile.money != 0 {
		classes += "svgGreen "
	}
	if tile.boosts != 0 {
		classes += "svgBlue "
	}

	return fmt.Sprintf(template, svgId, classes)
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
	// uses ws bypass to function
	return `[~ id="Lx1-0-0" class="container"][~ id="Lx1-0-1" class="container hidden"]`
}

func divInputShift() string {
	// uses ws bypass to function
	return `[~ id="Lx1-0-0" class="container hidden"][~ id="Lx1-0-1" class="container"]`
}

func divInputDisabled() string {
	return `
	<div id="input">

	</div>
`
}

func divLogOutResume(text, domain string) []byte {
	var buf bytes.Buffer
	logOutSuccess := `
	  <div id="page">
	      <div id="logo">
	          <img src="/assets/blooplogo2.webp" width="400" height="400" alt="Welcome to bloopworld"><br />
	      </div>
	      <div id="landing">   
		  	  <span>%s</span><br />
	          <a class="large-font" href="#" hx-post="%s/play" hx-target="#page">Resume</a><br />
	      </div>
	  </div>`
	fmt.Fprintf(&buf, logOutSuccess, text, domain)
	return buf.Bytes()
}
