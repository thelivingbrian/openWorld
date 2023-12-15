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
	send chan string
}

var clients = make(map[*client]bool)
var broadcast = make(chan string)

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	newClient := &client{
		conn: conn,
		send: make(chan string),
	}
	clients[newClient] = true

	go handleMessages(newClient)

	for {
		_, bytes, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			delete(clients, newClient)
			close(newClient.send)
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

func handleMessages(c *client) {
	for {
		select {
		case message := <-c.send:
			if err := c.conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func ws_screen(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
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
		for {
			if existingPlayer.viewIsDirty {
				if err := conn.WriteMessage(websocket.TextMessage, []byte(printStageFor(existingPlayer))); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	} else {
		fmt.Println("player not found with token: " + token)

	}
}
