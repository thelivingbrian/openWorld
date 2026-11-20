package main

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
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
	http.HandleFunc("/api/manifest", c.apiManifestHandler)
	http.HandleFunc("/api/compile", c.apiCompileHandler)
	http.HandleFunc("/api/deploy", c.apiDeployHandler)
	http.HandleFunc("/", c.spaHandler)

	port := os.Getenv("TOOLS_PORT")
	if port == "" {
		port = "4444"
	}
	address := net.JoinHostPort("127.0.0.1", port)
	fmt.Printf("Design workspace: http://%s/design/\n", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
