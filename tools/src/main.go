package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var materials []Material

func getLevel(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./tools/level")
}

func populateJson() {
	jsonData, err := os.ReadFile("./tools/level/data/materials.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(jsonData, &materials); err != nil {
		panic(err)
	}

	fmt.Println("Got Materials")
	fmt.Println(materials[0].CommonName)

}

func main() {
	fmt.Println("Attempting to start server...")
	populateJson()

	http.HandleFunc("/level/", getLevel)
	http.Handle("/level/assets/", http.StripPrefix("/level/assets/", http.FileServer(http.Dir("./tools/level/assets"))))

	http.HandleFunc("/materialPage", getMaterialPage)
	http.HandleFunc("/material", getMaterial)
	http.HandleFunc("/materialEdit", materialEdit)
	http.HandleFunc("/newMaterialForm", newMaterialForm)
	http.HandleFunc("/materialNew", materialNew)
	http.HandleFunc("/submit", submit)

	http.HandleFunc("/areaPage", getAreaPage)
	http.HandleFunc("/createGrid", createGrid)
	http.HandleFunc("/replace", replaceSquare)
	http.HandleFunc("/select", selectColor)
	http.HandleFunc("/exampleSquare", exampleSquare)
	http.HandleFunc("/saveArea", saveArea)

	err := http.ListenAndServe(":4444", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
