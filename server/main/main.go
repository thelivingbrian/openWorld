package main

import (
	"fmt"
	"html/template"
	"net/http"

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

/*
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
*/
