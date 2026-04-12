// Command svelgo is the SvelGo CLI.
//
// Usage:
//
//	svelgo new <appname>    scaffold a new SvelGo application in ./<appname>/
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "new" {
		printUsage()
		os.Exit(1)
	}
	appName := os.Args[2]
	if appName == "" {
		printUsage()
		os.Exit(1)
	}

	if err := scaffold(appName); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage: svelgo new <appname>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Creates a new SvelGo application in ./<appname>/.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Assumption: this command is run from inside a directory one level")
	fmt.Fprintln(os.Stderr, "below the framework repo root (e.g. from <repo>/demo/). The generated")
	fmt.Fprintln(os.Stderr, "go.mod replace directive and npm svelgo path assume two levels of")
	fmt.Fprintln(os.Stderr, "nesting above the app directory (../../ to reach the framework root).")
}

type scaffoldData struct {
	AppName string
}

func scaffold(appName string) error {
	appDir := filepath.Join(".", appName)

	// Refuse to overwrite an existing directory.
	if _, err := os.Stat(appDir); err == nil {
		return fmt.Errorf("directory %q already exists — choose a different name or remove it first", appDir)
	}

	data := scaffoldData{AppName: appName}

	files := []struct {
		path    string
		content string
	}{
		{"go.mod", goModTmpl},
		{"embed.go", embedGoTmpl},
		{"main.go", mainGoTmpl},
		{"Makefile", makefileTmpl},
		{filepath.Join("static", ".gitkeep"), ""},
		{filepath.Join("frontend", "package.json"), packageJSONTmpl},
		{filepath.Join("frontend", "svelte.config.js"), svelteConfigTmpl},
		{filepath.Join("frontend", "vite.config.ts"), viteConfigTmpl},
		{filepath.Join("frontend", "index.html"), indexHTMLTmpl},
		{filepath.Join("frontend", "src", "main.ts"), mainTSTmpl},
	}

	// Create all directories and files first.
	for _, f := range files {
		fullPath := filepath.Join(appDir, f.path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("create directory for %s: %w", f.path, err)
		}
		if err := writeTemplate(fullPath, f.content, data); err != nil {
			return fmt.Errorf("write %s: %w", f.path, err)
		}
	}

	fmt.Printf("Created %s/\n", appName)

	// Run go mod tidy inside the new app directory.
	fmt.Println("Running go mod tidy...")
	if err := runCmd(appDir, "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy failed (files are left in place for debugging): %w", err)
	}

	// Run npm install inside the new app's frontend directory.
	frontendDir := filepath.Join(appDir, "frontend")
	fmt.Println("Running npm install...")
	if err := runCmd(frontendDir, "npm", "install"); err != nil {
		return fmt.Errorf("npm install failed (files are left in place for debugging): %w", err)
	}

	fmt.Printf(`
Done! Your new SvelGo app is ready.

  cd %s
  # Edit main.go with your app logic
  make dev

Visit http://localhost:8080 once both processes are running.

Note: the go.mod replace directive points to ../../ (two levels up). This
assumes your app sits one level below the framework repo root
(e.g. <repo>/demo/%s/). If your nesting depth differs, update the
replace directive in %s/go.mod and the svelgo path in
%s/frontend/package.json accordingly.
`, appName, appName, appName, appName)

	return nil
}

// writeTemplate executes a Go text/template into a file, creating it with
// 0644 permissions. An empty content string writes an empty file.
func writeTemplate(path, tmplText string, data scaffoldData) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if tmplText == "" {
		return nil
	}

	tmpl, err := template.New("").Parse(tmplText)
	if err != nil {
		return err
	}
	return tmpl.Execute(f, data)
}

// runCmd runs a command in dir, streaming its combined output to stdout/stderr.
func runCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ---------------------------------------------------------------------------
// File templates
// ---------------------------------------------------------------------------

const goModTmpl = `module {{.AppName}}

go 1.26

// Remove this line when the framework is published to a module proxy.
// This path assumes the app sits one level below the framework repo root
// (e.g. <framework-repo>/demo/{{.AppName}}/). Adjust if your layout differs.
replace github.com/hawkhero/svelgo => ../../

require github.com/hawkhero/svelgo v0.0.0
`

const embedGoTmpl = `package main

import (
	"embed"
	"io/fs"
	"log"

	svelgo "github.com/hawkhero/svelgo"
)

//go:embed all:static
var embeddedStatic embed.FS

func init() {
	sub, err := fs.Sub(embeddedStatic, "static")
	if err != nil {
		log.Fatal("embed: could not sub static/:", err)
	}
	svelgo.SetStaticFS(sub)
}
`

const mainGoTmpl = `package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	svelgo "github.com/hawkhero/svelgo"
	"github.com/hawkhero/svelgo/component"
)

func main() {
	svelgo.Setup() // must be called before http.HandleFunc

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		clickCount := 0

		btn := &component.Button{ID: "btn-1", Label: "Click me (0 clicks)"}
		lbl := &component.Label{ID: "lbl-1", Text: "Count: 0"}

		btn.OnClick = func(ctx context.Context) error {
			clickCount++
			btn.Label = fmt.Sprintf("Click me (%d clicks)", clickCount)
			lbl.Text = fmt.Sprintf("Count: %d", clickCount)
			return nil
		}

		page := svelgo.NewPage()
		page.Add(btn).Add(lbl)
		page.Render(w, r)
	})

	log.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
`

const makefileTmpl = `.PHONY: dev build clean

dev:
	cd frontend && npm run dev &
	SVELGO_DEV=1 go run .

build:
	cd frontend && npm run build
	go build -o dist/{{.AppName}} .

clean:
	rm -rf dist/ static/assets static/.vite
`

// packageJSONTmpl uses the ../../../frontend path because the app lives two
// levels below the framework root (e.g. <repo>/demo/<appname>/frontend/).
const packageJSONTmpl = `{
  "name": "{{.AppName}}-frontend",
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "devDependencies": {
    "@sveltejs/vite-plugin-svelte": "^6.0.0",
    "svelte": "^5.0.0",
    "typescript": "^5.0.0",
    "vite": "^6.3.0",
    "svelgo": "file:../../../frontend"
  },
  "dependencies": {
    "protobufjs": "^7.4.0"
  }
}
`

const svelteConfigTmpl = `export default {}
`

const viteConfigTmpl = `import { svelte } from '@sveltejs/vite-plugin-svelte'
import { defineConfig } from 'vite'
import { resolve } from 'path'

export default defineConfig({
  plugins: [svelte()],
  build: {
    manifest: true,
    rollupOptions: {
      input: resolve(__dirname, 'src/main.ts'),
    },
    outDir: resolve(__dirname, '../static'),
    emptyOutDir: true,
  },
  server: {
    cors: true,
  },
})
`

const indexHTMLTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>{{.AppName}} (dev)</title>
</head>
<body>
  <div id="svelgo-root"></div>
  <script>
    window.__SVELGO_PAGE_ID__  = "dev-page";
    window.__SVELGO_STATE__    = "";
    window.__SVELGO_MANIFEST__ = [];
  </script>
  <script type="module" src="/src/main.ts"></script>
</body>
</html>
`

const mainTSTmpl = `import { bootstrap } from 'svelgo/runtime/client'

bootstrap()
`
