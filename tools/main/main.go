package main

import (
	"fmt"
	"net/http"
)

/*
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
*/

func main() {
	fmt.Println("Attempting to start server...")
	populateFromJson()

	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./level/assets"))))

	http.HandleFunc("/materialPage", getMaterialPage)

	http.HandleFunc("/getEditMaterial", getEditMaterial)
	http.HandleFunc("/editMaterial", editMaterial)
	http.HandleFunc("/getNewMaterial", getNewMaterial)
	http.HandleFunc("/newMaterial", newMaterial)
	http.HandleFunc("/exampleMaterial", exampleMaterial)

	http.HandleFunc("/getEditColor", getEditColor)
	http.HandleFunc("/editColor", editColor)
	http.HandleFunc("/getNewColor", getNewColor)
	http.HandleFunc("/newColor", newColor)
	http.HandleFunc("/exampleSquare", exampleSquare)

	http.HandleFunc("/outputIngredients", outputIngredients)

	http.HandleFunc("/areaPage", getCreateArea)
	http.HandleFunc("/createGrid", createGrid)
	http.HandleFunc("/clickOnSquare", clickOnSquare)
	http.HandleFunc("/selectMaterial", selectMaterial)
	http.HandleFunc("/saveArea", saveArea)
	http.HandleFunc("/editAreaPage", getEditAreaPage)
	http.HandleFunc("/edit", edit)
	http.HandleFunc("/editTransports", editTransports)
	http.HandleFunc("/editTransport", editTransport)
	http.HandleFunc("/dupeTransport", dupeTransport)
	http.HandleFunc("/deleteTransport", deleteTransport)
	http.HandleFunc("/editDisplay", editDisplay)
	http.HandleFunc("/getEditNeighbors", getEditNeighbors)
	http.HandleFunc("/editNeighbors", editNeighbors)
	http.HandleFunc("/editFromTransport", editFromTransport)

	err := http.ListenAndServe(":4444", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
