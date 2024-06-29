package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

/////////////////////////////////////////////
// User Creation

func getSignUp(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, signUpPage())
}

func (db *DB) postSignUp(w http.ResponseWriter, r *http.Request) {
	props, success := requestToProperties(r)
	if !success {
		log.Fatal("Failed to retreive properties")
	}

	email, err := url.QueryUnescape(props["email"])
	emailLowercase := strings.ToLower(email)
	if err != nil {
		log.Fatal("Unescape failed")

	}

	username, err := url.QueryUnescape(props["username"])
	if err != nil {
		log.Fatal("Unescape failed")

	}

	password, err := url.QueryUnescape(props["password"])
	if err != nil {
		log.Fatal("Unescape failed")

	}
	if len(password) < 8 {
		io.WriteString(w, passwordTooShortHTML()) // Use template to avoid duplication, merge error messages into one
		return
	}

	hashword, err := hashPassword(password)
	if err != nil {
		log.Fatal(err)
	}

	// New user creation is disabled
	fmt.Printf("Using variables %s %s %s \n", emailLowercase, username, hashword)
	io.WriteString(w, "<h2>Sorry, Bloop World is currently under developmment.</h2>") //
	//db.newUser(emailLowercase, username, hashword)
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

	user, err := world.db.getUserByEmail(strings.ToLower(email))
	if err != nil {
		io.WriteString(w, invalidSignin())
		return
	}
	fmt.Printf("Found a user: %+v\n", user.Username)

	worked := checkPasswordHash(password, user.Hashword)
	if worked {
		player, err := world.db.getPlayerRecord(user.Username)
		if err != nil {
			log.Fatal("No player found for user")
		}
		world.join(w, player)
	} else {
		io.WriteString(w, invalidSignin())
		return
	}

}

func (world *World) join(w http.ResponseWriter, record *PlayerRecord) {
	token := uuid.New().String()
	fmt.Println("New Player: " + record.Username)
	fmt.Println("Token: " + token)

	if world.isLoggedInAlready(record.Username) {
		fmt.Println("User attempting to log in but is logged in already: " + record.Username)
		io.WriteString(w, "<h2>Invalid (User logged in already)</h2>")
		return
	}

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

	//New Method
	world.wPlayerMutex.Lock()
	world.worldPlayers[token] = newPlayer
	world.wPlayerMutex.Unlock()

	fmt.Println("Printing Page Headers")
	io.WriteString(w, printPageFor(newPlayer)) // interesting...
}

func (world *World) isLoggedInAlready(username string) bool {
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()
	for _, player := range world.worldPlayers {
		if player.username == username {
			return true
		}
	}
	return false
}

/////////////////////////////////////////////
// Game Controls

func clearScreen(w http.ResponseWriter, r *http.Request) {
	output := `<div id="screen" class="grid">
				
	</div>`
	io.WriteString(w, output)
}
