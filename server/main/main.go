package main

import (
	"fmt"
	"html/template"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

var store *sessions.CookieStore
var tmpl = template.Must(template.ParseGlob("templates/*.tmpl.html"))

func main() {
	fmt.Println("Initializing...")
	config := getConfiguration()

	fmt.Println("Configuring session storage...")
	store = config.createCookieStore()
	gothic.Store = store
	goth.UseProviders(google.New(config.googleClientId, config.googleClientSecret, config.googleCallbackUrl))

	fmt.Println("Initializing database connection..")
	db := createDbConnection(config)

	if pProfEnabled() {
		go initiatePProf()
	}

	fmt.Println("Establishing Routes...")
	mux := http.NewServeMux()

	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	mux.HandleFunc("/images/", imageHandler)

	if config.isHub {
		// Home
		fmt.Println("Setting up hub...")
		mux.HandleFunc("/{$}", homeHandler)

		// Oauth
		mux.HandleFunc("/auth", auth)
		mux.HandleFunc("/callback", db.callback)

		// Select World
		mux.HandleFunc("/worlds", createWorldSelectHandler(config))
		mux.HandleFunc("/unavailable", unavailable)

		// New Account
		mux.HandleFunc("/new", db.postNew)
	}

	if config.isServer() {
		fmt.Println("Starting game world...")
		world := createGameWorld(db, config)
		loadFromJson()

		// Process Logouts, should remove global.
		go processLogouts(playersToLogout)

		// World status and play
		mux.HandleFunc("/status", world.statusHandler)
		mux.HandleFunc("/play", world.playHandler)

		// Historical
		mux.HandleFunc("/homesignin", getSignIn)
		mux.HandleFunc("/signin", world.postSignin)

		fmt.Println("Preparing for interactions...")
		mux.HandleFunc("/clear", clearScreen)
		mux.HandleFunc("/insert", world.postHorribleBypass)
		mux.HandleFunc("/stats", world.getStats)

		fmt.Println("Initiating Websockets...")
		mux.HandleFunc("/screen", world.NewSocketConnection)
	}

	fmt.Println("Starting server, listening on port " + config.port)
	var err error
	if config.usesTLS {
		err = http.ListenAndServeTLS(config.port, config.tlsCertPath, config.tlsKeyPath, mux)
	} else {
		err = http.ListenAndServe(config.port, mux)
	}
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}

///////////////////////////////////////////////////////
// Pprof

func pProfEnabled() bool {
	rawValue := os.Getenv("PPROF_ENABLED")
	featureEnabled, err := strconv.ParseBool(rawValue)
	if err != nil {
		fmt.Printf("Error parsing PPROF_ENABLED: %v. Defaulting to false.\n", err)
		return false
	}
	return featureEnabled
}

func initiatePProf() {
	fmt.Println("Starting pprof HTTP server on :6060")
	fmt.Println(http.ListenAndServe("localhost:6060", nil))
}

////////////////////////////////////////////////////////
// Homepage

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Home page accessed.")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, identifierFound := getUserIdFromSession(r)
	tmpl.ExecuteTemplate(w, "homepage", identifierFound)
}
