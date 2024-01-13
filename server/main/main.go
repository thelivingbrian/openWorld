package main

import (
	"fmt"
	"net/http"
	"sync"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
)

type App struct {
	db *mongo.Database
}

var (
	playerMap   = make(map[string]*Player) // Consider sync.Map
	playerMutex sync.Mutex
	stageMap    = make(map[string]*Stage) // Make a game struct that includes this and has needed handlers
	stageMutex  sync.Mutex
	broadcast   = make(chan string)
)

func main() {
	app := App{mongoClient().Database("bloopdb")}
	//collection := app.db.Collection("users")

	fmt.Println("Loading data...")
	loadFromJson()

	fmt.Println("Establishing Routes...")
	// Last Handle take priority so dirs in /assets/ will be overwritten by handled funcs
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./assets"))))

	// Home Page
	http.HandleFunc("/homesignup", getSignUp)
	http.HandleFunc("/signup", app.postSignUp)
	http.HandleFunc("/homesignin", getSignIn)
	http.HandleFunc("/signin", app.postSignin) // This is a gross overload now, but maybe an alternitave to gorilla mux for verbs

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
}
