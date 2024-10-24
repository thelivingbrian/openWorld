package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
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
	//w.Header().Set("Content-Type", "text/html; charset=utf-8")
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
		record, err := world.db.getPlayerRecord(user.Username)
		if err != nil {
			log.Fatal("No player found for user") // lol too extreme
		}
		player := world.join(record)
		if player != nil {
			io.WriteString(w, printPageFor(player))
		} else {
			io.WriteString(w, "<h2>Invalid (User logged in already)</h2>")
		}
	} else {
		io.WriteString(w, invalidSignin())
		return
	}
}

func (world *World) postResume(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "user-session")
	if err != nil {
		io.WriteString(w, homepage)
		return
	}

	id, ok := session.Values["identifier"].(string)
	if !ok {
		io.WriteString(w, homepage)
		return
	}
	userRecord, err := world.db.getAuthorizedUserById(id)
	if userRecord == nil {
		// deeply confusing
	}

	fmt.Println("have user")

	if userRecord.Username == "" {
		fmt.Println("no name")
		io.WriteString(w, `<div id="page" hx-swap-oob="true">Choose your color</div>`)
	} else {
		record, err := world.db.getPlayerRecord(userRecord.Username)
		if err != nil {
			log.Fatal("No player found for user") // lol too extreme
		}
		player := world.join(record)
		if player != nil {
			io.WriteString(w, printPageFor(player))
		} else {
			io.WriteString(w, "<h2>Invalid (User logged in already)</h2>")
		}

		io.WriteString(w, homepageSignedin)

	}

}

/////////////////////////////////////////////
// Game Controls

func clearScreen(w http.ResponseWriter, r *http.Request) {
	output := `<div id="screen" class="grid">
				
	</div>`
	io.WriteString(w, output)
}
