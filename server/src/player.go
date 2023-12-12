package main

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

func printStageFor(player *Player) string {
	var output string = `
	<div id="screen" class="grid" hx-swap-oob="true">	
	`
	for y, row := range player.stage.tiles {
		output += printRow(row, y)
	}
	output += `</div>`
	return output
}
