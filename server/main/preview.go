package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

// A loopback-only, ephemeral world using the real compiler, websocket gameplay,
// templates and renderer. It never opens a database or publishes a release.
func runLocalPreview() error {
	collection, err := loadSeedCollection("../../tools/main/data/collections", "bloop")
	if err != nil {
		return err
	}
	data, err := os.ReadFile("../../tools/main/data/colors/colors.json")
	if err != nil {
		return err
	}
	palette, err := decodeWorldPalette(data)
	if err != nil {
		return err
	}
	files, err := compileCollectionFiles(collection, palette)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(files["areas.json"], &areas); err != nil {
		return err
	}
	data, err = os.ReadFile("../../tools/main/data/collections/bloop/manifest.json")
	if err != nil {
		return err
	}
	config := &Configuration{serverName: "Willow Reach · Preview", worldID: "local-preview"}
	if err := json.Unmarshal(data, &config.manifest); err != nil {
		return err
	}
	if err := json.Unmarshal(files["maps.json"], &config.mapAreas); err != nil {
		return err
	}
	world := createGameWorld(nil, config)
	var visitors atomic.Int64
	mux := http.NewServeMux()
	mux.HandleFunc("/assets/colors.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write(files["colors.css"])
	})
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	mux.HandleFunc("/images/", func(w http.ResponseWriter, r *http.Request) {
		data, ok := files[strings.TrimPrefix(r.URL.Path, "/")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write(data)
	})
	mux.HandleFunc("/screen", world.NewSocketConnection)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/play" {
			http.NotFound(w, r)
			return
		}
		spawn := config.manifest.Entry
		login := world.initiateLogin(PlayerRecord{Username: fmt.Sprintf("Wanderer %d", visitors.Add(1)), Team: "sky-blue", Health: 100, Money: 32, StageName: spawn.Stage, Y: spawn.Y, X: spawn.X})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		if r.URL.Path == "/play" {
			tmpl.ExecuteTemplate(w, "player-page", login)
			return
		}
		fmt.Fprint(w, "<!doctype html><html lang=\"en\"><title>Willow Reach · Bloopworld</title>")
		tmpl.ExecuteTemplate(w, "page-header", nil)
		fmt.Fprint(w, "<body>")
		tmpl.ExecuteTemplate(w, "player-page", login)
		fmt.Fprint(w, `<div id="sound-trigger" class="invisible"></div></body></html>`)
	})
	fmt.Println("Willow Reach preview: http://127.0.0.1:4445 (ephemeral; no database or publication)")
	return (&http.Server{Addr: "127.0.0.1:4445", Handler: mux, ReadHeaderTimeout: 5 * time.Second}).ListenAndServe()
}
