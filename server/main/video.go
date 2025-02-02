package main

import "fmt"

func flashBackgroundColor(player *Player, color string) {
	script := fmt.Sprintf(`<div id="script"><script>flashBg("%s")</script></div>`, color)
	updateOne(script, player)
}

func changePageBackgroundColor(player *Player, bgColor string) {
	script := fmt.Sprintf(`<div id="script"><script>document.body.className="%s"</script></div>`, bgColor)
	updateOne(script, player)
}
