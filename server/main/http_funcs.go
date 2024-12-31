package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/markbates/goth/gothic"
)

/////////////////////////////////////////////
// User Signin and Creation

func (world *World) postPlay(w http.ResponseWriter, r *http.Request) {
	id, ok := getUserIdFromSession(r)
	if !ok {
		tmpl.ExecuteTemplate(w, "homepage", false)
		return
	}
	userRecord := world.db.getAuthorizedUserById(id)
	if userRecord == nil {
		// deeply confusing
		// Could imply hacked cookie?
		// Has happened when db record is lost/destroyed
		return
	}

	fmt.Println("have user")

	if userRecord.Username == "" {
		fmt.Println("no name")
		io.WriteString(w, chooseYourColor())
	} else {
		record, err := world.db.getPlayerRecord(userRecord.Username)
		if err != nil {
			log.Fatal("No player found for user") // Too extreme.
		}
		loginRequest := createLoginRequest(record)
		world.addIncoming(loginRequest)

		fmt.Println("loginRequest for: " + loginRequest.Record.Username)
		tmpl.ExecuteTemplate(w, "player-page", loginRequest)
	}
}

func (world *World) postNew(w http.ResponseWriter, r *http.Request) {
	id, ok := getUserIdFromSession(r)
	if !ok {
		tmpl.ExecuteTemplate(w, "homepage", false)
		return
	}
	userRecord := world.db.getAuthorizedUserById(id)
	if userRecord == nil {
		// deeply confusing
		// Could imply hacked cookie?
		return
	}

	fmt.Println("have user")

	props, ok := requestToProperties(r)
	if !ok {
		fmt.Println("Invalid properties")
		tmpl.ExecuteTemplate(w, "homepage", false)
		return
	}

	team := props["player-team"]
	username := props["player-name"]

	fmt.Println(team)
	fmt.Println(username)

	if !validTeam(team) {
		io.WriteString(w, divBottomInvalid("Invalid Player Color"))
		return
	}

	if world.db.foundUsername(username) {
		io.WriteString(w, divBottomInvalid("Username unavailable. Try again."))
		return
	}

	// new method
	record := PlayerRecord{
		Username:  username,
		Team:      team,
		Trim:      "",
		Health:    100,
		StageName: "tutorial:0-0",
		X:         4,
		Y:         4,
		Money:     80,
	}

	err := world.db.InsertPlayerRecord(record)
	if err != nil {
		io.WriteString(w, divBottomInvalid("Error saving new player"))
		return
	}
	ok = world.db.updateUsernameForUserWithId(id, username)
	if !ok {
		io.WriteString(w, divBottomInvalid("Error, username not updated"))
		return
	}

	loginRequest := createLoginRequest(record)
	world.addIncoming(loginRequest)

	tmpl.ExecuteTemplate(w, "player-page", loginRequest)
}

func validTeam(team string) bool {
	validTeams := []string{"fuchsia", "sky-blue"}
	for i := range validTeams {
		if validTeams[i] == team {
			return true
		}
	}
	return false
}

func getUserIdFromSession(r *http.Request) (string, bool) {
	session, err := store.Get(r, "user-session")
	if err != nil {
		fmt.Println("Error with session: ")
		fmt.Println(err)
		return "", false
	}

	id, ok := session.Values["identifier"].(string)
	if !ok {
		return "", false
	}
	return id, true
}

/////////////////////////////////////////////
//  Oauth

func auth(w http.ResponseWriter, r *http.Request) {
	/*
		 // Force Google to show account selection
		q := r.URL.Query()
		q.Add("prompt", "select_account")
		r.URL.RawQuery = q.Encode()
	*/
	gothic.BeginAuthHandler(w, r)
}

func (db *DB) callback(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		// This should fail for random additional requests,
		// other routes will be able to grab a pre-existing session
		fmt.Println("Callback error: " + err.Error())
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	fmt.Println("New Sign in from: " + user.Email)
	if user.UserID == "" || user.Provider == "" {
		fmt.Printf("Invalid id: %s or provider %s ", user.UserID, user.Provider)
	}
	identifier := user.Provider + ":" + user.UserID

	userRecord := db.getAuthorizedUserById(identifier)
	if userRecord == nil {
		fmt.Println("Creating new user with identifier: " + identifier)
		newUser := AuthorizedUser{Identifier: identifier, Username: "", Created: time.Now(), LastLogin: time.Now()}
		err := db.insertAuthorizedUser(newUser)
		if err != nil {
			fmt.Println("New User creation in mongo failed")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}

	session, err := store.Get(r, "user-session")
	if err != nil {
		fmt.Println("Error getting new session?")
	}
	session.Values["identifier"] = identifier
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound) // redirects.
}

/////////////////////////////////////////////
// Game Controls

func clearScreen(w http.ResponseWriter, r *http.Request) {
	output := `<div id="screen" class="grid">
				
	</div>`
	io.WriteString(w, output)
}

/////////////////////////////////////////////
// Old sign in (Still used for testing)

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
		record, err := world.db.getPlayerRecord(user.Username)
		if err != nil {
			log.Fatal("No player found for user") // Too extreme.
		}
		loginRequest := createLoginRequest(record)
		world.addIncoming(loginRequest)

		tmpl.ExecuteTemplate(w, "player-page", loginRequest)
	} else {
		io.WriteString(w, invalidSignin())
		return
	}
}

func signInPage() string {
	return `
	<form hx-post="/signin" hx-trigger="click from:#link_submit, keydown[key=='Enter']" hx-target="#landing">
		<div>
			<label>Email:</label><br />
			<input class="retro-input" type="text" name="email" value=""><br />
			<label>Password:</label><br />
			<input class="retro-input" type="text" name="password" value=""><br />
			<a id="link_submit" href="#">Submit</a><br />
			<a id="link_back" href="/">Back</a>
		</div>
	</form>
	`
}

func invalidSignin() string {
	return `
	<form hx-post="/signin" hx-trigger="click from:#link_submit, keydown[key=='Enter']" hx-target="#landing">
		<div>
			<p style='color:red'> Invalid Sign-in. </p>
			<label>Email:</label>
			<input class="retro-input" type="text" name="email" value=""><br />
			<label>Password:</label>
			<input class="retro-input" type="text" name="password" value=""><br />
			<a id="link_submit" href="#">Submit</a><br />
			<a id="link_back" href="/">Back</a>
		</div>
	</form>
	`
}
