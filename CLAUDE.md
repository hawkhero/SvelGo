# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Development (visit http://localhost:8080)
make dev
# Equivalent to:
cd frontend && npm run dev &
SVELGO_DEV=1 go run ./cmd/app/

# Production build
make build
# Equivalent to:
cd frontend && npm run build   # outputs to ../static/
go build -o dist/svelgo-app ./cmd/app/

# Regenerate protobuf artifacts after editing proto/ui.proto
make proto

# Clean build artifacts
make clean
```

There are no automated tests in this codebase yet.

## Architecture

SvelGo is a **Go-first UI framework** — a single Go binary that serves HTML pages with embedded Svelte components. There is no separate Node process in production.

### Request lifecycle

1. **HTTP GET** → Go handler calls `svelgo.NewPage()`, adds components, calls `page.Render(w, r)`
2. `Render()` serializes all component states as protobuf → base64, registers a `PageSession`, then executes an HTML shell template that injects:
   - `window.__SVELGO_PAGE_ID__` — unique UUID for this page load
   - `window.__SVELGO_STATE__` — base64 protobuf blob of all component states
   - `window.__SVELGO_MANIFEST__` — JSON array of `{id, type, slot}` entries
3. The browser loads the Svelte bundle, `bootstrap()` decodes the state blob, creates per-component Svelte `writable` stores, mounts each Svelte component, then opens a WebSocket to `/ws?page-id=...`

### Event cycle (WebSocket)

Browser → `sendEvent(componentId, eventType)` → binary protobuf `ClientEvent` frame → Go `WSHandler` → looks up component in `PageSession` → calls `HandleEvent()` → Go struct mutates its state → serializes updated state as `StateUpdate` protobuf → sends binary frame back → browser decodes and patches the writable store → Svelte re-renders

### Key files

| File | Role |
|---|---|
| `component.go` | `Component` and `EventHandler` interfaces |
| `page.go` | `Page` builder — `NewPage()`, `Add()`, `Render()` |
| `session.go` | `PageSession` — holds live component map + WebSocket conn per page load |
| `ws.go` | `WSHandler` — WebSocket endpoint, event dispatch |
| `assets.go` | `Setup()` — resolves asset paths (dev: Vite, prod: embedded), registers `/ws` and `/assets/` |
| `embed.go` | `//go:embed all:static` — bundles compiled Svelte into binary |
| `proto/ui.proto` | Shared wire types: `PageState`, `ComponentState`, `ClientEvent`, `StateUpdate` |
| `gen/ui/ui.pb.go` | Auto-generated Go protobuf types (import as `svelgo/gen/ui`) |
| `cmd/app/main.go` | Example app — shows how to define components and register routes |
| `frontend/src/runtime/client.ts` | `bootstrap()` — entry point for browser runtime |
| `frontend/src/runtime/proto.ts` | All protobuf encode/decode; add new component type decoders here |
| `frontend/src/runtime/state.ts` | Per-component Svelte `writable` stores |
| `frontend/src/runtime/ws.ts` | WebSocket client — sends events, receives state updates |
| `frontend/src/runtime/registry.ts` | Maps type strings (e.g. `"Button"`) to Svelte component constructors |

## Adding a New Component (6 steps)

1. **Proto** — add a `message FooState { ... }` to `proto/ui.proto`, then run `make proto`
2. **Go struct** — implement `Component` (+ optionally `EventHandler`) interface in Go
3. **Register in app** — `page.Add(&Foo{...})` in the route handler
4. **Svelte component** — create `frontend/src/components/Foo.svelte`; use `$effect` + `store.subscribe(getComponentStore(id))` to read state, call `sendEvent(id, 'eventType')` to fire events
5. **Registry** — add `Foo` to `ComponentRegistry` in `frontend/src/runtime/registry.ts`
6. **Proto decoder** — add `Foo: root.lookupType('ui.FooState')` to `componentTypes` in `frontend/src/runtime/proto.ts`

## Gotchas

- `//go:embed all:static` — the `all:` prefix is required; without it Go skips the hidden `.vite/` directory which contains the Vite manifest
- `protoc` emits `gen/ui.pb.go`; `make proto` moves it to `gen/ui/ui.pb.go` so the import path `svelgo/gen/ui` resolves correctly
- `SVELGO_DEV=1` switches asset serving from embedded binary to the Vite dev server at `:5173`
- In Svelte 5 components, use `$effect` + `store.subscribe(...)` — not the `$store` shorthand — when the store comes from a function call like `getComponentStore(id)`
- `ws.binaryType = 'arraybuffer'` must be set before any messages arrive (done in `ws.ts`)
- `svelgo.Setup()` must be called before registering application routes (it registers `/ws` and `/assets/`)