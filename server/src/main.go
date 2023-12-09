package main

import (
	"fmt"
	"io"
	"net/http"
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

func getIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./client/src")
}

func main() {
	fmt.Println("Attempting to start server...")
	http.HandleFunc("/hello", getHello)
	http.HandleFunc("/bye", getBye)
	http.HandleFunc("/", getIndex)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./client/src/assets"))))
	//err := http.ListenAndServe(":9090", http.FileServer(http.Dir("./client/src")))
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
