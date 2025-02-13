package main

import (
	"fmt"
	"html/template"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

var store *sessions.CookieStore
var tmpl = template.Must(template.ParseGlob("templates/*.tmpl.html"))

func main() {
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
	world := createGameWorld(db, config.domainName)
	loadFromJson()

	// Process Loggouts, should remove global ?
	go processLogouts(playersToLogout)

	// start pprof
	if pProfEnabled() {
		go initiatePProf()
	}

	fmt.Println("Establishing Routes...")
	mux := http.NewServeMux()

	// Serve assets
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	mux.HandleFunc("/images/", imageHandler)

	// home
	mux.HandleFunc("/{$}", homeHandler)

	// Account creation and sign in
	mux.HandleFunc("/homesignin", getSignIn)
	mux.HandleFunc("/signin", world.postSignin)
	mux.HandleFunc("/worlds", world.getWorlds)
	mux.HandleFunc("/status", world.getStatus)
	mux.HandleFunc("/play", world.postPlay)
	mux.HandleFunc("/new", world.postNew)

	// Oauth
	mux.HandleFunc("/auth", auth)
	mux.HandleFunc("/callback", db.callback)

	fmt.Println("Preparing for interactions...")
	mux.HandleFunc("/clear", clearScreen)
	mux.HandleFunc("/insert", world.postHorribleBypass)
	mux.HandleFunc("/stats", world.getStats)

	fmt.Println("Initiating Websockets...")
	mux.HandleFunc("/screen", world.NewSocketConnection)

	fmt.Println("Starting server, listening on port " + config.port)
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

var domains = []string{"http://localhost:9090"}

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

func initiatePProf() {
	fmt.Println("Starting pprof HTTP server on :6060")
	fmt.Println(http.ListenAndServe("localhost:6060", nil))
}

func pProfEnabled() bool {
	rawValue := os.Getenv("PPROF_ENABLED")
	featureEnabled, err := strconv.ParseBool(rawValue)
	if err != nil {
		fmt.Printf("Error parsing PPROF_ENABLED: %v. Defaulting to false.\n", err)
		return false
	}
	return featureEnabled
}
