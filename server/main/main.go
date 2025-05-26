package main

import (
	"html/template"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/rs/zerolog"
)

var tmpl = template.Must(template.ParseGlob("templates/*.tmpl.html"))
var logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
var store *sessions.CookieStore // Move in app?

type App struct {
	db           *DB
	config       *Configuration
	guestLimiter *GuestLimiter
}

type GuestLimiter struct {
	seen sync.Map // key -> time.Time
}

const GUEST_WINDOW = 2 * time.Minute

func (app *App) AllowGuest(key string) bool {
	if app.guestLimiter == nil {
		return true
	}
	now := time.Now()
	if v, ok := app.guestLimiter.seen.Load(key); ok && now.Sub(v.(time.Time)) < GUEST_WINDOW {
		return false
	}
	app.guestLimiter.seen.Store(key, now)
	return true
}

func (app *App) peekPermission(key string) bool {
	if app.guestLimiter == nil {
		return true
	}
	now := time.Now()
	if v, ok := app.guestLimiter.seen.Load(key); ok && now.Sub(v.(time.Time)) < GUEST_WINDOW {
		return false
	}
	return true
}

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

	if config.isHub {
		logger.Info().Msg("Setting up hub...")
		app := App{db, config, &GuestLimiter{}}
		hub := createDefaultHub(db) // rename to LeaderBoards?

		// Static Assets
		mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

		// Pages
		mux.HandleFunc("/", app.homeHandler) // "/{$}" c43bd8b “end‑of‑path” anchor go 1.22
		mux.HandleFunc("/about", aboutHandler)
		mux.HandleFunc("/highscore", hub.highscoreHandler)

		// Oauth
		mux.HandleFunc("/auth", auth)
		mux.HandleFunc("/callback", db.callback)
		mux.HandleFunc("/guests", app.guestsHandler)
		mux.HandleFunc("/signout", signOutHandler)

		// Select World
		mux.HandleFunc("/worlds", createWorldSelectHandler(config))
		mux.HandleFunc("/unavailable", unavailable)
		mux.HandleFunc("/wrong", somethingWentWrong)

		// New Account
		mux.HandleFunc("/new", db.postNew)
	}

	if config.isServer() {
		logger.Info().Msg("Starting game world...")
		world := createGameWorld(db, config)
		go periodicSnapshot(world)
		loadFromJson()

		// Game Fucntionality
		mux.HandleFunc("/status", world.statusHandler)
		mux.HandleFunc("/play", world.playHandler)
		mux.HandleFunc("/images/", imageHandler) // note: trailing '/'

		// REST helper endpoints
		mux.HandleFunc("/insert", world.postHorribleBypass)
		mux.HandleFunc("/stats", world.getStats)

		// Websockets
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
