package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	playerMap   = make(map[string]*Player) // Consider sync.Map
	playerMutex sync.Mutex
	stageMap    = make(map[string]*Stage)
	stageMutex  sync.Mutex
	broadcast   = make(chan string)
	updates     = make(chan Update)
)

func main() {
	fmt.Println("Loading data...")
	loadFromJson()

	fmt.Println("Establishing Routes...")
	http.Handle("/home/", http.StripPrefix("/home/", http.FileServer(http.Dir("./assets"))))
	http.HandleFunc("/signin", postSignin)

	fmt.Println("Preparing for interactions...")
	http.HandleFunc("/w", postMovement(moveNorth))
	http.HandleFunc("/s", postMovement(moveSouth))
	http.HandleFunc("/a", postMovement(moveWest))
	http.HandleFunc("/d", postMovement(moveEast))

	http.HandleFunc("/clear", clearScreen)
	http.HandleFunc("/spaceOn", postSpaceOn)
	http.HandleFunc("/spaceOff", postSpaceOff)

	fmt.Println("Initiating Websockets...")
	http.HandleFunc("/screen", ws_screen) // Inject Channels here
	http.HandleFunc("/chat", ws_chat)
	go func() {
		for {
			message := <-broadcast
			sendMessageToAll(websocket.TextMessage, []byte(message))
		}
	}()
	go func() {
		for {
			update := <-updates
			sendUpdate(websocket.TextMessage, update)
		}
	}()

	fmt.Println("Attempting to start server...")
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
