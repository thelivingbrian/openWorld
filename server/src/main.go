package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Attempting to start server...")

	http.HandleFunc("/home/", getIndex)
	http.Handle("/home/assets/", http.StripPrefix("/home/assets/", http.FileServer(http.Dir("./client/src/assets"))))

	http.HandleFunc("/signin", postSignin)

	http.HandleFunc("/hello", getHello)
	http.HandleFunc("/bye", getBye)
	http.HandleFunc("/w", postMovement(moveNorth))
	http.HandleFunc("/s", postMovement(moveSouth))
	http.HandleFunc("/a", postMovement(moveWest))
	http.HandleFunc("/d", postMovement(moveEast))
	http.HandleFunc("/screen", postPlayerScreen)
	http.HandleFunc("/spaceOn", postSpaceOn)
	http.HandleFunc("/spaceOff", postSpaceOff)

	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
