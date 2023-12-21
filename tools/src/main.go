package main

import (
	"fmt"
	"net/http"
)

func getLevel(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./tools/level")
}

func main() {
	fmt.Println("Attempting to start server...")

	http.HandleFunc("/level/", getLevel)
	http.Handle("/level/assets/", http.StripPrefix("/level/assets/", http.FileServer(http.Dir("./tools/level/assets"))))

	//http.HandleFunc("/createGrid", createGrid)
	//http.HandleFunc("/new", selectColor)
	//http.HandleFunc("/select", selectColor)

	http.HandleFunc("/materialPage", getMaterialPage)
	http.HandleFunc("/material", getMaterial)
	http.HandleFunc("/materialEdit", materialEdit)
	http.HandleFunc("/newMaterialForm", newMaterialForm)
	http.HandleFunc("/materialNew", materialNew)
	http.HandleFunc("/submit", submit)

	err := http.ListenAndServe(":4444", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
