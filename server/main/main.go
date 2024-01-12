package main

import (
	"sync"
	"time"

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
	//connectDB()

	//newDoc()
	client := mongoClient()
	collection := client.Database("bloopdb").Collection("users")
	//addToPeople()

	// Only needed once
	//createIndex(client)

	// Create an instance of the Person struct
	person := User{
		Email:     "example@example.com",
		Verified:  true,
		Username:  "exampleuser",
		Hashword:  "hashedpassword",
		CSSClass:  "exampleClass",
		Created:   time.Now(),
		LastLogin: time.Now(),
		// Initialize other fields as required
		Health:    100,
		StageName: "big",
		X:         2,
		Y:         2,
	}
	addUser(collection, person)
	setUserHealth(collection, person.Email)
	//addToUsers(client, person)
	//addToPeople(client)

	/*fmt.Println("Loading data...")
	loadFromJson()

	fmt.Println("Establishing Routes...")
	// Last Handle take priority so dirs in /assets/ will be overwritten by handled funcs
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./assets"))))
	http.HandleFunc("/signin", postSignin)

	fmt.Println("Preparing for interactions...")
	http.HandleFunc("/w", postMovement(moveNorth)) // consider .Methods(http.MethodGet)
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
