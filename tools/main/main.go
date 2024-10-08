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
	http.HandleFunc("/space", c.spaceHandler)
	http.HandleFunc("/space/map", c.spaceMapHandler)
	http.HandleFunc("/space/details", c.spaceDetailsHandler)
	http.HandleFunc("/space/structures", c.spaceStructuresHandler)
	http.HandleFunc("/space/structure", c.spaceStructureHandler) // needs DELETE
	http.HandleFunc("/structure", c.structureHandler)            // needs PUT
	http.HandleFunc("/areas", c.areasHandler)
	http.HandleFunc("/areas/new", c.newAreaHandler)
	http.HandleFunc("/area", c.areaHandler)
	http.HandleFunc("/area/details", c.areaDetailsHandler)
	http.HandleFunc("/area/display", c.areaDisplayHandler)
	http.HandleFunc("/area/neighbors", c.areaNeighborsHandler)
	http.HandleFunc("/area/blueprint", c.areaBlueprintHandler)
	http.HandleFunc("/area/blueprint/instruction", c.blueprintInstructionHandler)
	http.HandleFunc("/area/blueprint/instruction/highlight", c.blueprintInstructionHighlightHandler)
	http.HandleFunc("/area/blueprint/instructions/order", c.instructionOrderHandler)
	http.HandleFunc("/fragments", c.fragmentsHandler)
	http.HandleFunc("/fragments/new", c.fragmentsNewHandler)
	http.HandleFunc("/fragment", c.fragmentHandler)
	http.HandleFunc("/fragment/new", c.fragmentNewHandler)
	http.HandleFunc("/prototypes", c.prototypesHandler)
	http.HandleFunc("/prototypes/new", c.prototypesNewHandler)
	http.HandleFunc("/prototype", c.prototypeHandler)
	http.HandleFunc("/prototype/new", c.prototypeNewHandler)
	http.HandleFunc("/prototype/example", examplePrototype)
	http.HandleFunc("/grid/edit", c.gridEditHandler)
	http.HandleFunc("/grid/click/area", c.gridClickAreaHandler)
	http.HandleFunc("/grid/click/fragment", c.gridClickFragmentHandler)
	http.HandleFunc("/images/", c.imageHandler)
	http.HandleFunc("/interactables", c.interactablesHandler)
	http.HandleFunc("/interactables/new", c.interactablesNewHandler)
	http.HandleFunc("/interactable", c.interactableHandler)
	http.HandleFunc("/interactable/new", c.interactableNewHandler)
	http.HandleFunc("/interactable/example", exampleInteractable)

	http.HandleFunc("/materialPage", c.getMaterialPage)
	//http.HandleFunc("/exampleMaterial", exampleMaterial) // Probably unused
	http.HandleFunc("/getEditColor", c.getEditColor)
	http.HandleFunc("/editColor", c.editColor)
	http.HandleFunc("/getNewColor", getNewColor)
	http.HandleFunc("/newColor", c.newColor)
	http.HandleFunc("/exampleSquare", exampleSquare)
	http.HandleFunc("/outputIngredients", c.outputIngredients)

	http.HandleFunc("/selectFixture", c.selectFixture)

	http.HandleFunc("/editTransports", c.getEditTransports)
	http.HandleFunc("/editTransport", c.editTransport)
	http.HandleFunc("/newTransport", c.newTransport)
	http.HandleFunc("/dupeTransport", c.dupeTransport)
	http.HandleFunc("/deleteTransport", c.deleteTransport)

	http.HandleFunc("/deploy", c.deploy)
	http.HandleFunc("/compile", c.compile)
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
