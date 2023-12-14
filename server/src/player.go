package main

import (
	"fmt"
	"strconv"
)

type Player struct {
	id          string
	stage       *Stage
	stageName   string
	viewIsDirty bool
	x           int
	y           int
	actions     *Actions
	health      int
}

type Actions struct {
	space bool
}

func printPageHeaderFor(player *Player) string {
	return `
	<div id="page">
    <div id="controls">      
		<input hx-post="/w" hx-trigger="keyup[key=='w'] from:body" type="hidden" name="token" value="` + player.id + `" />
		<input hx-post="/s" hx-trigger="keyup[key=='s'] from:body" type="hidden" name="token" value="` + player.id + `" />
		<input hx-post="/a" hx-trigger="keyup[key=='a'] from:body" type="hidden" name="token" value="` + player.id + `" />
		<input hx-post="/d" hx-trigger="keyup[key=='d'] from:body" type="hidden" name="token" value="` + player.id + `" />
		<input id="spaceOn" hx-post="/spaceOn" hx-trigger="keydown[key==' '] from:body once" type="hidden" name="token" value="` + player.id + `" />
		<input hx-post="/spaceOff" hx-trigger="keyup[key==' '] from:body" hx-target="#spaceOn" hx-swap="outerHTML" type="hidden" name="token" value="` + player.id + `" />	
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

func spaceHighlight(tile *Tile) string {
	if walkable(tile) {
		return "green"
	} else {
		return "red"
	}
}

func livingView(player *Player) string {
	output := ""

	// Get defaul colors
	var tileColors [][]string = make([][]string, len(player.stage.tiles))
	for i, row := range player.stage.tiles {
		tileColors[i] = colorArray(row)
	}

	// Add player
	tileColors[player.y][player.x] = "fusia"

	// Add Space
	if player.actions.space {
		hiY := player.y + 2
		loY := player.y - 2
		hiX := player.x + 2
		loX := player.x - 2
		validHighY := (len(player.stage.tiles)-hiY-1 >= 0)
		validHighX := (len(player.stage.tiles[0])-hiX-1 >= 0)
		if validHighY {
			tileColors[player.y+2][player.x] = spaceHighlight(&player.stage.tiles[player.y+2][player.x])
		}
		if loY >= 0 {
			tileColors[player.y-2][player.x] = spaceHighlight(&player.stage.tiles[player.y-2][player.x])
		}
		if validHighX {
			tileColors[player.y][player.x+2] = spaceHighlight(&player.stage.tiles[player.y][player.x+2])
		}
		if loX >= 0 {
			tileColors[player.y][player.x-2] = spaceHighlight(&player.stage.tiles[player.y][player.x-2])
		}
	}

	output += htmlFromColorMatrix(tileColors)

	return output
}

func printStageFor(player *Player) string {
	var output string = `
	<div id="screen" class="grid" hx-swap-oob="true">	
	`

	if player.health > 0 {
		output += livingView(player)
	} else {
		output += `<h2>You Died.</h2>`
		stageMutex.Lock()
		existingStage, stageExists := stageMap["clinic"]
		if !stageExists {
			fmt.Println("New Stage")
			newStage := getStageByName("clinic")
			stagePtr := &newStage
			stageMap["clinic"] = stagePtr
			existingStage = stagePtr
		}
		stageMutex.Unlock()

		player.health = 100
		player.stage = existingStage
		existingStage.placeOnStage(player)
		output += livingView(player)
	}

	output += `</div>`
	player.viewIsDirty = false
	return output

}
