package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

var store *sessions.CookieStore

type state struct {
	DB
	store *sessions.CookieStore
}

func main() {
	fmt.Println("Initializing...")
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	fmt.Println("Configuring session storage...")
	config := getConfiguration()
	store = config.createCookieStore()
	store.Options = &sessions.Options{
		MaxAge: 60 * 60 * 24,
	}
	gothic.Store = store
	goth.UseProviders(google.New(config.googleClientId, config.googleClientSecret, config.googleCallbackUrl))

	fmt.Println("Starting game world...")
	db := createDbConnection(config)
	world := createGameWorld(db)
	loadFromJson()

	fmt.Println("Establishing Routes...")

	// Serve assets
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	http.HandleFunc("/images/", imageHandler)

	// home
	http.HandleFunc("/{$}", homeHandler)

	// Account creation and sign in
	http.HandleFunc("/homesignin", getSignIn)
	http.HandleFunc("/signin", world.postSignin)
	http.HandleFunc("/play", world.postPlay)
	http.HandleFunc("/new", world.postNew)

	// Oauth
	http.HandleFunc("/auth", auth)
	http.HandleFunc("/callback", db.callback)

	fmt.Println("Preparing for interactions...")
	http.HandleFunc("/clear", clearScreen)

	fmt.Println("Initiating Websockets...")
	http.HandleFunc("/screen", world.NewSocketConnection)

	fmt.Println("Starting server, listening on port " + config.port)
	if config.usesTLS {
		err = http.ListenAndServeTLS(config.port, config.tlsCertPath, config.tlsKeyPath, nil)
	} else {
		err = http.ListenAndServe(config.port, nil)
	}
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hello from home")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	session, err := store.Get(r, "user-session")
	if err != nil {
		fmt.Println("No session")
		fmt.Println(err)
		io.WriteString(w, homepage)
		return
	}

	_, ok := session.Values["identifier"].(string)
	if !ok {
		fmt.Println("No Identifier")
		io.WriteString(w, homepage)
		return
	}

	io.WriteString(w, homepageSignedin)
}

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

	// store id to the session
	session, err := store.Get(r, "user-session")
	if err != nil {
		fmt.Println("Error getting new session?")
	}
	session.Values["identifier"] = identifier // Additional layer of symmetric encryption here?
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound) // Do I want redirects?
}

// template
var homepage = `
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<script src="/assets/htmx-mod.js"></script>
<script src="/assets/ws.js"></script>
<link rel="stylesheet" href="/assets/style.css">
<link rel="stylesheet" href="/assets/colors.css">
<body>
    <div id="page">
        <div id="logo">
            <img class="logo-img" src="/assets/blooplogo2.webp" width="80%" alt="Welcome to bloopworld"><br />
        </div>
        <div id="landing">
            <a class="large-font" href="/auth?provider=google"> Sign in with Google </a>
        </div>
    </div>
</body>
`

var homepageSignedin = `
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<script src="/assets/htmx-mod.js"></script>
<script src="/assets/ws.js"></script>
<link rel="stylesheet" href="/assets/style.css">
<link rel="stylesheet" href="/assets/colors.css">
<body>
    <div id="page">
        <div id="logo">
            <img class="logo-img" src="/assets/blooplogo2.webp" width="80%" alt="Welcome to bloopworld"><br />
        </div>
        <div id="landing">
			<a class="large-font" href="#" hx-post="/play" hx-target="#landing">Play</a><br />
        </div>
    </div>
</body>
`
