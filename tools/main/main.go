package main

import (
	"fmt"
	"html/template"
	"net/http"
)

var tmpl = template.New("legacy")

func main() {
	fmt.Println("Initializing...")
	c := populateFromJson() // shouldn't this be a pointer?
	ExecuteCLICommands(&c)

	fmt.Println("Attempting to start server...")
	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./assets"))))
	http.HandleFunc("/images/", c.imageHandler)
	http.HandleFunc("/api/bootstrap", c.apiBootstrapHandler)
	http.HandleFunc("/api/collection", c.apiCollectionHandler)
	http.HandleFunc("/api/space", c.apiSaveSpaceHandler)
	http.HandleFunc("/api/space/flatten", c.apiFlattenSpaceHandler)
	http.HandleFunc("/api/space/create", c.apiCreateSpaceHandler)
	http.HandleFunc("/api/area/create", c.apiCreateAreaHandler)
	http.HandleFunc("/api/fragment-set", c.apiSaveFragmentSetHandler)
	http.HandleFunc("/api/prototype-set", c.apiSavePrototypeSetHandler)
	http.HandleFunc("/api/interactable-set", c.apiSaveInteractableSetHandler)
	http.HandleFunc("/api/colors", c.apiColorsHandler)
	http.HandleFunc("/api/compile", c.apiCompileHandler)
	http.HandleFunc("/api/deploy", c.apiDeployHandler)
	http.HandleFunc("/", c.spaHandler)

	err := http.ListenAndServe(":4444", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
