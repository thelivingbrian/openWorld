package main

import (
	"fmt"
	"net/http"
	"os"
	"text/template"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

var store = sessions.NewCookieStore([]byte("very-secret-hash-key-abc1234567890"))

func init() {
	gothic.Store = store
	store.Options = &sessions.Options{
		MaxAge: 60,
	}
}

func main() {
	fmt.Println("Initializing...")
	config := getConfiguration()
	db := createDbConnection(config)
	world := createGameWorld(db)
	loadFromJson()

	fmt.Println("Establishing Routes...")

	// Serve assets
	// Last Handle takes priority so dirs in /assets/ will be overwritten by handled funcs
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./assets"))))
	http.HandleFunc("/images/", imageHandler)

	// Account creation and sign in
	http.HandleFunc("/homesignup", getSignUp)
	http.HandleFunc("/signup", world.db.postSignUp)
	http.HandleFunc("/homesignin", getSignIn)
	http.HandleFunc("/signin", world.postSignin)

	// Oauth
	clientId := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	sessionSecret := os.Getenv("SESSION_SECRET")
	fmt.Println(sessionSecret)
	goth.UseProviders(
		google.New(clientId, clientSecret, "http://localhost:9090/callback?provider=google"))
	http.HandleFunc("/auth", auth)
	http.HandleFunc("/callback", callback)
	http.HandleFunc("/profile", profile)

	fmt.Println("Preparing for interactions...")
	http.HandleFunc("/clear", clearScreen)

	fmt.Println("Initiating Websockets...")
	http.HandleFunc("/screen", world.NewSocketConnection)

	fmt.Println("Starting server, listening on port " + config.port)
	var err error
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
		 // Weirdly stopped being needed
		q := r.URL.Query()
		q.Add("prompt", "select_account")
		r.URL.RawQuery = q.Encode()
	*/
	gothic.BeginAuthHandler(w, r)
}

func callback(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		// This should fail for random additional requests,
		// other routes will be able to grab a pre-existing session
		// so behavior is expected but what triggers the failure?
		fmt.Println("Callback error: " + err.Error())
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	fmt.Println("New Sign in from: " + user.Email)

	// store user to the session
	session, err := store.Get(r, "user-session")
	if err != nil {
		fmt.Println("Error getting new session?")
	}
	session.Values["user"] = user // Map to smaller struct
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusFound)
}

func profile(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, ok := session.Values["user"].(goth.User) // Map to smaller struct
	if !ok {
		fmt.Println("No user in session")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	t, _ := template.New("foo").Parse(tinyTemplate)
	t.Execute(w, user)

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
