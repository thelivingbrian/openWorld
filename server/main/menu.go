package main

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
	"strings"
)

type Menu struct {
	Name       string
	CssClass   string
	InfoHtml   template.HTML
	Links      []MenuLink
	ScriptHtml string
}

type MenuLink struct {
	Text         string
	eventHandler func(*Player) // should handler and auth be consolidated?
	auth         func(*Player) bool
}

/////////////////////////////////////////////////////
// Default menus

var pauseMenu = Menu{
	Name:     "pause",
	CssClass: "",
	InfoHtml: "",
	Links: []MenuLink{
		{Text: "Resume", eventHandler: turnMenuOff, auth: nil},
		{Text: "You", eventHandler: openStatsMenu, auth: nil},
		{Text: "Map", eventHandler: openMapMenu, auth: nil},
		{Text: "Respawn", eventHandler: openRespawnMenu, auth: excludeSpecialStages},
		{Text: "Quit", eventHandler: Quit, auth: nil},
	},
}
var mapMenu = Menu{
	Name:     "map",
	CssClass: "",
	InfoHtml: "",
	Links: []MenuLink{
		{Text: "Back", eventHandler: openPauseMenu, auth: nil},
	},
}
var statsMenu = Menu{
	Name:     "stats",
	CssClass: "",
	InfoHtml: `<h2>Stat population error.</h2>`,
	Links: []MenuLink{
		{Text: "Back", eventHandler: openPauseMenu, auth: nil},
	},
}

var respawnMenu = Menu{
	Name:     "respawn",
	CssClass: "",
	InfoHtml: "<h3>Are you sure? (you will die)</h3>",
	Links: []MenuLink{
		{Text: "Yes", eventHandler: turnMenuOffAnd(handleDeath), auth: excludeSpecialStages},
		{Text: "No", eventHandler: openPauseMenu, auth: nil},
	},
}

//////////////////////////////////////////////////////
// Menu send

func turnMenuOnByName(p *Player, menuName string) {
	menu, ok := p.menues[menuName]
	if ok {
		sendMenu(p, menu)
	}
}

func sendMenu(p *Player, menu Menu) {
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, "menu", menu)
	if err != nil {
		fmt.Println(err)
	}
	buf.WriteString(divInputDisabled())
	p.updates <- buf.Bytes()
}

/////////////////////////////////////////////////////
// Menu events

func (m *Menu) attemptClick(p *Player, e PlayerSocketEvent) {
	i, err := strconv.Atoi(e.Arg0)
	if err != nil {
		fmt.Println(err)
	}
	if i < 0 || i > len(m.Links) {
		fmt.Println("Invalid index")
	}
	auth := m.Links[i].auth
	handler := m.Links[i].eventHandler
	if handler != nil && (auth == nil || auth(p)) {
		handler(p)
	}
}

func menuUp(p *Player, event PlayerSocketEvent) {
	menu, ok := p.menues[event.MenuName]
	if ok {
		updateOne(menu.menuSelectUp(event.Arg0), p)
	}
}

func (menu *Menu) menuSelectUp(index string) string {
	i, err := strconv.Atoi(index)
	if err != nil {
		return ""
	}
	return menu.selectedLinkAt(i-1) + menu.unselectedLinkAt(i)
}

func menuDown(p *Player, event PlayerSocketEvent) {
	menu, ok := p.menues[event.MenuName]
	if ok {
		updateOne(menu.menuSelectDown(event.Arg0), p)
	}
}

func (menu *Menu) menuSelectDown(index string) string {
	i, err := strconv.Atoi(index)
	if err != nil {
		return ""
	}
	return menu.selectedLinkAt(i+1) + menu.unselectedLinkAt(i)
}

// view updates
func (m *Menu) selectedLinkAt(i int) string {
	index := mod(i, len(m.Links)) // divide by 0
	out := `
	<input id="menu_selected_index" type="hidden" name="arg0" value="%d" />
	<a id="menulink_%s_%d" class="selected" href="#"> %s </a><br />`
	return fmt.Sprintf(out, index, m.Name, index, m.Links[index].Text)
}

func (m *Menu) unselectedLinkAt(i int) string {
	index := mod(i, len(m.Links)) // divide by 0
	out := `<a id="menulink_%s_%d" href="#"> %s </a><br />`
	return fmt.Sprintf(out, m.Name, index, m.Links[index].Text)
}

func mod(i, n int) int {
	return ((i % n) + n) % n
}

// Menu event handlers

func turnMenuOff(p *Player) {
	var buffer bytes.Buffer
	buffer.Write([]byte(divModalDisabled()))
	tmpl.ExecuteTemplate(&buffer, "input", nil)
	sendUpdate(p, buffer.Bytes())
}
func turnMenuOffAnd(f func(*Player)) func(*Player) {
	return func(p *Player) {
		turnMenuOff(p)
		f(p)
	}
}

func Quit(p *Player) {
	logOutSuccess := `
	  <div id="page">
	      <div id="logo">
	          <img src="/assets/blooplogo2.webp" width="400" height="400" alt="Welcome to bloopworld"><br />
	      </div>
	      <div id="landing">   
		  	  <span>Log out success!</span><br />
	          <a class="large-font" href="#" hx-post="/play" hx-target="#page">Resume</a><br />
	      </div>
	  </div>`

	sendUpdate(p, []byte(logOutSuccess))
	p.closeConnectionSync() // This will initiate log out
}

func openMapMenu(p *Player) {
	var buf bytes.Buffer
	copy := mapMenu
	if p.stage.mapId != "" {
		mapPath := "/images/" + p.stage.mapId
		copy.InfoHtml = template.HTML(`<img src="` + mapPath + `" width="100%" alt="map of space" />`)
	} else {
		copy.InfoHtml = `<h2>unavailable</h2>`

	}
	err := tmpl.ExecuteTemplate(&buf, "menu", copy)
	if err != nil {
		fmt.Println(err)
	}
	//p.trySend(buf.Bytes())
	p.updates <- buf.Bytes()
}

func openPauseMenu(p *Player) {
	turnMenuOnByName(p, "pause")
}

func openStatsMenu(p *Player) {
	menu := statsMenu
	menu.InfoHtml = createInfoHtmlForPlayer(p)
	sendMenu(p, menu)
}

func createInfoHtmlForPlayer(p *Player) template.HTML {
	htmlContent := fmt.Sprintf(
		`<div class="player-stats">
			<p><strong>  Total  </strong></p>
			<p>Goals: %d</p>
			<p>Kills: %d</p>
			<p>Deaths: %d</p>
		</div>`,
		p.getGoalsScored(), p.getKillCountSync(), p.getDeathCountSync(),
	)

	return template.HTML(htmlContent)
}

func openRespawnMenu(p *Player) {
	turnMenuOnByName(p, "respawn")
}

// Player specific menues

func continueTeleporting(teleport *Teleport) Menu {
	return Menu{
		Name:     "teleport",
		CssClass: "",
		InfoHtml: "<h2>Continue?</h2>",
		Links: []MenuLink{
			{Text: "Yes", eventHandler: teleportEventHandler(teleport), auth: sourceStageAuthorizerAffirmative(teleport.sourceStage)},
			{Text: "No", eventHandler: turnMenuOff, auth: nil},
		},
	}
}

func teleportEventHandler(teleport *Teleport) func(*Player) {
	return func(player *Player) {
		// No need for new routine?
		go func() {
			player.applyTeleport(teleport)
		}()
		turnMenuOff(player) // try other order
	}
}

func sourceStageAuthorizerAffirmative(source string) func(*Player) bool {
	return func(p *Player) bool {
		return p.getStageNameSync() == source
	}
}

func sourceStageAuthorizerExclude(source string) func(*Player) bool {
	return func(p *Player) bool {
		return p.getStageNameSync() != source
	}
}

func excludeSpecialStages(p *Player) bool {
	stagename := p.getStageNameSync()
	if strings.HasPrefix(stagename, "infirmary") {
		return false
	}
	if strings.HasPrefix(stagename, "tutorial") {
		return false
	}
	return true
}
