package main

type Player struct {
	id          string
	stage       *Stage
	stageName   string
	viewIsDirty bool
	x           int
	y           int
}

func (player *Player) printDirty() string {
	if player.viewIsDirty {
		return ` 
		<form id="dirtyForm" hx-post="/dirty" hx-trigger="load" hx-target="#dirtyForm" hx-swap="outerHTML">
			<input type="hidden" name="token" value="` + player.id + `" />
			<input id="dirty" type="hidden" name="dirty" value="true" />
		</form>`
	} else {
		return `
		<form id="dirtyForm" hx-post="/dirty" hx-trigger="load" hx-target="#dirtyForm" hx-swap="outerHTML">
			<input type="hidden" name="token" value="` + player.id + `" />
			<input id="dirty" type="hidden" name="dirty" value="false" />
		</form>`
	}
}
