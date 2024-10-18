package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

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
	fmt.Println(clientId)
	goth.UseProviders(
		google.New(clientId, clientSecret, "http://localhost:9090/callback/google"))
	http.HandleFunc("/auth/google", auth)
	http.HandleFunc("/callback/google", callback)

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
	gothic.BeginAuthHandler(w, r)
}

func callback(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	fmt.Println(user.Email)
}
