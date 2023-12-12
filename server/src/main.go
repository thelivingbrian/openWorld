package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

var (
	playerMap   = make(map[string]*Player)
	playerMutex sync.Mutex
	stageMap    = make(map[string]*Stage)
	stageMutex  sync.Mutex
)

func getIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./client/src")
}

func postSignin(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	bodyS := string(body[:])
	input := strings.Split(bodyS, "&")
	token := strings.Split(input[0], "=")[1]
	stage := strings.Split(input[1], "=")[1]

	playerMutex.Lock()
	existingPlayer, playerExists := playerMap[token]
	playerMutex.Unlock()

	if !playerExists {
		fmt.Println("New Player")
		newPlayer := &Player{
			id:          token,
			stage:       nil,
			stageName:   stage,
			viewIsDirty: true,
			x:           2,
			y:           2,
		}

		playerMutex.Lock()
		defer playerMutex.Unlock()
		playerMap[token] = newPlayer
		existingPlayer = newPlayer
	}

	// Player with the given token exists
	existingStageName := existingPlayer.stageName

	stageMutex.Lock()
	existingStage, stageExists := stageMap[existingStageName]
	if !stageExists {
		fmt.Println("New Stage")
		// If the Stage doesn't exist, create a new one and store it in the map
		newStage := getStageByName(stage)
		stagePtr := &newStage
		stageMap[existingStageName] = stagePtr
		existingStage = stagePtr
	}
	stageMutex.Unlock()

	existingPlayer.stage = existingStage
	existingStage.placeOnStage(existingPlayer)

	fmt.Println("Printing Stage")
	io.WriteString(w, printPageHeaderFor(existingPlayer))
}

func playerFromRequest(r *http.Request) (*Player, bool) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v", err)
		return nil, false
	}

	bodyS := string(body[:])
	input := strings.Split(bodyS, "&")
	token := strings.Split(input[0], "=")[1]

	existingPlayer, playerExists := playerMap[token]
	if !playerExists {
		fmt.Println("player not found with token: " + token)
		return nil, false
	}

	return existingPlayer, true
}

func postMovement(f func(*Stage, *Player)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		existingPlayer, success := playerFromRequest(r)
		if !success {
			panic(0)
		}
		currentStage := existingPlayer.stage // This is a bug? Is stage always legit? Login?

		f(currentStage, existingPlayer)

		fmt.Println("moving")
	}
}

func postPlayerScreen(w http.ResponseWriter, r *http.Request) {
	existingPlayer, success := playerFromRequest(r)
	if !success {
		panic(0) // Handle this gracefully
	}
	if existingPlayer.viewIsDirty {
		fmt.Println("View is Dirty")
		existingPlayer.viewIsDirty = false
		io.WriteString(w, printStageFor(existingPlayer))
	} else {
		io.WriteString(w, "")
	}
}

func main() {
	fmt.Println("Attempting to start server...")

	http.HandleFunc("/home/", getIndex)
	http.Handle("/home/assets/", http.StripPrefix("/home/assets/", http.FileServer(http.Dir("./client/src/assets"))))

	http.HandleFunc("/signin", postSignin)

	http.HandleFunc("/hello", getHello)
	http.HandleFunc("/bye", getBye)
	http.HandleFunc("/activate", postActivate) // unused?
	http.HandleFunc("/w", postMovement(moveNorth))
	http.HandleFunc("/s", postMovement(moveSouth))
	http.HandleFunc("/a", postMovement(moveWest))
	http.HandleFunc("/d", postMovement(moveEast))
	http.HandleFunc("/screen", postPlayerScreen)

	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
