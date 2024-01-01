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

/*func parseTemplate() {
	tmpl, err := template.New("playerScreen").Parse(screenTemplate)
	if err != nil {
		panic(err)
	}
	parsedScreenTemplate = tmpl
}*/

func screenHtmlFromTemplate(player *Player) string {
	//parseTemplate()
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
