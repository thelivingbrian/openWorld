package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type ChatRequest struct {
	Token       string `json:"token"`
	ChatMessage string `json:"chat_message"`
}

type ScreenRequest struct {
	Token string `json:"token"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type client struct {
	conn *websocket.Conn
}

type Update struct {
	player *Player
	update string
}

var clients = make(map[*client]bool)
var broadcast = make(chan string)
var updates = make(chan Update)

func ws_chat(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	newClient := &client{
		conn: conn,
	}
	clients[newClient] = true

	for {
		_, bytes, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			delete(clients, newClient)
			return
		}

		fmt.Printf("Received message: %s\n", bytes)
		jsonData := string(bytes)

		var msg ChatRequest
		err = json.Unmarshal([]byte(jsonData), &msg)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			return
		}

		fmt.Println("Token:", msg.Token)
		fmt.Println("Chat Message:", msg.ChatMessage)

		// This is a bug because it will wipe keyed messages if a new message comes in
		message := `
		<input id="msg" type="text" name="chat_message" value="">
		
		<div id="chat_room" hx-swap-oob="beforeend:#chat_room">
			<p>` + msg.Token + `: ` + msg.ChatMessage + `</p>
		</div>`

		broadcast <- message
	}
}

func sendMessageToAll(messageType int, message []byte) {
	for client := range clients {
		if err := client.conn.WriteMessage(messageType, message); err != nil {
			fmt.Println(err)
			return
		}
	}
}

func sendUpdate(messageType int, update Update) {
	update.player.conn.WriteMessage(messageType, []byte(update.update))
}

func ws_screen(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	_, bytes, err := conn.ReadMessage()
	if err != nil {
		fmt.Println(err)
		return
	}
	jsonData := string(bytes)

	var msg ScreenRequest
	err = json.Unmarshal([]byte(jsonData), &msg)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}
	token := msg.Token

	existingPlayer, playerExists := playerMap[token]
	if playerExists {
		existingPlayer.conn = conn
		placeOnStage(existingPlayer)

		fmt.Println("Placed on stage")
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println("new screen" + existingPlayer.id)
		}
	} else {
		fmt.Println("player not found with token: " + token)
	}
}
