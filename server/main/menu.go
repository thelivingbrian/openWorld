package main

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"

	"github.com/gorilla/websocket"
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

//menu left goes back.
//color select as a menu
// [] red
// [] pink
// etc

var menuTemplate = `
<div id="modal_background" class="modal_bg">
	<div id="modal_menu" class="modal_content {{.CssClass}}">

		<div id="modal_information">
			{{.InfoHtml}}
		</div>
		
		<div id="modal_options">
			{{$name := .Name}}
			<input id="menu_name" type="hidden" name="menuName" value="{{$name}}" />
			<input id="menu_click_indicator" type="hidden" name="eventname" value="menuClick" />
			<input id="menu_selected_index" type="hidden" name="arg0" value="0" />
			
			<input id="menuOff" type="hidden" ws-send hx-trigger="keydown[key=='m'||key=='M'||key=='Escape'] from:body" hx-include="#token" name="eventname" value="menuOff" />
			<input id="menuUp" type="hidden" ws-send hx-trigger="keydown[key=='w'||key=='W'||key=='ArrowUp'] from:body" hx-include="#token, #menu_selected_index, #menu_name" name="eventname" value="menuUp" />
			<input id="menuDown" type="hidden" ws-send hx-trigger="keydown[key=='s'||key=='S'||key=='ArrowDown'] from:body" hx-include="#token, #menu_selected_index, #menu_name" name="eventname" value="menuDown" />
			<input id="menuKey" type="hidden" ws-send hx-trigger="keydown[key=='Enter'] from:body" hx-include="#token, #menu_selected_index, #menu_name" name="eventname" value="menuClick" />
			
			{{range  $i, $link := .Links}}
				<input id="menuClick_{{$name}}_{{$i}}" type="hidden" ws-send hx-trigger="click from:#menu_{{$name}}_{{$i}}" hx-include="#token, #menu_click_indicator, #menu_name" name="arg0" value="{{$i}}" />
				<span id="menu_{{$name}}_{{$i}}">
					<a id="menulink_{{$name}}_{{$i}}" {{if eq $i 0}} class="selected"{{end}} href="#"> {{$link.Text}} </a>
				</span><br />
			{{end}}
		</div>

		<div id="modal_script">
			<script>
				// Script for horizantal and vertical scrolling?
				// Future state? 
					// This is inconvinient because a keydown listener may need to be on body and 
					// then removing the event listener requires some extra sauce when the menu closes
			</script>
		</div>
	</div>
</div>
`

var menuTmpl = template.Must(template.New("menu").Parse(menuTemplate))

var pauseMenu = Menu{
	Name:     "pause",
	CssClass: "",
	InfoHtml: "",
	Links: []MenuLink{
		{Text: "Resume", eventHandler: turnMenuOff, auth: nil},
		{Text: "You", eventHandler: openStatsMenu, auth: nil},
		{Text: "Map", eventHandler: openMapMenu, auth: nil},
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
	InfoHtml: "<h2>Coming Soon</h2>",
	Links: []MenuLink{
		{Text: "Back", eventHandler: openPauseMenu, auth: nil},
	},
}

func turnMenuOn(p *Player, menuName string) {
	menu, ok := p.menues[menuName]
	if ok {
		sendMenu(p, menu)
	}
}

func sendMenu(p *Player, menu Menu) {
	var buf bytes.Buffer
	err := menuTmpl.Execute(&buf, menu)
	if err != nil {
		fmt.Println(err)
	}
	buf.WriteString(divInputDisabled())
	p.trySend(buf.Bytes())

}

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
		p.trySend([]byte(menu.menuSelectUp(event.Arg0)))
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
		p.trySend([]byte(menu.menuSelectDown(event.Arg0)))
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
	p.trySend([]byte(divModalDisabled() + divInput()))
}

func Quit(p *Player) {
	defer logOut(p)
	logOutSuccess := `
	  <div id="page">
	      <div id="logo">
	          <img src="/assets/blooplogo2.webp" width="400" height="400" alt="Welcome to bloopworld"><br />
	      </div>
	      <div id="landing">   
		  	  <span>Log out success!</span><br />
	          <a class="large-font" href="#" hx-post="/resume" hx-target="#landing">Resume</a><br />
	      </div>
	  </div>`

	//p.trySend([]byte(logOutSuccess)) // This races with logout
	p.connLock.Lock()
	if p.conn != nil {
		p.conn.WriteMessage(websocket.TextMessage, []byte(logOutSuccess))
	}
	p.connLock.Unlock()
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
	err := menuTmpl.Execute(&buf, copy)
	if err != nil {
		fmt.Println(err)
	}
	p.trySend(buf.Bytes())
}

func openPauseMenu(p *Player) {
	turnMenuOn(p, "pause")
}

func openStatsMenu(p *Player) {
	turnMenuOn(p, "stats")
}

// Player specific menues

func continueTeleporting(teleport *Teleport) Menu {
	return Menu{
		Name:     "teleport",
		CssClass: "",
		InfoHtml: "<h2>Continue?</h2>",
		Links: []MenuLink{
			{Text: "Yes", eventHandler: teleportEventHandler(teleport), auth: sourceStageAuthorizer(teleport.sourceStage)},
			{Text: "No", eventHandler: turnMenuOff, auth: nil},
		},
	}
}

func teleportEventHandler(teleport *Teleport) func(*Player) {
	return func(player *Player) {
		previousTile := player.tile
		player.applyTeleport(teleport)

		impactedTiles := player.updateSpaceHighlights()
		updateOneAfterMovement(player, impactedTiles, previousTile)
		turnMenuOff(player) // try other order
	}
}

func sourceStageAuthorizer(source string) func(*Player) bool {
	return func(p *Player) bool {
		return p.stageName == source
	}
}
