package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

// []byte("01234567890123456789012345678944")
var store = sessions.NewCookieStore([]byte("hash-key"), []byte("01234567890123456789012345678944"))

func init() {
	gothic.Store = store
	store.Options = &sessions.Options{
		MaxAge: 60 * 60 * 24,
	}
}

func main() {
	fmt.Println("Initializing...")
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	config := getConfiguration()
	db := createDbConnection(config)
	world := createGameWorld(db)
	loadFromJson()

	fmt.Println("Establishing Routes...")

	// Serve assets
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	http.HandleFunc("/images/", imageHandler)

	// home
	http.HandleFunc("/{$}", world.homeHandler)

	// Account creation and sign in
	//http.HandleFunc("/homesignup", getSignUp)
	//http.HandleFunc("/signup", world.db.postSignUp)
	http.HandleFunc("/homesignin", getSignIn)
	http.HandleFunc("/signin", world.postSignin)
	http.HandleFunc("/play", world.postPlay)
	http.HandleFunc("/new", world.postNew)

	// Oauth
	clientId := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	//sessionSecret := os.Getenv("SESSION_SECRET")
	//fmt.Println(sessionSecret)
	goth.UseProviders(
		google.New(clientId, clientSecret, "http://localhost:9090/callback?provider=google"))
	http.HandleFunc("/auth", auth)
	http.HandleFunc("/callback", db.callback)
	//http.HandleFunc("/profile", profile)

	fmt.Println("Preparing for interactions...")
	http.HandleFunc("/clear", clearScreen)

	fmt.Println("Initiating Websockets...")
	http.HandleFunc("/screen", world.NewSocketConnection)

	fmt.Println("Starting server, listening on port " + config.port)
	//var err error
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

func (world *World) homeHandler(w http.ResponseWriter, r *http.Request) {
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

var tinyTemplate = `
<p>{{.Provider}}:{{.UserID}}</p>
`

var userTemplate = `
<p><a href="/logout/{{.Provider}}">logout</a></p>
<p>Name: {{.Name}} [{{.LastName}}, {{.FirstName}}]</p>
<p>Email: {{.Email}}</p>
<p>NickName: {{.NickName}}</p>
<p>Location: {{.Location}}</p>
<p>AvatarURL: {{.AvatarURL}} <img src="{{.AvatarURL}}"></p>
<p>Description: {{.Description}}</p>
<p>UserID: {{.UserID}}</p>
<p>AccessToken: {{.AccessToken}}</p>
<p>ExpiresAt: {{.ExpiresAt}}</p>
<p>RefreshToken: {{.RefreshToken}}</p>
`
