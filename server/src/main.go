package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func getHello(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /hello request\n")
	fmt.Printf(r.Method)
	button := `<button hx-post="/bye"
                        hx-trigger="click"
                        hx-target="#parent-div"
                        hx-swap="innerHTML">
                        Goodbye!
                 </button>`
	io.WriteString(w, button)
}

func getBye(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /bye request\n")
	button := `<button hx-post="/hello"
        hx-trigger="click"
        hx-target="#parent-div"
        hx-swap="innerHTML">
        Hello!
 </button>`
	io.WriteString(w, button)
}

func postActivate(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /activate request\n")
	fmt.Printf(r.Method)
	fmt.Printf("\n")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	bodyS := string(body[:])
	input := strings.Split(bodyS, "&")
	id := input[0]
	token := strings.Split(input[1], "=")[1]

	fmt.Printf(bodyS)
	fmt.Printf("\n")
	fmt.Printf(id)
	fmt.Printf("\n")
	fmt.Printf(token)
	fmt.Printf("\n")

	resp := `<div class="grid-square ` + token + `" id="c2-3"></div>`
	io.WriteString(w, resp)
}

func getScreen(w http.ResponseWriter, r *http.Request) {
	stage := Stage{
		tiles: [][]Tile{
			{Tile{""}, Tile{"blue"}, Tile{"red"}, Tile{"green"}, Tile{""}, Tile{""}},
			{Tile{""}, Tile{""}, Tile{""}, Tile{"red"}, Tile{"green"}, Tile{"red"}},
			{Tile{"green"}, Tile{""}, Tile{"red"}, Tile{""}, Tile{"blue"}, Tile{""}},
			{Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}},
			{Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{""}},
		},
	}
	io.WriteString(w, stage.printStage())
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./client/src")
}

func main() {
	fmt.Println("Attempting to start server...")

	http.HandleFunc("/home/", getIndex)
	http.Handle("/home/assets/", http.StripPrefix("/home/assets/", http.FileServer(http.Dir("./client/src/assets"))))

	http.HandleFunc("/hello", getHello)
	http.HandleFunc("/bye", getBye)
	http.HandleFunc("/activate", postActivate)
	http.HandleFunc("/screen", getScreen)

	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
