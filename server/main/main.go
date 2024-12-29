package main

import (
	"fmt"
	"html/template"
	"net/http"
	_ "net/http/pprof"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

var store *sessions.CookieStore
var tmpl = template.Must(template.ParseGlob("templates/*.tmpl.html"))

func main() {
	go processLogouts(playersToLogout)

	fmt.Println("Initializing...")
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	config := getConfiguration()

	fmt.Println("Configuring session storage...")
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

	// start pprof
	go func() {
		fmt.Println("Starting pprof HTTP server on :6060")
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

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
	http.HandleFunc("/insert", world.postHorribleBypass)

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
	fmt.Println("Home page accessed.")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	session, err := store.Get(r, "user-session")
	if err != nil {
		tmpl.ExecuteTemplate(w, "homepage", false)
		return
	}
	_, ok := session.Values["identifier"].(string)
	if !ok {
		fmt.Println("No Identifier")
		tmpl.ExecuteTemplate(w, "homepage", false)
		return
	}

	tmpl.ExecuteTemplate(w, "homepage", true)
}
