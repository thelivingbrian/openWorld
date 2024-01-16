package main

import (
	"fmt"
	"net/http"
	"sync"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
)

type DB struct {
	users         *mongo.Collection
	playerRecords *mongo.Collection
	events        *mongo.Collection
}

type World struct {
	db           DB
	worldPlayers map[string]*Player // Consider sync.Map
	wPlayerMutex sync.Mutex
	worldStages  map[string]*Stage
	wStageMutex  sync.Mutex
}

func main() {
	mongodb := mongoClient().Database("bloopdb")
	db := DB{mongodb.Collection("users"), mongodb.Collection("players"), mongodb.Collection("events")}
	world := World{db: db, worldPlayers: make(map[string]*Player), worldStages: make(map[string]*Stage)}

	fmt.Println("Loading data...")
	loadFromJson()

	fmt.Println("Establishing Routes...")
	// Last Handle take priority so dirs in /assets/ will be overwritten by handled funcs
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./assets"))))

	// Home Page
	http.HandleFunc("/homesignup", getSignUp)
	http.HandleFunc("/signup", db.postSignUp)
	http.HandleFunc("/homesignin", getSignIn)
	http.HandleFunc("/signin", world.postSignin)

	fmt.Println("Preparing for interactions...")
	http.HandleFunc("/clear", clearScreen)

	fmt.Println("Initiating Websockets...")
	http.HandleFunc("/screen", world.NewSocketConnection)

	port := ":9090"
	fmt.Println("Starting server, listen on port " + port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
