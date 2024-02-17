package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var materials []Material
var areas []Area

func populateMaterialsFromJson() {
	jsonData, err := os.ReadFile("./level/data/materials.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(jsonData, &materials); err != nil {
		panic(err)
	}

	fmt.Printf("Loaded %d material(s).\n", len(materials))
}

func populateAreasFromJson() {
	jsonData, err := os.ReadFile("./level/data/areas.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(jsonData, &areas); err != nil {
		panic(err)
	}

	fmt.Printf("Loaded %d area(s).\n", len(areas))
}

func populateFromJson() {
	populateMaterialsFromJson()
	populateAreasFromJson()
}

func main() {
	fmt.Println("Attempting to start server...")
	populateFromJson()

	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./level/assets"))))

	http.HandleFunc("/materialPage", getMaterialPage)
	http.HandleFunc("/material", getMaterial)
	http.HandleFunc("/materialEdit", materialEdit)
	http.HandleFunc("/newMaterialForm", newMaterialForm)
	http.HandleFunc("/materialNew", materialNew)
	http.HandleFunc("/submit", submit)

	http.HandleFunc("/areaPage", getCreateArea)
	http.HandleFunc("/createGrid", createGrid)
	http.HandleFunc("/clickOnSquare", clickOnSquare)
	http.HandleFunc("/selectMaterial", selectMaterial)
	http.HandleFunc("/exampleSquare", exampleSquare)
	http.HandleFunc("/saveArea", saveArea)
	http.HandleFunc("/editAreaPage", getEditAreaPage)
	http.HandleFunc("/edit", edit)
	http.HandleFunc("/editTransports", editTransports)
	http.HandleFunc("/editTransport", editTransport)
	http.HandleFunc("/editDisplay", editDisplay)
	http.HandleFunc("/editNeighbors", editNeighbors)

	err := http.ListenAndServe(":4444", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
