# Framework Code vs Application Code

This document maps every file in the repository to its owner — the **framework** (code
that ships as the reusable `svelgo` package) or the **application** (code that belongs
to the specific app being built on top of it).

---

## Quick summary

| Owner | Go | TypeScript / Svelte | Proto |
|---|---|---|---|
| **Framework** | `*.go` at root, `gen/ui/ui.pb.go`, `embed.go` | `src/runtime/*.ts`, `src/main.ts` | `PageState`, `ComponentState`, `ClientEvent`, `StateUpdate` |
| **Application** | `cmd/app/main.go` | `src/components/*.svelte`, entries in `registry.ts` and `proto.ts` | Component-specific messages (e.g. `ButtonState`) |

---

## Go — server side

### Framework (`package svelgo` — root `*.go` files)

| File | What it does |
|---|---|
| `component.go` | Defines the `Component` and `EventHandler` interfaces every component must satisfy |
| `page.go` | `Page` builder — `NewPage()`, `Add()`, `Render()` — serialises state and writes the HTML shell |
| `session.go` | `PageSession` — maps component IDs to live Go structs; holds the WebSocket connection per page load |
| `ws.go` | `WSHandler` — the WebSocket endpoint; decodes `ClientEvent`, dispatches to the right component, encodes and sends `StateUpdate` |
| `assets.go` | `Setup()` — reads env vars, resolves asset paths (Vite URL in dev, hashed bundle in prod), registers `/ws` and `/assets/` handlers |
| `template.go` | The one HTML shell template the framework ever writes; injects page globals into `<script>` |
| `embed.go` | Embeds the compiled frontend (`static/`) into the Go binary |
| `gen/ui/ui.pb.go` | Auto-generated from `proto/ui.proto` — do not edit by hand |

You never modify these files for a new application.

### Application (`cmd/app/main.go`)

Everything inside `cmd/app/` is application code:

- **Component structs** — e.g. `Button` — holding application state (`label`, `clickCount`)
- **`ProtoState()`** — returns the component-specific protobuf message that matches the state struct
- **`HandleEvent()`** — contains the business logic that runs when the user interacts with a component
- **Route handlers** — `http.HandleFunc("/", ...)` — compose pages with `svelgo.NewPage()` and `page.Add(...)`
- **`main()`** — calls `svelgo.Setup()` then starts the HTTP server

---

## Proto — wire contract (`proto/ui.proto`)

The proto file is split into two sections:

### Framework messages (top of file)
These are owned by the framework and must not be changed:

```
PageState       — envelope for all component states on a page load
ComponentState  — id + type + opaque state_bytes for one component
ClientEvent     — browser → server event wire type
StateUpdate     — server → browser state patch wire type
```

### Application messages (bottom of file, after the divider comment)
One message per component type, owned by the application:

```
ButtonState     — label + click_count for the Button component
```

When you add a new component you add a new message here and run `make proto`.

---

## TypeScript / Svelte — client side

### Framework (`frontend/src/runtime/`)

| File | What it does |
|---|---|
| `client.ts` | `bootstrap()` — decodes the state blob, mounts components, opens the WebSocket |
| `ws.ts` | WebSocket client — sends `ClientEvent` frames, receives `StateUpdate` frames, updates stores |
| `state.ts` | Per-component Svelte `writable` stores — the reactive state layer |
| `proto.ts` | All protobuf encode/decode — `decodePageState`, `decodeStateUpdate`, `encodeClientEvent`, `decodeComponentState` |
| `main.ts` | Entry point — just calls `bootstrap()` |

You never modify these files for a new component.

### Application (`frontend/src/`)

| File | What it does |
|---|---|
| `components/Button.svelte` | Presentational component — reads its state store, fires events; pure UI, no business logic |
| `runtime/registry.ts` | Maps type strings (`"Button"`) to Svelte constructors — **add one line per new component** |
| `runtime/proto.ts` (partial) | The `componentTypes` map at the top — **add one entry per new component** to register the protobuf decoder |

`registry.ts` and `proto.ts` straddle the boundary: their infrastructure is framework
code, but the entries inside the maps are application code.

---

## The boundary rule

A useful test: if you deleted `cmd/app/` and `src/components/` and started a completely
different application, what would you keep unchanged?

- **Keep** — everything in `package svelgo` (root `*.go`), `src/runtime/`, `src/main.ts`, and the top four proto messages.
- **Replace** — `cmd/app/main.go`, `src/components/*.svelte`, the entries in `registry.ts` and `proto.ts`, and the component-specific proto messages.

That line is the framework / application boundary.
