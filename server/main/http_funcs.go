package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

/////////////////////////////////////////////
// Home Page / CSS / JS

func getIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./client/src")
}

/////////////////////////////////////////////
// User Creation

func getSignUp(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, signUpPage())
}

func (app *DB) postSignUp(w http.ResponseWriter, r *http.Request) {
	props, success := requestToProperties(r)
	if !success {
		log.Fatal("Failed to retreive properties")
	}
	email, err := url.QueryUnescape(props["email"])
	username, err := url.QueryUnescape(props["username"])
	password, err := url.QueryUnescape(props["password"])
	if err != nil {
		log.Fatal("Unescape failed")

	}
	if len(password) < 8 {
		io.WriteString(w, passwordTooShortHTML())
	}

	hashword, err := hashPassword(props["password"])
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(email)

	io.WriteString(w, app.newUser(email, username, hashword))
}

func getSignIn(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, signInPage())
}

func (world *World) postSignin(w http.ResponseWriter, r *http.Request) {
	props, success := requestToProperties(r)
	if !success {
		log.Fatal("Failed to retreive properties")
	}
	email, err := url.QueryUnescape(props["email"])
	if err != nil {
		log.Fatal("Query unescape failed")
	}
	password, err := url.QueryUnescape(props["password"])
	if err != nil {
		log.Fatal("Password unescape failed")
	}

	user, err := world.db.getUserByEmail(email)
	if err != nil {
		io.WriteString(w, invalidSignin())
		return
	}
	fmt.Printf("Found a user: %+v\n", user.Username)
	worked := checkPasswordHash(password, user.Hashword)

	if worked {
		token := uuid.New().String()
		player, err := world.db.getPlayerRecord(user.Username)
		if err != nil {
			log.Fatal("No player found for user")
		}
		world.join(w, player, token)
	}

}

func (world *World) join(w http.ResponseWriter, record *PlayerRecord, token string) {
	fmt.Println("New Player: " + token)
	newPlayer := &Player{
		id:        token,
		username:  record.Username,
		stage:     nil,
		stageName: record.StageName,
		x:         record.X,
		y:         record.Y,
		actions:   createDefaultActions(),
		health:    record.Health,
		money:     record.Money,
	}

	//playerMutex.Lock()
	//defer playerMutex.Unlock() //sketchy?
	//playerMap[token] = newPlayer
	world.wPlayerMutex.Lock()
	world.worldPlayers[token] = newPlayer
	world.wPlayerMutex.Unlock()

	fmt.Println("Printing Page Headers")
	io.WriteString(w, printPageFor(newPlayer))
}

/////////////////////////////////////////////
// Game Controls

/*func postMovement(f func(*Player)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		existingPlayer, success := playerFromRequest(r)
		if !success {
			fmt.Println("Invalid Request: ")
			fmt.Println(r)
			return
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


func playerFromRequest(r *http.Request) (*Player, bool) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v", err)
		return nil, false
	}

	bodyS := string(body[:]) // Benchmark this vs using proprties
	input := strings.Split(bodyS, "&")
	token := strings.Split(input[0], "=")[1]

	existingPlayer, playerExists := playerMap[token]
	if !playerExists {
		fmt.Println("player not found with token: " + token)
		return nil, false
	}

	return existingPlayer, true
}
*/

func clearScreen(w http.ResponseWriter, r *http.Request) {
	output := `<div id="screen" class="grid">
				
	</div>`
	io.WriteString(w, output)
}
