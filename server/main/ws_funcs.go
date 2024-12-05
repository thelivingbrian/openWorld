package main

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// ...move to different file?
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
		(*existingPlayer).conn = conn
		//existingPlayer.world = world

		fmt.Println(fmt.Sprintf("world locator - %p : %p - %s - %s", existingPlayer, (*existingPlayer), (*existingPlayer).id, (*existingPlayer).username))
		handleNewPlayer(existingPlayer)
	} else {
		fmt.Println("player not found with token: " + token)
	}
}

func handleNewPlayer(existingPlayer **Player) {
	go (*existingPlayer).sendUpdates()
	assignStageAndListen(existingPlayer)
	fmt.Println(fmt.Sprintf("placing - %p : %p - %s - %s", existingPlayer, (*existingPlayer), (*existingPlayer).id, (*existingPlayer).username))
	placeOnStage(existingPlayer)
	fmt.Println("New Connection")
	player := (*existingPlayer)
	for {
		_, msg, err := player.conn.ReadMessage()
		if err != nil {
			// This allows for rage quit by pressing X, should add timeout to encourage finding safety
			logOut(player)
			return
		}

		event, success := getKeyPress(msg)
		if !success {
			fmt.Println("Invalid input")
			continue
		}
		if event.Token != player.id {
			fmt.Println("Cheating")
			break
		}
		// Throttle input here?

		player.handlePress(event)
		if player.conn == nil {
			return
		}
	}
}

func logOut(player *Player) {
	player.updateRecord() // Should return error
	player.removeFromStage()
	player.world.wPlayerMutex.Lock()
	delete(player.world.worldPlayers, player.id)
	index, exists := player.world.leaderBoard.mostDangerous.index[player]
	if exists {
		heap.Remove(&player.world.leaderBoard.mostDangerous, index)
		//  If index was 0 before, need to update new most dangerous
		if index == 0 {
			fmt.Println("New Most Dangerous!")
			mostDangerous := player.world.leaderBoard.mostDangerous.Peek()
			if mostDangerous != nil {
				notifyChangeInMostDangerous(mostDangerous)
			}
		}
	}
	player.world.wPlayerMutex.Unlock()

	player.connLock.Lock()
	defer player.connLock.Unlock()
	player.conn = nil

	close(player.updates)

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

var npcs = 0

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
		player.tile.playerMutex.Lock()
		defer player.tile.playerMutex.Unlock()
		for _, player := range player.tile.playerMap {
			fmt.Println(fmt.Sprintf("%p : %p - %s - %s", player, (*player), (*player).id, (*player).username))
			//return *player
		}
	}
	if event.Name == "g" {
		/*
			Full swap takes priority in either order, otherwise both may apply
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
		*/
		//player.updateBottomText("Heyo ;) ")
		go func() {
			randstr := fmt.Sprint(rand.Intn(50000000))
			p1 := player.world.join(&PlayerRecord{Username: "hello" + randstr, Health: 50, Y: player.y, X: player.x, StageName: player.stage.name, Team: "fuchsia", Trim: "white-b thick"})
			go func() {
				for {
					<-p1.updates
				}
			}()
			npcs++
			fmt.Println(npcs)

			player.world.wPlayerMutex.Lock()
			loc, ok := player.world.worldPlayers[p1.id]
			player.world.wPlayerMutex.Unlock()
			if !ok {
				return
			}
			assignStageAndListen(loc)
			placeOnStage(loc)
			fmt.Println(p1.stage.name + "Has spawned new npc")
			for {
				time.Sleep(250 * time.Millisecond)
				randn := rand.Intn(5000)

				if randn%4 == 0 {
					//fmt.Println(randn)
					p1.moveNorth()
				}
				if randn%4 == 1 {
					p1.moveSouth()
				}
				if randn%4 == 2 {
					p1.moveEast()
				}
				if randn%4 == 3 {
					p1.moveWest()
				}
			}
		}()
	}
	if event.Name == "Space-On" {
		if player.actions.spaceStack.hasPower() {
			player.activatePower()
		}
	}
	if event.Name == "Shift-On" {
		updateOne(divInputShift(), player)
	}
	if event.Name == "Shift-Off" {
		updateOne(divInput(), player)
	}
	if event.Name == "menuOn" {
		openPauseMenu(player)
	}
	if event.Name == "menuOff" {
		turnMenuOff(player)
	}
	if event.Name == "menuDown" {
		menuDown(player, *event)
	}
	if event.Name == "menuUp" {
		menuUp(player, *event)
	}
	if event.Name == "menuClick" {
		menu, ok := player.menues[event.MenuName]
		if ok {
			menu.attemptClick(player, *event)
		}
	}
}
