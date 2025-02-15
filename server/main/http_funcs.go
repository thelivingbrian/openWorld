package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/markbates/goth/gothic"
)

const ALLOWED_HEADERS = "Content-Type, hx-current-url, HX-Request, HX-Target, HX-Trigger"

// ///////////////////////////////////////////
// World Select and Status

func createWorldSelectHandler(config *Configuration) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := getUserIdFromSession(r)
		if !ok {
			tmpl.ExecuteTemplate(w, "homepage", false)
			return
		}
		tmpl.ExecuteTemplate(w, "world-select", config.domains)
	}
}

func (world *World) statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", world.config.originForCORS())
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", ALLOWED_HEADERS)
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method == http.MethodGet {
		world.getStatus(w, r)
		return
	}
}

func (world *World) getStatus(w http.ResponseWriter, r *http.Request) {
	_, ok := getUserIdFromSession(r)
	if !ok {
		io.WriteString(w, "<div>Invalid Sign in</div>")
		return
	}
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()
	fuchsiaPlayers := world.teamQuantities["fushsia"]
	skyBluePlayers := world.teamQuantities["sky-blue"]
	s := struct {
		ServerName   string
		DomainName   string
		FuchsiaCount int
		SkyBlueCount int
	}{
		ServerName:   world.config.serverName,
		DomainName:   world.config.domainName,
		FuchsiaCount: fuchsiaPlayers,
		SkyBlueCount: skyBluePlayers,
	}
	tmpl.ExecuteTemplate(w, "world-status", s)
}

var unavailableMessage = `Server unavailable :( <a href="#" hx-get="/worlds" hx-target="#page"> Try again</a>`

func unavailable(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, unavailableMessage)
}

// ///////////////////////////////////////////
// Player Sign-in and Create

func (world *World) playHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", world.config.originForCORS())
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", ALLOWED_HEADERS)
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method == http.MethodPost {
		world.postPlay(w, r)
		return
	}
}

func (world *World) postPlay(w http.ResponseWriter, r *http.Request) {
	id, ok := getUserIdFromSession(r)
	if !ok {
		tmpl.ExecuteTemplate(w, "homepage", false)
		return
	}
	userRecord := world.db.getAuthorizedUserById(id)
	if userRecord == nil {
		// Could imply hacked cookie?
		// Has happened when db record is lost/destroyed
		// Is confusing if this happens because you get a blank page with no explanation
		return
	}

	fmt.Println("have user")

	if userRecord.Username == "" {
		fmt.Println("no username")
		tmpl.ExecuteTemplate(w, "choose-your-color", world.config.domainName)
	} else {
		record, err := world.db.getPlayerRecord(userRecord.Username)
		if err != nil {
			log.Fatal("No player found for user") // Too extreme.
		}
		loginRequest := createLoginRequest(record)
		world.addIncoming(loginRequest)

		s := struct {
			LoginRequest *LoginRequest
			DomainName   string
		}{
			LoginRequest: loginRequest,
			DomainName:   world.config.domainName,
		}

		fmt.Println("loginRequest for: " + loginRequest.Record.Username)
		tmpl.ExecuteTemplate(w, "player-page", s)
	}
}

func (db *DB) postNew(w http.ResponseWriter, r *http.Request) {
	id, ok := getUserIdFromSession(r)
	if !ok {
		tmpl.ExecuteTemplate(w, "homepage", false)
		return
	}
	userRecord := db.getAuthorizedUserById(id)
	if userRecord == nil {
		// deeply confusing
		// Could imply hacked cookie?
		return
	}

	fmt.Println("New player request from: " + userRecord.Username)

	props, ok := requestToProperties(r)
	if !ok {
		fmt.Println("Invalid properties")
		tmpl.ExecuteTemplate(w, "homepage", false)
		return
	}

	team := props["player-team"]
	username := props["player-name"]
	desiredHostUrlEncoded := props["desired-host"]
	desiredHost, err := url.QueryUnescape(desiredHostUrlEncoded)
	if err != nil {
		fmt.Println("Error decoding host:", err)
		return
	}

	if !validTeam(team) {
		io.WriteString(w, divBottomInvalid("Invalid Player Color"))
		return
	}

	if db.foundUsername(username) {
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

	err = db.InsertPlayerRecord(record)
	if err != nil {
		io.WriteString(w, divBottomInvalid("Error saving new player"))
		return
	}
	ok = db.updateUsernameForUserWithId(id, username)
	if !ok {
		io.WriteString(w, divBottomInvalid("Error, username not updated"))
		return
	}

	tmpl.ExecuteTemplate(w, "post-play-on-load", desiredHost)
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

/////////////////////////////////////////////
// Stats

func (world *World) getStats(w http.ResponseWriter, r *http.Request) {
	world.wPlayerMutex.Lock()
	defer world.wPlayerMutex.Unlock()
	out := fmt.Sprintf("World Player Count: %d\n", len(world.worldPlayers))
	for key, val := range world.teamQuantities {
		out += fmt.Sprintf("%s: %d\n", key, val)
	}
	io.WriteString(w, out)
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
		newUser := AuthorizedUser{Identifier: identifier, Username: "", CreationEmail: user.Email, Created: time.Now(), LastLogin: time.Now()}
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
// Integration Endpoint

func (world *World) postHorribleBypass(w http.ResponseWriter, r *http.Request) {
	secret := os.Getenv("AUTO_PLAYER_PASSWORD")
	if secret == "" {
		fmt.Println("Bypass is disabled - but has been requested.")
		return
	}
	props, ok := requestToProperties(r)
	if !ok {
		fmt.Println("invalid props")
	}
	if props["secret"] != secret {
		fmt.Println("Bypass is disabled - but has been requested.")
		return
	}
	countString := props["count"]
	count, err := strconv.Atoi(countString)
	if err != nil {
		fmt.Println("Invalid count")
		return
	}
	username := props["username"]
	stage := props["stagename"]
	team := props["team"]
	tokens := make([]string, 0, count)
	for i := 0; i < count; i++ {
		iStr := strconv.Itoa(i)
		record := PlayerRecord{Username: username + iStr, Health: 50, Y: 6, X: 15, StageName: stage, Team: team, Trim: "white-b thick"}
		loginRequest := createLoginRequest(record)
		world.addIncoming(loginRequest)
		fmt.Println(loginRequest.Token)
		tokens = append(tokens, loginRequest.Token)
	}
	io.WriteString(w, "[\""+strings.Join(tokens, "\",\"")+"\"]")
}

/////////////////////////////////////////////
// Game Controls

func clearScreen(w http.ResponseWriter, r *http.Request) {
	output := `
	<div id="screen" class="grid">
				
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

		s := struct {
			LoginRequest *LoginRequest
			DomainName   string
		}{
			LoginRequest: loginRequest,
			DomainName:   world.config.domainName,
		}
		err = tmpl.ExecuteTemplate(w, "player-page", s)
		if err != nil {
			fmt.Println(err)
		}
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

////////////////////////////////////
// Utilities for templates

func (record PlayerRecord) HeartsFromRecord() string {
	return getHeartsFromHealth(record.Health)
}
