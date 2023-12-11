package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

var (
	playerMap   = make(map[string]*Player) // Map to store Player instances with token as key
	playerMutex sync.Mutex                 // Mutex for synchronization when accessing playerMap
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
		// Create a new Player instance
		newPlayer := &Player{
			id:        token,
			stage:     nil,
			stageName: stage,
			x:         2,
			y:         2,
		}

		// Store the new Player in the map with the token as the key
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
	io.WriteString(w, existingStage.printStageFor(existingPlayer))
}

func postw(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	bodyS := string(body[:])
	input := strings.Split(bodyS, "&")
	token := strings.Split(input[0], "=")[1]
	//stage := strings.Split(input[1], "=")[1]

	existingPlayer, playerExists := playerMap[token]
	if !playerExists {
		err := errors.New("player not found with token: " + token)
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	currentStage := existingPlayer.stage

	moveNorth(currentStage, existingPlayer)

	fmt.Println("Printing update")
	io.WriteString(w, existingPlayer.stage.printStageFor(existingPlayer))
}

func posts(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	bodyS := string(body[:])
	input := strings.Split(bodyS, "&")
	token := strings.Split(input[0], "=")[1]
	//stage := strings.Split(input[1], "=")[1]

	existingPlayer, playerExists := playerMap[token]
	if !playerExists {
		err := errors.New("player not found with token: " + token)
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	currentStage := existingPlayer.stage

	moveSouth(currentStage, existingPlayer)

	fmt.Println("Printing update")
	io.WriteString(w, existingPlayer.stage.printStageFor(existingPlayer))
}

func main() {
	fmt.Println("Attempting to start server...")

	http.HandleFunc("/home/", getIndex)
	http.Handle("/home/assets/", http.StripPrefix("/home/assets/", http.FileServer(http.Dir("./client/src/assets"))))

	http.HandleFunc("/signin", postSignin)

	http.HandleFunc("/hello", getHello)
	http.HandleFunc("/bye", getBye)
	http.HandleFunc("/activate", postActivate)
	http.HandleFunc("/w", postw)
	http.HandleFunc("/s", posts)

	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
