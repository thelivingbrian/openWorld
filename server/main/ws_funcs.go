package main

import (
	"encoding/json"
	"fmt"
	"net/http"

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
		player.moveWest()
	}
	if event.Name == "s" {
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
		if player.actions.spaceStack.hasPower() {
			player.activatePower()
		}
	}
	if event.Name == "Shift-On" {
		updateOne(divInputMobileShift(), player)
	}
	if event.Name == "Shift-Off" {
		updateOne(divInputMobile(), player)
	}
	if event.Name == "menuOn" {
		event.MenuName = "pause" // Gross but need to think about
		turnMenuOn(player, *event)
	}
	if event.Name == "menuOff" {
		turnMenuOff(player, *event)
	}
	if event.Name == "menuDown" {
		menuDown(player, *event)
	}
	if event.Name == "menuUp" {
		menuUp(player, *event)
	}
	if event.Name == "menuClick" {
		menu, ok := menues[event.MenuName]
		if ok {
			menu.attemptClick(player, *event)
		}
	}
}
