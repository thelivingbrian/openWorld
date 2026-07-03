package main

import (
	"context"
	"html/template"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
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
	platform     *WorldPlatform
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
	var activeWorld *World
	app := App{db: db, config: config, guestLimiter: &GuestLimiter{}}
	if config.platformEnabled || config.mode == "controller" {
		platform, err := newWorldPlatform(context.Background(), db, config)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to initialize world platform")
		}
		app.platform = platform
		platform.register(mux, &app)
	}

	if config.isHub || config.mode == "controller" {
		logger.Info().Msg("Setting up hub...")
		hub := createDefaultHub(db) // rename ?

		// Static Assets
		mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

		// Pages
		mux.HandleFunc("/", app.homeHandler) // "/{$}" end‑of‑path anchor go 1.22
		mux.HandleFunc("/about", aboutHandler)
		mux.HandleFunc("/highscore", hub.highscoreHandler)

		// Oauth
		mux.HandleFunc("/auth", auth)
		mux.HandleFunc("/callback", db.callback)
		mux.HandleFunc("/guests", app.guestsHandler)
		mux.HandleFunc("/signout", signOutHandler)

		// Select World
		mux.HandleFunc("/worlds", app.worldSelectHandler)
		mux.HandleFunc("/unavailable", unavailable)
		mux.HandleFunc("/wrong", somethingWentWrong)

		// New Account
		mux.HandleFunc("/new", db.postNew)
	}

	if config.isServer() {
		logger.Info().Msg("Starting game world...")
		if config.contentDir != "" {
			loadFromDirectory(config.contentDir)
			config.manifest = loadWorldManifest(config.contentDir)
			config.mapAreas = loadWorldMapAreas(config.contentDir)
		} else {
			loadFromJson()
		}
		world := createGameWorld(db, config)
		activeWorld = world
		go periodicSnapshot(world)
		if config.mode == "runtime" && config.worldID != legacyWorldID {
			go periodicRuntimeHeartbeat(world)
		}
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

		// Game Fucntionality
		mux.HandleFunc("/status", world.statusHandler)
		mux.HandleFunc("/play", world.playHandler)
		imageDirectory := "./data/images"
		if config.contentDir != "" {
			imageDirectory = filepath.Join(config.contentDir, "images")
		}
		if !(config.isHub || config.mode == "controller") {
			mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
		}
		mux.HandleFunc("/images/", createImageHandler(imageDirectory)) // note: trailing '/'
		if config.contentDir != "" {
			mux.HandleFunc("/assets/colors.css", func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, filepath.Join(config.contentDir, "colors.css"))
			})
		}

		// REST helper endpoints
		mux.HandleFunc("/insert", world.postHorribleBypass)
		mux.HandleFunc("/stats", world.getStats)
		mux.HandleFunc("/admin", world.adminHandler)
		mux.HandleFunc("/admin/player/update", world.adminUpdatePlayerHandler)
		mux.HandleFunc("/admin/player/kick", world.adminKickPlayerHandler)
		mux.HandleFunc("/admin/player/ban", world.adminBanPlayerHandler)
		mux.HandleFunc("/admin/watch", world.adminWatchPageHandler)
		mux.HandleFunc("/admin/watch/screen", world.adminWatchSocketHandler)

		// Websockets
		logger.Info().Msg("Initiating Websockets...")
		mux.HandleFunc("/screen", world.NewSocketConnection)
	}

	logger.Info().Msg("Starting server, listening on port " + config.port)
	server := &http.Server{Addr: config.port, Handler: mux}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		if activeWorld != nil {
			gracefulWorldShutdown(activeWorld)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if app.platform != nil {
			app.platform.manager.Shutdown(ctx)
		}
		_ = server.Shutdown(ctx)
	}()
	var err error
	if config.usesTLS {
		err = server.ListenAndServeTLS(config.tlsCertPath, config.tlsKeyPath)
	} else {
		err = server.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
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
