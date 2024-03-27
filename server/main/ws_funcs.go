package main

import (
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

		key, token, arg0, success := getKeyPress(msg)
		if !success {
			fmt.Println("Invalid input")
			continue
		}
		if token != existingPlayer.id {
			fmt.Println("Cheating")
			break
		}

		existingPlayer.handlePress(key, arg0)
	}
}

func logOut(player *Player) {
	player.updateRecord() // Should return error
	player.removeFromStage()
	player.world.wPlayerMutex.Lock()
	delete(player.world.worldPlayers, player.id)
	player.world.wPlayerMutex.Unlock()
	player.conn = nil
	fmt.Println("Logging Out")
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

func getKeyPress(input []byte) (key string, token string, arg0 string, success bool) {
	// rename Keypress to event?
	// Reuse this struct somehow? Player.LatestMessage *msg
	var msg struct {
		Token    string `json:"token"`
		KeyPress string `json:"keypress"`
		Arg0     string `json:"arg0"`
	}
	err := json.Unmarshal(input, &msg)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return "", "", "", false
	}
	return msg.KeyPress, msg.Token, msg.Arg0, true
}

func (player *Player) handlePress(key string, arg0 string) {
	if key == "w" {
		/*class := `<div id="script" hx-swap-oob="true"> <script>document.body.className = "twilight"</script> </div>`
		updateOne(class, player)*/
		player.moveNorth()
	}
	if key == "a" {
		/*class := `<div id="script" hx-swap-oob="true"> <script>document.body.className = "day"</script> </div>`
		updateOne(class, player)*/
		player.moveWest()
	}
	if key == "s" {
		/*class := `<div id="script" hx-swap-oob="true"> <script>document.body.className = "night"</script> </div>`
		updateOne(class, player)*/
		player.moveSouth()
	}
	if key == "d" {
		player.moveEast()
	}
	if key == "W" {
		player.moveNorthBoost()
	}
	if key == "A" {
		player.moveWestBoost()
	}
	if key == "S" {
		player.moveSouthBoost()
	}
	if key == "D" {
		player.moveEastBoost()
	}
	if key == "f" {
		updateScreenFromScratch(player)
	}
	if key == "g" {
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
		exTile += `<div id="t1-0" class="box top green"></div>`
		updateOne(exTile, player)
	}
	if key == "menuOn" {
		updateOne(divModalMenu()+divInputDisabled(), player)
	}
	if key == "menuOff" {
		updateOne(divModalDisabled()+divInputDesktop(), player)
	}
	if key == "menuDown" {
		updateOne(menuSelectDown(arg0), player)
	}
	if key == "menuUp" {
		updateOne(menuSelectUp(arg0), player)
	}
	if key == "Space-On" {
		reactivate := `<input id="space-on" type="hidden" ws-send hx-trigger="keydown[key==' '] from:body once" hx-include="#token" name="keypress" value="Space-On" />`
		updateOne(reactivate, player)
		if player.actions.spaceStack.hasPower() {
			player.activatePower()
		}
	}
}

func menuSelectDown(index string) string {
	i, err := strconv.Atoi(index)
	if err != nil {
		return ""
	}
	return divPauseMenu(i + 1)
}

func menuSelectUp(index string) string {
	i, err := strconv.Atoi(index)
	if err != nil {
		return ""
	}
	return divPauseMenu(i - 1)
}
