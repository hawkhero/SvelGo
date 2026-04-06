package svelgo

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
)

var devMode = os.Getenv("SVELGO_DEV") == "1"

// staticFS holds the embedded frontend assets. Applications must call
// SetStaticFS before calling Setup (unless running in dev mode).
var staticFS fs.FS

// SetStaticFS registers the embedded static filesystem. Call this from an
// init() function in your application's embed.go, before Setup().
func SetStaticFS(f fs.FS) {
	staticFS = f
}
var debugMode = os.Getenv("SVELGO_DEBUG") == "1"

var (
	resolvedScript string
	resolvedCSS    string
)

type viteManifestEntry struct {
	File    string   `json:"file"`
	CSS     []string `json:"css"`
	IsEntry bool     `json:"isEntry"`
}

// Setup initialises the asset resolver and registers HTTP handlers for
// /ws and /assets/. Call this before registering your own routes.
func Setup() {
	if !devMode && staticFS == nil {
		log.Fatal("SvelGo: call svelgo.SetStaticFS() before Setup() — see https://github.com/hawkhero/svelgo#embedding")
	}
	if debugMode {
		log.Println("SvelGo: debug mode enabled")
	}
	if devMode {
		resolvedScript = "http://localhost:5173/src/main.ts"
		resolvedCSS = ""
		log.Println("SvelGo: dev mode — Vite dev server expected at :5173")
	} else {
		// Parse the Vite manifest to find the hashed bundle filenames
		manifestData, err := fs.ReadFile(staticFS, "static/.vite/manifest.json")
		if err != nil {
			log.Fatal("SvelGo: could not read static/.vite/manifest.json — run `cd frontend && npm run build` first")
		}
		var manifest map[string]viteManifestEntry
		if err := json.Unmarshal(manifestData, &manifest); err != nil {
			log.Fatal("SvelGo: could not parse manifest.json:", err)
		}
		for _, entry := range manifest {
			if entry.IsEntry {
				resolvedScript = "/" + entry.File
				if len(entry.CSS) > 0 {
					resolvedCSS = "/" + entry.CSS[0]
				}
				break
			}
		}
		if resolvedScript == "" {
			log.Fatal("SvelGo: no entry point found in Vite manifest")
		}
		log.Printf("SvelGo: serving embedded assets (script: %s)", resolvedScript)

		// Serve embedded static assets
		staticRoot, _ := fs.Sub(staticFS, "static")
		http.Handle("/assets/", http.FileServer(http.FS(staticRoot)))
	}

	// WebSocket handler
	http.HandleFunc("/ws", WSHandler)
}

func resolveAssets() (script, css string) {
	return resolvedScript, resolvedCSS
}
