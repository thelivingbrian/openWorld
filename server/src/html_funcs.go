package main

import (
	"bytes"
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

func screenHtmlFromTemplate(player *Player) string {
	var buf bytes.Buffer
	tileColors := make([][]string, len(player.stage.tiles))
	for y := range tileColors {
		tileColors[y] = make([]string, len(player.stage.tiles[y]))
		for x := range tileColors[y] {
			tileColors[y][x] = player.stage.tiles[y][x].CurrentCssClass
		}
	}
	tileColors[player.y][player.x] = "fusia"
	if player.actions.space {
		applyHighlights(player, tileColors, cross(), spaceHighlighter)
	}

	err := parsedScreenTemplate.Execute(&buf, tileColors)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

func printPageFor(player *Player) string {
	return `
	<div id="page">
		<div id="controls">      
			<input hx-post="/w" hx-trigger="keydown[key=='w'] from:body" type="hidden" name="token" value="` + player.id + `" />
			<input hx-post="/s" hx-trigger="keydown[key=='s'] from:body" type="hidden" name="token" value="` + player.id + `" />
			<input hx-post="/a" hx-trigger="keydown[key=='a'] from:body" type="hidden" name="token" value="` + player.id + `" />
			<input hx-post="/d" hx-trigger="keydown[key=='d'] from:body" type="hidden" name="token" value="` + player.id + `" />
			<input hx-post="/clear" hx-target="#screen" hx-swap="outerHTML" hx-trigger="keydown[key=='c'] from:body" type="hidden" name="token" value="` + player.id + `" />
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
