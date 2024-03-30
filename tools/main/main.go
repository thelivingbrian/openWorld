package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Attempting to start server...")
	c := populateFromJson()

	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./assets"))))

	http.HandleFunc("/materialPage", c.getMaterialPage)

	http.HandleFunc("/getEditMaterial", c.getEditMaterial)
	http.HandleFunc("/editMaterial", c.editMaterial)
	http.HandleFunc("/getNewMaterial", getNewMaterial)
	http.HandleFunc("/newMaterial", c.newMaterial)
	http.HandleFunc("/exampleMaterial", exampleMaterial)

	http.HandleFunc("/getEditColor", c.getEditColor)
	http.HandleFunc("/editColor", c.editColor)
	http.HandleFunc("/getNewColor", getNewColor)
	http.HandleFunc("/newColor", c.newColor)
	http.HandleFunc("/exampleSquare", exampleSquare)

	http.HandleFunc("/outputIngredients", c.outputIngredients)

	http.HandleFunc("/clickOnSquare", c.clickOnSquare)
	http.HandleFunc("/selectMaterial", c.selectMaterial)
	http.HandleFunc("/saveArea", c.saveArea)
	http.HandleFunc("/edit", c.getEditArea)
	http.HandleFunc("/editTransports", c.getEditTransports)
	http.HandleFunc("/editTransport", c.editTransport)
	http.HandleFunc("/newTransport", c.newTransport)
	http.HandleFunc("/dupeTransport", c.dupeTransport)
	http.HandleFunc("/deleteTransport", c.deleteTransport)
	http.HandleFunc("/editDisplay", editDisplay)
	http.HandleFunc("/getEditNeighbors", c.getEditNeighbors)
	http.HandleFunc("/editNeighbors", c.editNeighbors)
	http.HandleFunc("/editFromTransport", c.editFromTransport)

	http.HandleFunc("/collections", c.collectionHandler)
	http.HandleFunc("/spaces", c.spacesHandler)
	http.HandleFunc("/spaces/new", c.newSpaceHandler)
	http.HandleFunc("/areas", c.areasHandler)

	http.HandleFunc("/deploy", c.deploy)

	err := http.ListenAndServe(":4444", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
