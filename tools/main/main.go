package main

import (
	"fmt"
	"html/template"
	"net/http"
)

var tmpl = template.Must(template.ParseGlob("templates/*.tmpl.html"))

func main() {
	fmt.Println("Attempting to start server...")
	c := populateFromJson()

	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./assets"))))

	http.HandleFunc("/collections", c.collectionsHandler)
	http.HandleFunc("/spaces", c.spacesHandler)
	http.HandleFunc("/spaces/new", c.newSpaceHandler)
	http.HandleFunc("/areas", c.areasHandler)
	http.HandleFunc("/area", c.areaHandler)
	http.HandleFunc("/area/details", c.areaDetailsHandler)
	http.HandleFunc("/area/display", c.areaDisplayHandler)
	http.HandleFunc("/area/neighbors", c.areaNeighborsHandler)
	http.HandleFunc("/fragments", c.fragmentsHandler)
	//http.HandleFunc("/fragments/new", c.fragmentsHandler)
	http.HandleFunc("/fragment", c.fragmentHandler)
	http.HandleFunc("/grid/edit", c.gridEditHandler)
	http.HandleFunc("/grid/click/area", c.gridClickAreaHandler)
	http.HandleFunc("/grid/click/fragment", c.gridClickFragmentHandler)

	//http.HandleFunc("/fragment/select", c.fragmentSelectHandler)

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

	//http.HandleFunc("/clickOnSquare", c.clickOnSquare)
	http.HandleFunc("/selectMaterial", c.selectMaterial)
	http.HandleFunc("/selectFixture", c.selectFixture)

	http.HandleFunc("/editTransports", c.getEditTransports)
	http.HandleFunc("/editTransport", c.editTransport)
	http.HandleFunc("/newTransport", c.newTransport)
	http.HandleFunc("/dupeTransport", c.dupeTransport)
	http.HandleFunc("/deleteTransport", c.deleteTransport)

	http.HandleFunc("/deploy", c.deploy)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		tmpl.ExecuteTemplate(w, "home", c.Collections)
	})

	err := http.ListenAndServe(":4444", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
