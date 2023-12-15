package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	playerMap   = make(map[string]*Player)
	playerMutex sync.Mutex
	stageMap    = make(map[string]*Stage)
	stageMutex  sync.Mutex
)

/*  This code originally worked, however it does not broadcast.
	Leaving in case useful reference for /screen whos websockets may not need to broadcast

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handler2(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// Read and write messages in a loop
	for {
		// Read message from the WebSocket connection
		messageType, bytes, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}

		// Print the received message to the console
		fmt.Printf("Received message: %s\n", bytes)

		htmlResponse := `
		<form id="form" ws-send>
			<input name="chat_message">
		</form>
		<div id="notifications" hx-swap-oob="beforeend">
			<p>New message received</p>
		</div>
		<div id="chat_room" hx-swap-oob="morphdom">
			<h1>Hello</h1>
		</div>`
		if err := conn.WriteMessage(messageType, []byte(htmlResponse)); err != nil {
			fmt.Println(err)
			return
		}
	}
} */

func main() {
	fmt.Println("Attempting to start server...")

	http.HandleFunc("/home/", getIndex)
	http.Handle("/home/assets/", http.StripPrefix("/home/assets/", http.FileServer(http.Dir("./client/src/assets"))))

	http.HandleFunc("/signin", postSignin)

	http.HandleFunc("/w", postMovement(moveNorth))
	http.HandleFunc("/s", postMovement(moveSouth))
	http.HandleFunc("/a", postMovement(moveWest))
	http.HandleFunc("/d", postMovement(moveEast))
	//http.HandleFunc("/screen", postPlayerScreen)
	http.HandleFunc("/screen", ws_screen)
	http.HandleFunc("/spaceOn", postSpaceOn)
	http.HandleFunc("/spaceOff", postSpaceOff)

	http.HandleFunc("/chat", handler)

	go func() {
		for {
			message := <-broadcast
			sendMessageToAll(websocket.TextMessage, []byte(message))
		}
	}()

	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
