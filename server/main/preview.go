package main

import (
	"encoding/json"
	"fmt"
	"net"
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
	if err := validatePreviewWorld(config.manifest); err != nil {
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
	mux.HandleFunc("/review", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, previewReviewPage)
	})
	mux.HandleFunc("/review/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(populationStatus(world))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/play" {
			http.NotFound(w, r)
			return
		}
		team := r.URL.Query().Get("team")
		if team != "fuchsia" {
			team = "sky-blue"
		}
		spawn := config.manifest.Entry
		for _, t := range config.manifest.Teams {
			if t.ID == team {
				spawn = t.Spawn
			}
		}
		login := world.initiateLogin(PlayerRecord{Username: fmt.Sprintf("Wanderer %d", visitors.Add(1)), Team: team, Health: 150, Money: 32, StageName: spawn.Stage, Y: spawn.Y, X: spawn.X})
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
	port := os.Getenv("OPENWORLD_PREVIEW_PORT")
	if port == "" {
		port = "4445"
	}
	address := net.JoinHostPort("127.0.0.1", port)
	fmt.Printf("Willow Reach preview: http://%s (ephemeral; no database or publication)\n", address)
	return (&http.Server{Addr: address, Handler: mux, ReadHeaderTimeout: 5 * time.Second}).ListenAndServe()
}

func validatePreviewWorld(manifest WorldManifest) error {
	byName := map[string]Area{}
	for _, a := range areas {
		byName[a.Name] = a
	}
	return validateWorldRules(manifest, byName)
}

const previewReviewPage = `<!doctype html><html lang="en"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Willow Frontier review</title><style>body{margin:30px auto;padding:0 20px;max-width:1100px;background:#142822;color:#e6efda;font:16px/1.5 system-ui}a{color:#a8dadc}nav{display:flex;gap:24px;flex-wrap:wrap}pre{background:#203c35;padding:20px;border-radius:12px;overflow:auto;font-size:13px}h1{color:#eedcb2}p{color:#b7ccba}</style><h1>Willow Frontier · Local review</h1><nav><a href="/?team=sky-blue" target="_blank">Play Sky Blue</a><a href="/?team=fuchsia" target="_blank">Play Fuchsia</a><a href="/review/status">Population JSON</a></nav><p>Ephemeral local sessions. Updates every two seconds. Spawn countdowns advance only while a stage is occupied; counts are shared by its players. Inspect NPC/pickup caps, replenishment, and team arrival positions here while playing.</p><pre id="population" role="status">Loading…</pre><script>async function refresh(){try{const r=await fetch('/review/status');document.getElementById('population').textContent=JSON.stringify(await r.json(),null,2)}catch(e){document.getElementById('population').textContent='Preview disconnected.'}}refresh();setInterval(refresh,2000)</script></html>`
