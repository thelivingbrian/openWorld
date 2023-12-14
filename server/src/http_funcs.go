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

	bodyS := string(body[:])
	input := strings.Split(bodyS, "&")
	token := strings.Split(input[0], "=")[1]
	stage := strings.Split(input[1], "=")[1]

	playerMutex.Lock()
	existingPlayer, playerExists := playerMap[token]
	playerMutex.Unlock()

	if !playerExists {
		fmt.Println("New Player")
		actions := Actions{false}
		newPlayer := &Player{
			id:          token,
			stage:       nil,
			stageName:   stage,
			viewIsDirty: true,
			x:           2,
			y:           2,
			actions:     &actions,
			health:      100,
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

		//fmt.Println("moving")
	}
}

func postSpaceOn(w http.ResponseWriter, r *http.Request) {
	existingPlayer, success := playerFromRequest(r)
	if success {
		existingPlayer.actions.space = true
		existingPlayer.viewIsDirty = true
	} else {
		io.WriteString(w, "")
	}
}

func postSpaceOff(w http.ResponseWriter, r *http.Request) {
	existingPlayer, success := playerFromRequest(r)
	if success {
		existingPlayer.actions.space = false
		existingPlayer.viewIsDirty = true
		existingPlayer.stage.damageAt(applyRelativeDistance(existingPlayer.y, existingPlayer.x, cross()))
		io.WriteString(w, `<input id="spaceOn" hx-post="/spaceOn" hx-trigger="keydown[key==' '] from:body once" type="hidden" name="token" value="`+existingPlayer.id+`" />`)
	} else {
		io.WriteString(w, "")
	}
}

func postPlayerScreen(w http.ResponseWriter, r *http.Request) {
	existingPlayer, success := playerFromRequest(r)
	if !success {
		panic(0) // Handle this gracefully
	}
	if existingPlayer.viewIsDirty {
		//fmt.Println("View is Dirty")
		io.WriteString(w, printStageFor(existingPlayer))
	} else {
		io.WriteString(w, "")
	}
}

func getHello(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /hello request\n")
	fmt.Printf(r.Method)
	button := `<button hx-post="/bye"
                        hx-trigger="click, keyup[key=='Alt'] from:body"
                        hx-target="#parent-div"
                        hx-swap="innerHTML">
                        Goodbye!
                 </button>`
	io.WriteString(w, button)
}

func getBye(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /bye request\n")
	button := `<button hx-post="/hello"
        hx-trigger="click, keyup[key=='Alt'] from:body"
        hx-target="#parent-div"
        hx-swap="innerHTML">
        Hello!
 </button>`
	io.WriteString(w, button)
}
