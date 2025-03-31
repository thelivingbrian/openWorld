package main

import (
	"html/template"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/rs/zerolog"
)

var tmpl = template.Must(template.ParseGlob("templates/*.tmpl.html"))
var logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
var store *sessions.CookieStore

func main() {
	logger.Info().Msg("Initializing...")
	config := getConfiguration()
	setGlobalLogLevel(config.logLevel)

	logger.Info().Msg("Configuring session storage...")
	store = config.createCookieStore()
	gothic.Store = store
	goth.UseProviders(google.New(config.googleClientId, config.googleClientSecret, config.googleCallbackUrl))

	logger.Info().Msg("Initializing database connection..")
	db := createDbConnection(config)

	if pProfEnabled() {
		go initiatePProf()
	}

	logger.Info().Msg("Establishing Routes...")
	mux := http.NewServeMux()

	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	mux.HandleFunc("/images/", imageHandler)

	if config.isHub {
		// Home
		logger.Info().Msg("Setting up hub...")
		mux.HandleFunc("/{$}", homeHandler)
		mux.HandleFunc("/about", aboutHandler)

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
		logger.Info().Msg("Starting game world...")
		world := createGameWorld(db, config)
		loadFromJson()

		// Process Logouts, should remove global.
		// go processLogouts(world.playersToLogout)

		// World status and play
		mux.HandleFunc("/status", world.statusHandler)
		mux.HandleFunc("/play", world.playHandler)

		// Historical
		mux.HandleFunc("/homesignin", getSignIn)
		mux.HandleFunc("/signin", world.postSignin)

		logger.Info().Msg("Preparing for interactions...")
		mux.HandleFunc("/clear", clearScreen) // would need to be in hub ?
		mux.HandleFunc("/insert", world.postHorribleBypass)
		mux.HandleFunc("/stats", world.getStats)

		logger.Info().Msg("Initiating Websockets...")
		mux.HandleFunc("/screen", world.NewSocketConnection)
	}

	logger.Info().Msg("Starting server, listening on port " + config.port)
	var err error
	if config.usesTLS {
		err = http.ListenAndServeTLS(config.port, config.tlsCertPath, config.tlsKeyPath, mux)
	} else {
		err = http.ListenAndServe(config.port, mux)
	}
	if err != nil {
		logger.Error().Err(err).Msg("Failed to start server")
		return
	}
}

///////////////////////////////////////////////////////
// Pprof

func pProfEnabled() bool {
	rawValue := os.Getenv("PPROF_ENABLED")
	featureEnabled, err := strconv.ParseBool(rawValue)
	if err != nil {
		logger.Error().Err(err).Msg("Error parsing PPROF_ENABLED: %v. Defaulting to false.")
		return false
	}
	return featureEnabled
}

func initiatePProf() {
	logger.Info().Msg("Starting pprof HTTP server on :6060")
	logger.Error().Err(http.ListenAndServe("localhost:6060", nil)).Msg("Failed to start Pprof")
}

////////////////////////////////////////////////////////
// Homepage

func homeHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info().Msg("Home page accessed.")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, identifierFound := getUserIdFromSession(r)
	tmpl.ExecuteTemplate(w, "homepage", identifierFound)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info().Msg("Home page accessed.")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl.ExecuteTemplate(w, "about", nil)
}
