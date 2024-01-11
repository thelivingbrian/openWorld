package main

import (
	"sync"

	_ "github.com/lib/pq"
)

var (
	playerMap   = make(map[string]*Player) // Consider sync.Map
	playerMutex sync.Mutex
	stageMap    = make(map[string]*Stage)
	stageMutex  sync.Mutex
	broadcast   = make(chan string)
)

func main() {
	connectDB()
	/*fmt.Println("Loading data...")
	loadFromJson()

	fmt.Println("Establishing Routes...")
	// Last Handle take priority so dirs in /assets/ will be overwritten by handled funcs
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./assets"))))
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
	http.HandleFunc("/screen", ws_screen)

	port := ":9090"
	fmt.Println("Starting server, listen on port " + port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
	*/
}
