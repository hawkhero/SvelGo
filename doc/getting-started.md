# Getting Started with SvelGo

SvelGo is a Go-first UI framework. A single Go binary serves HTML pages containing reactive Svelte components. There is no separate Node process in production — the frontend is compiled once, embedded in the binary, and served by the Go HTTP server.

This guide assumes you know Go well and have a passing familiarity with TypeScript. You do not need to know Svelte internals to use built-in components.

---

## Prerequisites

| Tool | Minimum version | Notes |
|---|---|---|
| Go | 1.26+ | `go version` |
| Node.js | 18+ | `node -v` |
| npm | 9+ | |
| protoc + protoc-gen-go | Any | Only needed if you add custom proto messages |

---

## Repository layout

The repository contains two separate Go modules:

```
svelgo/                          ← module github.com/hawkhero/svelgo (the framework)
├── *.go                         ← package svelgo — the entire Go API surface
├── component/                   ← built-in components (Button, Label)
├── proto/ui.proto               ← framework wire types (4 core messages + built-in states)
├── gen/ui/ui.pb.go              ← auto-generated; never edit by hand
└── frontend/src/runtime/       ← TypeScript runtime (client.ts, ws.ts, state.ts, …)

example/                         ← module example/buttonapp (a self-contained app)
├── go.mod                       ← has: replace github.com/hawkhero/svelgo => ../
├── main.go                      ← HTTP server + route handlers
├── embed.go                     ← //go:embed all:static + svelgo.SetStaticFS()
└── frontend/                    ← Vite + Svelte app (imports the svelgo npm package)
    └── src/
        ├── main.ts              ← calls bootstrap() — this is the entire entry point
        ├── proto.ts             ← only needed if you add custom components
        └── registry.ts          ← only needed if you add custom components
```

**The framework is never modified for a new application.** Everything app-specific lives in the app module (the `example/` directory here).

### The replace directive

When developing against a local framework checkout, add this to your app's `go.mod`:

```
replace github.com/hawkhero/svelgo => /path/to/svelgo
```

Remove this line (and change the `require` to a real version tag) when you publish the framework or vendor it.

---

## Built-in components

The framework ships two components that require no proto definition, no Svelte file, and no registration step:

| Component | Go type | Fields | Events |
|---|---|---|---|
| Button | `component.Button` | `ID`, `Label`, `Disabled`, `OnClick func()` | `click` |
| Label | `component.Label` | `ID`, `Text` | none (display only) |

```go
import "github.com/hawkhero/svelgo/component"

btn := &component.Button{ID: "btn-1", Label: "Click me"}
btn.OnClick = func() {
    btn.Label = "Clicked!"
}

lbl := &component.Label{ID: "lbl-1", Text: "Hello World"}
```

If your app only uses built-in components, `proto/`, `gen/`, `frontend/src/proto.ts`, and `frontend/src/registry.ts` are all unnecessary.

See `doc/built-in-components.md` for full reference and wiring examples.

---

## Creating a new application

Scaffold from scratch — do not copy `example/`. The following steps build the minimal app directory structure:

```
myapp/
├── Makefile
├── go.mod
├── go.sum
├── main.go          ← HTTP handler + component wiring
├── embed.go         ← //go:embed all:static
├── static/
│   └── .gitkeep     ← lets go:embed compile before the first npm run build
└── frontend/
    ├── package.json
    ├── svelte.config.ts     ← minimal Svelte config (prevents vite-plugin-svelte warning)
    ├── vite.config.ts
    ├── index.html
    └── src/
        ├── main.ts          ← calls bootstrap() — the entire entry point for built-in-only apps
        ├── proto.ts         ← only needed for custom components; do not create for built-in-only apps
        ├── registry.ts      ← only needed for custom components; do not create for built-in-only apps
        └── components/      ← only needed for custom components
```

> **Built-in-only apps:** if you only use `component.Button` and `component.Label`, create only `main.ts`. Do **not** create `proto.ts`, `registry.ts`, or `components/` — they are not needed and their absence is correct. If those files are absent and `main.ts` imports them, the build will fail.

### Step 1 — Initialize the Go module

```bash
mkdir myapp && cd myapp
go mod init myapp
go get github.com/hawkhero/svelgo
```

**go.mod** when developing against a local framework checkout:

```
module myapp

go 1.26

require (
    github.com/hawkhero/svelgo v0.0.0
)

// Remove this line when the framework is published
replace github.com/hawkhero/svelgo => /path/to/svelgo
```

### Step 2 — Write the Go files

**embed.go** — embeds the compiled frontend into the binary:

```go
package main

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
```

> The `all:` prefix in `//go:embed all:static` is **required**. Without it, Go silently omits the hidden `.vite/` directory that the framework reads at startup.

**main.go** — minimal working example with both built-in components:

```go
package main

import (
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

        btn := &component.Button{ID: "btn-1", Label: "Click me (0)"}
        lbl := &component.Label{ID: "lbl-1", Text: "Count: 0"}

        btn.OnClick = func() {
            clickCount++
            btn.Label = fmt.Sprintf("Click me (%d)", clickCount)
            lbl.Text = fmt.Sprintf("Count: %d", clickCount)
        }

        page := svelgo.NewPage()
        page.Add(btn).Add(lbl)
        page.Render(w, r)
    })

    log.Println("Listening on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Step 3 — Set up the frontend

**frontend/package.json:**

```json
{
  "name": "myapp-frontend",
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
    "svelgo": "file:../../frontend"
  },
  "dependencies": {
    "protobufjs": "^7.4.0"
  }
}
```

> **Path convention for the `svelgo` local dependency:** the path is relative from your app's `frontend/` directory to the framework's `frontend/` directory. The standard layout places your app one level under the framework root (e.g. `svelgo/myapp/`), making the correct path `../../frontend` — two levels up to reach the framework root, then into `frontend/`. If your app is nested deeper (e.g. `svelgo/demo/clickcounter/`), add one extra `../` per additional level: `../../../frontend`. `example/frontend/package.json` uses `../../frontend` as the reference implementation. When the framework is published to npm, replace the `file:` path with the published package name.

```bash
cd frontend && npm install
```

**frontend/svelte.config.ts:**

```ts
export default {}
```

This file suppresses the `[vite-plugin-svelte] no Svelte config found` warning emitted on every dev-server startup. The empty export is correct — the plugin's defaults are suitable for all SvelGo apps.

**frontend/vite.config.ts:**

```ts
import { svelte } from '@sveltejs/vite-plugin-svelte'
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
```

**frontend/index.html** (used only by the Vite dev server):

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>MyApp (dev)</title>
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
```

### Step 4 — Write the frontend entry point

**frontend/src/main.ts** — when using only built-in components, this is the entire entry point:

```ts
import { bootstrap } from 'svelgo/runtime/client'

bootstrap()
```

`bootstrap()` decodes the initial page state injected by the server, mounts each Svelte component, and opens the WebSocket connection. Built-in component constructors and proto decoders are registered automatically by the framework — no `proto.ts` or `registry.ts` needed.

### Step 5 — Write the Makefile

```makefile
.PHONY: dev build clean

dev:
	cd frontend && npm run dev &
	SVELGO_DEV=1 go run .

build:
	cd frontend && npm run build
	go build -o dist/myapp .

clean:
	rm -rf dist/ static/assets static/.vite
```

### Step 6 — Create the static placeholder

```bash
mkdir -p static && touch static/.gitkeep
```

This empty file lets `//go:embed all:static` compile before the first `npm run build`.

---

## Dev mode vs. production mode

> **Important — run this once after a fresh checkout, before `make dev`:**
>
> ```bash
> make proto
> ```
>
> The repository ships with a pre-generated `gen/ui/ui.pb.go`, but it may be stale or built against a different `protoc-gen-go` version. Running `make proto` regenerates it from `proto/ui.proto` using your local toolchain. If you skip this step, the Go server will panic at startup with a `slice bounds out of range` error inside the protobuf initializer. You only need to rerun `make proto` if you edit `proto/ui.proto`.

### Development (`make dev`)

```bash
make dev
```

Starts two processes:

1. `cd frontend && npm run dev` — Vite dev server at `http://localhost:5173` with hot-module replacement
2. `SVELGO_DEV=1 go run .` — Go server at `http://localhost:8080`

Visit `http://localhost:8080`. The Go server renders the HTML shell; the browser fetches the live JS bundle from Vite at `:5173`. Svelte component changes reload instantly without restarting Go.

`SVELGO_DEV=1` tells the framework to point the `<script>` tag at Vite (`http://localhost:5173/src/main.ts`) rather than the embedded bundle. `SetStaticFS` is **not** required in dev mode.

### Production (`make build`)

```bash
make build
./dist/myapp
```

Two sequential steps:

1. `cd frontend && npm run build` — compiles the Svelte app to `static/` (hashed filenames, Vite manifest)
2. `go build -o dist/myapp .` — compiles the Go binary with the static files embedded

The resulting binary is fully self-contained — no Node, no Vite, no separate static directory needed at runtime.

### Environment variables

| Variable | Effect |
|---|---|
| `SVELGO_DEV=1` | Serve JS from Vite at `:5173` instead of embedded assets |
| `SVELGO_DEBUG=1` | Enable verbose browser console logging (page state, WebSocket frames) |

---

## Request and event lifecycle

Understanding the data flow helps when debugging:

1. Browser sends `GET /` → your Go handler builds a `Page`, adds components, calls `Render`.
2. `Render` serializes all component states as protobuf, base64-encodes the blob, registers a `PageSession`, then writes an HTML shell that injects three JavaScript globals: `__SVELGO_PAGE_ID__`, `__SVELGO_STATE__`, and `__SVELGO_MANIFEST__`.
3. The browser loads the Svelte bundle. `bootstrap()` decodes the state blob, creates per-component Svelte stores, mounts each component, and opens a WebSocket to `/ws?page-id=...`.
4. When a user interacts with a component (e.g., clicks a button), the Svelte component calls `sendEvent(id, 'click')`, which sends a binary protobuf frame over the WebSocket.
5. The Go server receives the frame, looks up the component in the `PageSession`, calls `HandleEvent()` on it, and the component mutates its state.
6. The framework serializes **all** component states as a single `StateUpdate` protobuf and sends it back. Every component in the session is included — cross-component mutations propagate in the same frame. If any component fails to serialize, **no update is sent at all** (all-or-nothing) — the client state stays consistent even under partial failures.
7. The browser decodes the update and patches each Svelte store, triggering re-renders.
8. When the browser tab closes or navigates away, the WebSocket disconnects and the Go server frees the `PageSession` immediately. Sessions do not leak across page loads.

---

## Page builder API

```go
// Create a new page with a unique session ID.
page := svelgo.NewPage()

// Add one or more components. Returns *Page for chaining.
page.Add(componentA).Add(componentB)

// Serialize state, register session, write HTML response.
page.Render(w, r)
```

`Add` accepts any value that implements the `svelgo.Component` interface (see `component.go`). The interface requires four methods: `ComponentID() string`, `ComponentType() string`, `Slot() string`, and `ProtoState() proto.Message`.

Components that also want to receive user events must implement `svelgo.EventHandler`, which adds `HandleEvent(eventType string, payload []byte) error`.

---

## Adding a custom component (6-step workflow)

When neither `component.Button` nor `component.Label` fits your needs, follow these six steps inside your app (not inside the framework):

### 1. Define a proto message

Create `proto/app.proto`:

```proto
syntax = "proto3";
package app;
option go_package = "myapp/gen/app";

message CounterState {
  string label = 1;
  int32  count = 2;
}
```

Generate Go + JS descriptor files (requires `npm install` first):

```bash
mkdir -p gen/app
protoc \
  --go_out=gen \
  --go_opt=paths=source_relative \
  --proto_path=proto \
  proto/app.proto
mv gen/app.pb.go gen/app/app.pb.go

# Generate JS descriptor for protobufjs
frontend/node_modules/.bin/pbjs \
  -t json proto/app.proto \
  -o frontend/src/app_descriptor.json
```

Add a `proto` target to your Makefile:

```makefile
proto:
	PATH="$$PATH:$$HOME/go/bin" protoc \
		--go_out=gen \
		--go_opt=paths=source_relative \
		--proto_path=proto \
		proto/app.proto
	mv gen/app.pb.go gen/app/app.pb.go
	frontend/node_modules/.bin/pbjs \
		-t json proto/app.proto \
		-o frontend/src/app_descriptor.json
```

### 2. Implement a Go struct

```go
import (
    apipb "myapp/gen/app"
    "google.golang.org/protobuf/proto"
)

type Counter struct {
    id    string
    label string
    count int
}

func (c *Counter) ComponentID()   string { return c.id }
func (c *Counter) ComponentType() string { return "Counter" }
func (c *Counter) Slot()          string { return "root" }

func (c *Counter) ProtoState() proto.Message {
    return &apipb.CounterState{Label: c.label, Count: int32(c.count)}
}

func (c *Counter) HandleEvent(eventType string, _ []byte) error {
    if eventType == "increment" {
        c.count++
        c.label = fmt.Sprintf("Count: %d", c.count)
    }
    return nil
}
```

### 3. Register in the route handler

```go
page.Add(&Counter{id: "counter-1", label: "Count: 0"})
```

### 4. Create a Svelte component

**frontend/src/components/Counter.svelte:**

```svelte
<script lang="ts">
  import type { Writable } from 'svelte/store'
  import { getComponentStore } from 'svelgo/runtime/state'
  import { sendEvent } from 'svelgo/runtime/ws'

  let { id }: { id: string } = $props()

  let state: Record<string, unknown> = $state({})
  $effect(() => {
    const store = getComponentStore(id) as Writable<Record<string, unknown>>
    return store.subscribe(s => { state = s })
  })
</script>

<div>
  <p>{state.label ?? ''}</p>
  <button onclick={() => sendEvent(id, 'increment')}>+1</button>
</div>
```

> **Svelte 5:** use `$effect` + `store.subscribe()`. Do not use the `$store` shorthand — it does not work with dynamically-keyed stores.

### 5. Register the Svelte constructor

**frontend/src/registry.ts:**

```ts
import { registerComponent } from 'svelgo/runtime/registry'
import Counter from './components/Counter.svelte'

registerComponent('Counter', Counter)
```

> **Duplicate registration:** if you call `registerComponent` twice with the same type name but a different constructor, the framework emits a `console.warn` and overwrites the earlier entry. Registering the same constructor twice (e.g. the same import from two files) is silently accepted. Each type name should be registered exactly once.

### 6. Register the proto decoder

**frontend/src/proto.ts:**

```ts
import protobuf from 'protobufjs/light'
import { registerComponentDecoder } from 'svelgo/runtime/proto'
import appDescriptor from './app_descriptor.json'

const appRoot = protobuf.Root.fromJSON(appDescriptor as protobuf.INamespace)

registerComponentDecoder('Counter', appRoot.lookupType('app.CounterState'))
```

**Update frontend/src/main.ts** to import both registration files before `bootstrap()`. These imports are only correct once `proto.ts` and `registry.ts` actually exist (you created them in steps 5 and 6 above). Do **not** add these lines to a built-in-only app — importing a file that does not exist will break the build with "cannot find module './proto'".

```ts
import './proto'
import './registry'
import { bootstrap } from 'svelgo/runtime/client'

bootstrap()
```

---

## Architecture notes

- **State lives in Go.** Svelte is the view layer only — it holds no business logic or mutable state.
- **`svelgo.Setup()`** must be called before all `http.HandleFunc` registrations.
- **`SetStaticFS()`** must be called before `Setup()` (production mode only). If `staticFS` has not been set in production mode, `Setup()` calls `log.Fatal` at startup rather than panicking later.
- **Session lifetime** mirrors the WebSocket connection. The `PageSession` (component map + connection) is created by `page.Render()` and freed immediately when the WebSocket disconnects. There is one session per page load; navigating back or refreshing creates a new session.
- **State updates are all-or-nothing.** If any component's `ProtoState()` fails to marshal, the entire `StateUpdate` is aborted — no partial frame is sent to the client.
- Built-in component types use the `svelgo.` namespace (`"svelgo.Button"`, `"svelgo.Label"`), which cannot conflict with custom component names.
- Proto field names are snake_case in `.proto`; protobufjs converts them to camelCase on the JavaScript side automatically.
