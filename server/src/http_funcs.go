package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

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

/*
	func getScreen(w http.ResponseWriter, r *http.Request) {
		stage := Stage{
			tiles: [][]Tile{
				{Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}},
				{Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{"half-gray"}},
				{Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{"half-gray"}},
				{Tile{"half-gray"}, Tile{""}, Tile{""}, Tile{""}, Tile{""}, Tile{"half-gray"}},
				{Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}, Tile{"half-gray"}},
			},
		}
		io.WriteString(w, stage.printStage())
	}
*/
func getHello(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /hello request\n")
	fmt.Printf(r.Method)
	button := `<button hx-post="/bye"
                        hx-trigger="click, keyup[key=='Alt'] from:body"
                        hx-target="#parent-div"
                        hx-swap="innerHTML">
                        Goodbye!
                 </button>`
	io.WriteString(w, button)
}

func getBye(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /bye request\n")
	button := `<button hx-post="/hello"
        hx-trigger="click, keyup[key=='Alt'] from:body"
        hx-target="#parent-div"
        hx-swap="innerHTML">
        Hello!
 </button>`
	io.WriteString(w, button)
}
