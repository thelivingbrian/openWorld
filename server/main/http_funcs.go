package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

/////////////////////////////////////////////
// User Creation

// Unused?
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
			return
		} else {
			io.WriteString(w, "<h2>Invalid (User logged in already)</h2>")
			return
		}
	} else {
		io.WriteString(w, invalidSignin())
		return
	}
}

func (world *World) postPlay(w http.ResponseWriter, r *http.Request) {
	id, ok := getUserIdFromSession(r)
	if !ok {
		io.WriteString(w, homepage)
		return
	}
	userRecord := world.db.getAuthorizedUserById(id)
	if userRecord == nil {
		// deeply confusing
		// Could imply hacked cookie?
		return
	}

	fmt.Println("have user")

	if userRecord.Username == "" {
		fmt.Println("no name")
		io.WriteString(w, chooseYourColor())
	} else {
		record, err := world.db.getPlayerRecord(userRecord.Username)
		if err != nil {
			log.Fatal("No player found for user") // lol too extreme
		}
		player := world.join(record)
		if player != nil {
			io.WriteString(w, printPageFor(player))
			return
		} else {
			io.WriteString(w, "<h2>Invalid (User logged in already)</h2>")
			return
		}
	}
}

func (world *World) postNew(w http.ResponseWriter, r *http.Request) {
	id, ok := getUserIdFromSession(r)
	if !ok {
		io.WriteString(w, homepage)
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
		io.WriteString(w, homepage)
		return
	}

	color := props["player-color"]
	username := props["player-name"]

	fmt.Println(color)
	fmt.Println(username)

	if !world.db.usernameExists(username) {
		if !validPlayerColor(color) {
			io.WriteString(w, divBottomInvalid("Invalid Player Color"))
			return

		}
		record := PlayerRecord{
			Username:  username,
			Color:     color,
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
		ok := world.db.updateUserName(id, username)
		if !ok {
			io.WriteString(w, divBottomInvalid("Error, username not updated"))
			return
		}

		player := world.join(&record)
		if player != nil {
			io.WriteString(w, printPageFor(player))
		} else {
			io.WriteString(w, "<h2>Invalid (User logged in already)</h2>") // Should be impossible?
		}
	} else {
		io.WriteString(w, divBottomInvalid("Username unavailable. Try again."))
		return
	}
}

func validPlayerColor(color string) bool {
	validColors := []string{"fusia", "sky-blue"}
	for i := range validColors {
		if validColors[i] == color {
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
// Game Controls

func clearScreen(w http.ResponseWriter, r *http.Request) {
	output := `<div id="screen" class="grid">
				
	</div>`
	io.WriteString(w, output)
}

/////////////////////////////////////////////
// Old Signup / sign in (Still used for testing?)

func signUpPage() string {

	// Trigger back link with backspace
	/*return `
	<form hx-post="/signup" hx-trigger="click from:#link_submit, keydown[key=='Enter']" hx-target="#landing"">
		<div>
			<label>Email:</label>
			<input class="retro-input" type="text" name="email" value=""><br />
			<label>Username:</label>
			<input class="retro-input" type="text" name="username" value=""><br />
			<label>Password:</label>
			<input class="retro-input" type="text" name="password" value=""><br />
			<a id="link_submit" href="#">Submit</a><br />
			<a id="link_back" href="/">Back</a>
		</div>
	</form>
	`
	*/
	return "<h3>Under development</h3>"
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

func (db *DB) newUser(email string, username string, hashword string) string {
	if !isEmailValid(email) {
		return invalidEmailHTML() // Use template to avoid duplication
	}
	user := User{Email: email, Verified: true, Username: username, Hashword: hashword, Created: time.Now()}
	err := db.newAccount(user)
	if err != nil {
		return failedToCreateHTML()
	}
	return "<h1>Success</h1>"
}

func failedToCreateHTML() string {
	return `
	<h2> Username or Email unavailable  </h2>
	<form hx-post="/signup" hx-trigger="click from:#link_submit, keydown[key=='Enter']" hx-target="#landing">
		<div>
			<label>Email:</label>
			<input class="retro-input" type="text" name="email" value=""><br />
			<label>Username:</label>
			<input class="retro-input" type="text" name="username" value=""><br />
			<label>Password:</label>
			<input class="retro-input" type="text" name="password" value=""><br />
			<a id="link_submit" href="#">Submit</a><br />
			<a id="link_back" href="/">Back</a>
		</div>
	</form>
	`
}

func invalidEmailHTML() string {
	return `
	<h3 style='color:red'> Invalid Email. </h3>
	<form hx-post="/signup" hx-trigger="click from:#link_submit, keydown[key=='Enter']" hx-target="#landing">
		<div>
			<label>Email:</label>
			<input class="retro-input" type="text" name="email" value=""><br />
			<label>Username:</label>
			<input class="retro-input" type="text" name="username" value=""><br />
			<label>Password:</label>
			<input class="retro-input" type="text" name="password" value=""><br />
			<a id="link_submit" href="#">Submit</a><br />
			<a id="link_back" href="/">Back</a>
		</div>
	</form>
	`
}

func passwordTooShortHTML() string {
	return `
	<form hx-post="/signup" hx-trigger="click from:#link_submit, keydown[key=='Enter']" hx-target="#landing">
		<div>
			<p style='color:red'> Password must have 8 characters. </p>
			<label>Email:</label>
			<input class="retro-input" type="text" name="email" value=""><br />
			<label>Username:</label>
			<input class="retro-input" type="text" name="username" value=""><br />
			<label>Password:</label>
			<input class="retro-input" type="text" name="password" value=""><br />
			<a id="link_submit" href="#">Submit</a><br />
			<a id="link_back" href="/">Back</a>
		</div>
	</form>
	`
}

func isEmailValid(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(email)
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
