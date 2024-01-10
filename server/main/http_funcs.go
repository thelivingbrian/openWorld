package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
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

	bodyS := string(body[:]) // Use property-ifier
	input := strings.Split(bodyS, "&")
	token := strings.Split(input[0], "=")[1]
	stage := strings.Split(input[1], "=")[1]

	playerMutex.Lock()
	existingPlayer, playerExists := playerMap[token]
	playerMutex.Unlock()

	if !playerExists {
		fmt.Println("New Player: ")
		newPlayer := &Player{
			id:        token,
			stage:     nil,
			stageName: stage,
			x:         2,
			y:         2,
			actions:   createDefaultActions(),
			health:    100,
		}

		playerMutex.Lock()
		defer playerMutex.Unlock() //sketchy?
		playerMap[token] = newPlayer
		existingPlayer = newPlayer
	}

	existingStageName := existingPlayer.stageName

	existingStage := getStageByName(existingStageName)

	existingPlayer.stage = existingStage

	fmt.Println("Printing Page Headers")
	io.WriteString(w, printPageFor(existingPlayer))
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

func postMovement(f func(*Player)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		existingPlayer, success := playerFromRequest(r)
		if !success {
			fmt.Println("Invalid Request: ")
			fmt.Println(r)
			return
			//panic(0) // This is bad because it means anyone can panic the server
		}
		f(existingPlayer)
	}
}

func postSpaceOn(w http.ResponseWriter, r *http.Request) {
	existingPlayer, success := playerFromRequest(r)
	if success {
		existingPlayer.turnSpaceOn()
	} else {
		io.WriteString(w, "")
	}
}

func postSpaceOff(w http.ResponseWriter, r *http.Request) {
	existingPlayer, success := playerFromRequest(r)
	if success {
		existingPlayer.turnSpaceOff()
		io.WriteString(w, `<input id="spaceOn" hx-post="/spaceOn" hx-trigger="keydown[key==' '] from:body once" type="hidden" name="token" value="`+existingPlayer.id+`" />`)
	} else {
		io.WriteString(w, "")
	}
}

func clearScreen(w http.ResponseWriter, r *http.Request) {
	output := `<div id="screen" class="grid">
				
	</div>`
	io.WriteString(w, output)
}
