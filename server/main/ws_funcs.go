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
			fmt.Println(err)
			fmt.Println("CONN ERROR HERE")
			// Close connection remove player etc
			return
		}

		key, token, success := getKeyPress(msg)
		if !success {
			fmt.Println("Invalid input")
			continue
		}
		if token != existingPlayer.id {
			fmt.Println("Cheating")
			break
		}

		existingPlayer.handlePress(key)

		//fmt.Println("new message: " + string(msg))
	}
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

func getKeyPress(input []byte) (key string, token string, success bool) {
	var msg struct {
		Token    string `json:"token"`
		KeyPress string `json:"keypress"`
	}
	err := json.Unmarshal(input, &msg)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return "", "", false
	}
	return msg.KeyPress, msg.Token, true
}

func (player *Player) handlePress(key string) {
	if key == "W" {
		moveNorth(player)
	}
	if key == "A" {
		moveWest(player)
	}
	if key == "S" {
		moveSouth(player)
	}
	if key == "D" {
		moveEast(player)
	}
	if key == "Space-On" {
		player.turnSpaceOn()
	}
	if key == "Space-Off" {
		reactivate := `<input id="space-on" type="hidden" ws-send hx-trigger="keydown[key==' '] from:body once" hx-include="#token" name="keypress" value="Space-On" />`
		updateOne(reactivate, player)
		player.turnSpaceOff()
	}

}
