# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository layout

```
svelgo/                          ← Go module: github.com/hawkhero/svelgo (the framework)
├── *.go                         ← Framework package (package svelgo)
├── proto/ui.proto               ← Framework wire types only (4 core messages)
├── gen/ui/ui.pb.go              ← Auto-generated; do not edit
├── frontend/                    ← npm package "svelgo" (TypeScript runtime)
│   └── src/runtime/             ← client.ts, ws.ts, state.ts, proto.ts, registry.ts
└── example/                     ← Self-contained example app (separate go.mod)
    ├── go.mod                   ← replace github.com/hawkhero/svelgo => ../
    ├── main.go, embed.go        ← App entry point + embed directive
    ├── proto/app.proto          ← App-specific messages (ButtonState)
    ├── gen/app/app.pb.go        ← Auto-generated; do not edit
    └── frontend/                ← App frontend (imports "svelgo" npm package)
        └── src/
            ├── main.ts          ← Imports ./proto, ./registry, then bootstrap()
            ├── proto.ts         ← Registers component decoders
            ├── registry.ts      ← Registers Svelte components
            └── components/      ← App Svelte components
```

## Commands

```bash
# Development (visit http://localhost:8080)
make dev
# Equivalent to:
cd example/frontend && npm run dev &
cd example && SVELGO_DEV=1 go run .

# Production build
make build
# Equivalent to:
cd example/frontend && npm run build   # outputs to example/static/
cd example && go build -o dist/buttonapp .

# Regenerate FRAMEWORK protobuf artifacts after editing proto/ui.proto
make proto

# Regenerate EXAMPLE app protobuf artifacts after editing example/proto/app.proto
make -C example proto

# Clean build artifacts
make clean
```

There are no automated tests in this codebase yet.

## Architecture

SvelGo is a **Go-first UI framework** — a single Go binary that serves HTML pages with embedded Svelte components. There is no separate Node process in production.

### Key boundary: framework vs. application

- **Framework** (`*.go` at root, `frontend/src/runtime/`): never modified for a new application
- **Application** (`example/`): all app-specific code lives here — Go structs, Svelte components, proto messages, embed directive

### Request lifecycle

1. **HTTP GET** → Go handler calls `svelgo.NewPage()`, adds components, calls `page.Render(w, r)`
2. `Render()` serializes all component states as protobuf → base64, registers a `PageSession`, then executes an HTML shell template that injects:
   - `window.__SVELGO_PAGE_ID__` — unique UUID for this page load
   - `window.__SVELGO_STATE__` — base64 protobuf blob of all component states
   - `window.__SVELGO_MANIFEST__` — JSON array of `{id, type, slot}` entries
3. The browser loads the Svelte bundle, `bootstrap()` decodes the state blob, creates per-component Svelte `writable` stores, mounts each Svelte component, then opens a WebSocket to `/ws?page-id=...`

### Event cycle (WebSocket)

Browser → `sendEvent(componentId, eventType)` → binary protobuf `ClientEvent` frame → Go `WSHandler` → looks up component in `PageSession` → calls `HandleEvent()` → Go struct mutates its state → serializes updated state as `StateUpdate` protobuf → sends binary frame back → browser decodes and patches the writable store → Svelte re-renders

### Key framework files

| File | Role |
|---|---|
| `component.go` | `Component` and `EventHandler` interfaces |
| `page.go` | `Page` builder — `NewPage()`, `Add()`, `Render()` |
| `session.go` | `PageSession` — holds live component map + WebSocket conn per page load |
| `ws.go` | `WSHandler` — WebSocket endpoint, event dispatch |
| `assets.go` | `Setup()`, `SetStaticFS()` — asset resolution, HTTP handler registration |
| `proto/ui.proto` | Framework wire types: `PageState`, `ComponentState`, `ClientEvent`, `StateUpdate` |
| `gen/ui/ui.pb.go` | Auto-generated (import as `github.com/hawkhero/svelgo/gen/ui`) |
| `frontend/src/runtime/client.ts` | `bootstrap()` — entry point for browser runtime |
| `frontend/src/runtime/proto.ts` | Protobuf encode/decode + `registerComponentDecoder()` |
| `frontend/src/runtime/state.ts` | Per-component Svelte `writable` stores |
| `frontend/src/runtime/ws.ts` | WebSocket client |
| `frontend/src/runtime/registry.ts` | `registerComponent()` + `ComponentRegistry` |

## Adding a New Component (6 steps)

Work inside `example/` (or your own app):

1. **Proto** — add `message FooState { ... }` to `example/proto/app.proto`, run `make -C example proto`
2. **Go struct** — implement `Component` (+ optionally `EventHandler`) in your app
3. **Register in route** — `page.Add(&Foo{...})` in the HTTP handler
4. **Svelte component** — create `example/frontend/src/components/Foo.svelte`; use `$effect` + `store.subscribe(getComponentStore(id))` from `svelgo/runtime/state`, call `sendEvent(id, 'eventType')` from `svelgo/runtime/ws`
5. **Registry** — call `registerComponent('Foo', Foo)` in `example/frontend/src/registry.ts`
6. **Proto decoder** — call `registerComponentDecoder('Foo', appRoot.lookupType('app.FooState'))` in `example/frontend/src/proto.ts`

## Embedding (app requirement)

Apps must embed their compiled frontend and call `SetStaticFS` before `Setup()`:

```go
//go:embed all:static
var embeddedStatic embed.FS

func init() {
    sub, _ := fs.Sub(embeddedStatic, "static")
    svelgo.SetStaticFS(sub)
}
```

## Gotchas

- `//go:embed all:static` — the `all:` prefix is required; without it Go skips the hidden `.vite/` directory
- `example/static/.gitkeep` — keep this file; it lets `go:embed` compile before the first `npm run build`
- `protoc` emits flat; Makefiles always `mv gen/*.pb.go gen/<pkg>/`
- `SVELGO_DEV=1` switches asset serving to the Vite dev server at `:5173`; `SetStaticFS` is not required in dev mode
- In Svelte 5 components, use `$effect` + `store.subscribe(...)` — not `$store` shorthand
- `ws.binaryType = 'arraybuffer'` must be set before messages arrive (done in `ws.ts`)
- `svelgo.Setup()` must be called before registering application routes
- The `example/go.mod` uses `replace github.com/hawkhero/svelgo => ../` for local development; remove it when the framework is published
