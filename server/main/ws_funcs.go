package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type Update struct {
	player *Player
	update []byte // should be []byte not string, conversion cost twice
}

var (
	clients  = make(map[*websocket.Conn]bool)
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func ws_chat(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	clients[conn] = true // Should this be seperate from players?

	broadcastMessages(conn, broadcast)
}

func broadcastMessages(conn *websocket.Conn, chats chan string) {
	for {
		_, bytes, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			delete(clients, conn)
			return
		}
		fmt.Printf("Received message: %s\n", bytes)

		var msg struct {
			Token       string `json:"token"`
			ChatMessage string `json:"chat_message"`
		}
		err = json.Unmarshal(bytes, &msg)
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

		chats <- message
	}
}

func sendMessageToAll(messageType int, message []byte) {
	for client := range clients {
		if err := client.WriteMessage(messageType, message); err != nil {
			fmt.Println(err)
			return
		}
	}
}

func ws_screen(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	token, success := getTokenFromFirstMessage(conn)
	if !success {
		fmt.Println("Invalid Connection")
		return
		//panic("oops")
	}

	existingPlayer, playerExists := playerMap[token]
	if playerExists {
		handleNewPlayer(existingPlayer, conn)
	} else {
		fmt.Println("player not found with token: " + token)
	}
}

func handleNewPlayer(existingPlayer *Player, conn *websocket.Conn) {
	existingPlayer.conn = conn
	placeOnStage(existingPlayer)
	fmt.Println("New Connection")
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("new message: " + existingPlayer.id)
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
