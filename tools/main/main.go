package main

import (
	"fmt"
	"html/template"
	"net/http"
)

var tmpl = template.Must(template.ParseGlob("templates/*.tmpl.html"))

func main() {
	fmt.Println("Initializing...")
	c := populateFromJson() // shouldn't this be a pointer?
	ExecuteCLICommands(&c)

	/*
		// Fuse two spaces
		col, ok := c.Collections["bloop"]
		if !ok {
			panic("no bloop collection")
		}
		blue := col.Spaces["team-blue"]
		fusia := col.Spaces["team-fusia"]
		var areaNameMaker = func(base string) func(y, x int) string {
			return func(y, x int) string {
				return fmt.Sprintf("%s:%d-%d", base, y, x)
			}
		}
		coordToBlue := areaNameMaker("team-blue")
		coordToFusia := areaNameMaker("team-fusia")
		for i := 0; i < 8; i++ {
			// 1st side N/S
			blue.getAreaByName(coordToBlue(0, i)).North = coordToFusia(7, i)
			fusia.getAreaByName(coordToFusia(7, i)).South = coordToBlue(0, i)

			blue.getAreaByName(coordToBlue(7, i)).South = coordToFusia(0, i)
			fusia.getAreaByName(coordToFusia(0, i)).North = coordToBlue(7, i)

			//2nd side E/W
			blue.getAreaByName(coordToBlue(i, 0)).West = coordToFusia(i, 7)
			fusia.getAreaByName(coordToFusia(i, 7)).East = coordToBlue(i, 0)

			blue.getAreaByName(coordToBlue(i, 7)).East = coordToFusia(i, 0)
			fusia.getAreaByName(coordToFusia(i, 0)).West = coordToBlue(i, 7)
		}
	*/

	fmt.Println("Attempting to start server...")
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./assets"))))

	http.HandleFunc("/collections", c.collectionsHandler)
	http.HandleFunc("/spaces", c.spacesHandler)
	http.HandleFunc("/spaces/new", c.newSpaceHandler)
	http.HandleFunc("/space", c.spaceHandler)
	http.HandleFunc("/space/map", c.spaceMapHandler)
	http.HandleFunc("/space/details", c.spaceDetailsHandler)
	http.HandleFunc("/space/structures", c.spaceStructuresHandler)
	http.HandleFunc("/space/structure", c.spaceStructureHandler)
	http.HandleFunc("/structure", c.structureHandler)
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

	http.HandleFunc("/deploy", c.deployHandler)
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
