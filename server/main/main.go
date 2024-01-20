package main

import (
	"fmt"
	"net/http"
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

	// Account creation and sign in
	http.HandleFunc("/homesignup", getSignUp)
	http.HandleFunc("/signup", world.db.postSignUp)
	http.HandleFunc("/homesignin", getSignIn)
	http.HandleFunc("/signin", world.postSignin)

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
