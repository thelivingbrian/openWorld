package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

type Update struct {
	player *Player
	update []byte
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func (world *World) NewSocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	token, success := getTokenFromFirstMessage(conn) // Pattern for input?
	if !success {
		fmt.Println("Invalid Connection")
		return
	}

	// New method on world
	world.wPlayerMutex.Lock()
	existingPlayer, playerExists := world.worldPlayers[token]
	world.wPlayerMutex.Unlock()

	if playerExists {
		existingPlayer.conn = conn
		existingPlayer.world = world
		handleNewPlayer(existingPlayer)
	} else {
		fmt.Println("player not found with token: " + token)
	}
}

func handleNewPlayer(existingPlayer *Player) {
	existingPlayer.assignStageAndListen()
	existingPlayer.placeOnStage()
	fmt.Println("New Connection")
	for {
		_, msg, err := existingPlayer.conn.ReadMessage()
		if err != nil {
			// This allows for rage quit by pressing X, should add timeout to encourage finding safety
			logOut(existingPlayer)
			return
		}

		event, success := getKeyPress(msg)
		if !success {
			fmt.Println("Invalid input")
			continue
		}
		if event.Token != existingPlayer.id {
			fmt.Println("Cheating")
			break
		}
		// Throttle input here?

		existingPlayer.handlePress(event)
		if existingPlayer.conn == nil {
			return
		}
	}
}

func logOut(player *Player) {
	player.updateRecord() // Should return error
	player.removeFromStage()
	player.world.wPlayerMutex.Lock()
	delete(player.world.worldPlayers, player.id)
	player.world.wPlayerMutex.Unlock()
	player.conn = nil
	fmt.Println("Logging Out " + player.username)
}

func getTokenFromFirstMessage(conn *websocket.Conn) (token string, success bool) {
	_, bytes, err := conn.ReadMessage()
	if err != nil {
		fmt.Println(err)
		return "", false
	}

	var msg struct {
		Token string `json:"token"`
	}
	err = json.Unmarshal(bytes, &msg)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return "", false
	}

	return msg.Token, true
}

type PlayerSocketEvent struct {
	Token    string `json:"token"`
	Name     string `json:"eventname"`
	MenuName string `json:"menuName"`
	Arg0     string `json:"arg0"`
}

func getKeyPress(input []byte) (event *PlayerSocketEvent, success bool) {
	// rename Keypress to event?
	// Reuse this struct somehow? Player.LatestMessage *event
	/*var event struct {
		Token    string `json:"token"`
		KeyPress string `json:"keypress"`
		MenuName string `json:"menuName"`
		Arg0     string `json:"arg0"`
	}*/
	event = &PlayerSocketEvent{}
	err := json.Unmarshal(input, event)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, false
	}
	return event, true
}

func (player *Player) handlePress(event *PlayerSocketEvent) {
	if event.Name == "w" {
		/*class := `<div id="script" hx-swap-oob="true"> <script>document.body.className = "twilight"</script> </div>`
		updateOne(class, player)*/
		player.moveNorth()
	}
	if event.Name == "a" {
		/*class := `<div id="script" hx-swap-oob="true"> <script>document.body.className = "day"</script> </div>`
		updateOne(class, player)*/
		player.moveWest()
	}
	if event.Name == "s" {
		/*class := `<div id="script" hx-swap-oob="true"> <script>document.body.className = "night"</script> </div>`
		updateOne(class, player)*/
		player.moveSouth()
	}
	if event.Name == "d" {
		player.moveEast()
	}
	if event.Name == "W" {
		player.moveNorthBoost()
	}
	if event.Name == "A" {
		player.moveWestBoost()
	}
	if event.Name == "S" {
		player.moveSouthBoost()
	}
	if event.Name == "D" {
		player.moveEastBoost()
	}
	if event.Name == "f" {
		updateScreenFromScratch(player)
	}
	if event.Name == "g" {
		exTile := `<div class="grid-square blue" id="c0-0">				
						<div id="p0-0" class="box zp "></div>
						<div id="s0-0" class="box zS"></div>
						<div id="t0-0" class="box top"></div>
					</div>
					<div class="grid-square blue" id="c0-1">				
						<div id="p0-1" class="box zp "></div>
						<div id="s0-1" class="box zS"></div>
						<div id="t0-1" class="box top"></div>
					</div>
					`
		exTile += `<div id="t1-0" class="box top green"></div>
					<div id="t0-0" class="box top green"></div>`
		updateOne(exTile, player)
	}
	if event.Name == "Space-On" {
		// I don't think there is any advantage in reactivating in this way?
		//reactivate := `<input id="space-on" type="hidden" ws-send hx-trigger="keydown[key==' '] from:body once" hx-include="#token" name="eventname" value="Space-On" />`
		//updateOne(reactivate, player)
		if player.actions.spaceStack.hasPower() {
			player.activatePower()
		}
	}
	if event.Name == "menuOn" {
		var buf bytes.Buffer
		err := menuTmpl.Execute(&buf, pauseMenu)
		if err != nil {
			fmt.Println(err)
		}

		buf.WriteString(divInputDisabled())
		player.conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}
	if event.Name == "menuOff" {
		turnMenuOff(player, *event)
	}
	if event.Name == "menuDown" {
		updateOne(menuSelectDown(event.Arg0), player)
	}
	if event.Name == "menuUp" {
		updateOne(menuSelectUp(event.Arg0), player)
	}
	if event.Name == "menuClick" {
		//fmt.Println(event.Arg0)
		//fmt.Println(event.MenuName)
		menu, ok := menues[event.MenuName]
		if ok {
			menu.attemptClick(player, *event)
		}
	}
}

// add recv of Menu
func menuSelectDown(index string) string {
	i, err := strconv.Atoi(index)
	if err != nil {
		return ""
	}
	return pauseMenu.selectedLinkAt(i+1) + pauseMenu.unselectedLinkAt(i)
}

func menuSelectUp(index string) string {
	i, err := strconv.Atoi(index)
	if err != nil {
		return ""
	}
	return pauseMenu.selectedLinkAt(i-1) + pauseMenu.unselectedLinkAt(i)
}
