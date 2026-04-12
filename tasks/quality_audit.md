# SvelGo Framework Quality Audit

**Date:** 2026-04-12  
**Auditor:** svelgo-dev (framework architect)  
**Scope:** All Go source files, proto definitions, TypeScript runtime, built-in components, Makefile  
**Standards:** 8 wire-protocol / runtime quality standards (see below)

---

## Files Reviewed

| File | Role |
|---|---|
| `component.go` | Component + EventHandler interfaces |
| `page.go` | Page builder — NewPage, Add, Render |
| `session.go` | Session store — Register, Get, Delete |
| `ws.go` | WebSocket handler |
| `assets.go` | Setup, SetStaticFS, asset resolution |
| `template.go` | HTML shell template |
| `component/button.go` | Built-in Button component |
| `component/label.go` | Built-in Label component |
| `proto/ui.proto` | Framework wire types |
| `gen/ui/ui.pb.go` | Generated protobuf (auto, reviewed for correctness) |
| `frontend/src/runtime/client.ts` | bootstrap() entry point |
| `frontend/src/runtime/ws.ts` | WebSocket client |
| `frontend/src/runtime/proto.ts` | Protobuf encode/decode |
| `frontend/src/runtime/registry.ts` | Component registry |
| `frontend/src/runtime/state.ts` | Svelte writable stores |
| `frontend/src/runtime/builtins.ts` | Built-in component auto-registration |
| `frontend/src/component/Button.svelte` | Built-in Button Svelte component |
| `frontend/src/component/Label.svelte` | Built-in Label Svelte component |
| `Makefile` | Build entrypoints |
| `example/main.go` | Example app |
| `example/embed.go` | Static asset embedding |
| `demo/clickcounter/main.go` | Demo app |
| `demo/clickcounter/embed.go` | Demo embed |

---

## Standard 1 — Wire Protocol Consistency

**Assessment: PASS with one dead field**

The Go proto (`proto/ui.proto`) and TypeScript runtime (`frontend/src/runtime/proto.ts`) are in sync:
- `PageState`, `ComponentState`, `ClientEvent`, `StateUpdate` — all 4 core messages match
- `ButtonState` and `LabelState` are in the framework proto and registered in `builtins.ts` under correct type names `ui.ButtonState` / `ui.LabelState`
- Field names use camelCase in TS (auto-converted by protobufjs) matching proto snake_case — correct

**Minor dead field:**  
`ClientEvent.page_id` (proto field 1) is sent from `ws.ts:29` (`{ pageId, componentId, eventType, payload }`) but in `ws.go`, the pageID comes from the URL query parameter (`r.URL.Query().Get("page-id")`) and the `clientEvent.PageId` field is never read. The field is not wrong — it's a harmless redundancy — but it inflates every event frame and creates the illusion of page-level routing in the event handler when the session lookup already handles it.

**No fix required** for correctness; noted as dead weight.

---

## Standard 2 — Error Propagation

**Assessment: FAIL — 4 violations**

### V2-A: `ws.go:105` — WriteMessage error silently discarded

```go
// ws.go:104-106
if sess.conn != nil {
    sess.conn.WriteMessage(websocket.BinaryMessage, updateBytes)  // error discarded
}
```

A failed write (closed connection, network error) is silently ignored. The caller (the browser) never knows if the update was delivered, and the server has no way to detect a half-dead connection.

**Severity:** High — silent data loss on write failure.  
**Fix:** Check the error; log it and clear `sess.conn` if the connection is broken.

### V2-B: `page.go:74` — json.Marshal error silently discarded

```go
// page.go:74
manifestJSON, _ := json.Marshal(manifest)
```

`ComponentManifestEntry` contains only string fields — in practice, `json.Marshal` cannot fail here — but the silent discard is a bad pattern that could mask a future regression.

**Severity:** Low — unreachable in practice with current types, but a bad pattern.  
**Fix:** Check and propagate the error.

### V2-C: `assets.go:73` — fs.Sub error silently discarded

```go
// assets.go:73
staticRoot, _ := fs.Sub(staticFS, "static")
http.Handle("/assets/", http.FileServer(http.FS(staticRoot)))
```

`fs.Sub(staticFS, "static")` can fail if `"static"` does not exist in the FS. With the current embed pattern this is reliable at startup, but a nil `staticRoot` passed to `http.FS` would panic silently at the first asset request, not at startup.

**Severity:** Medium — error surfaces as a panic at request time rather than at startup.  
**Fix:** Check the error and call `log.Fatal` if it fails.

### V2-D: `ws.go:71` — HandleEvent error logged, no client notification

```go
// ws.go:71
if err := handler.HandleEvent(clientEvent.EventType, clientEvent.Payload); err != nil {
    log.Printf("WS event handler error: %v", err)
    continue  // client receives nothing — no state update, no error message
}
```

When a component's event handler returns an error, the server logs it and silently skips the state update. The browser receives no feedback — the UI appears to have worked. The quality standard requires: "Every framework-originated error names the component type, the page ID, and the event that triggered it."

**Severity:** Medium — developer debugging is impaired; the log message omits component type and page ID.  
**Fix:** Improve the log message to include component type, component ID, page ID, and event type.

---

## Standard 3 — Type Safety

**Assessment: FAIL — TypeScript `any` in public API surface**

### V3-A: `registry.ts:3-6` — `any` in public API

```ts
// registry.ts:3-6
export const ComponentRegistry: Record<string, any> = {}
export function registerComponent(typeName: string, ctor: any) {
```

`ComponentRegistry` and `registerComponent` are public API surfaces that app-side code calls directly. Using `any` for Svelte component constructors means no compile-time checking on registration.

**Severity:** Medium — the `any` type propagates into `client.ts` where `ComponentRegistry[entry.type]` is passed to `mount()` without type checking.  
**Fix:** Use `ComponentType` from `svelte` for the constructor type.

### V3-B: `client.ts:21` — unnecessary `as any` cast on `decodePageState` return

```ts
// client.ts:21
const pageState = decodePageState(stateBlob) as any
```

`decodePageState` returns `protobuf.Message`. The subsequent access of `.components` requires either a cast or a typed interface. The cast to `any` is overly broad.

**Severity:** Low — internal to `bootstrap()`, not public API.  
**Fix:** Cast to a typed interface instead of `any`.

### V3-C: `ws.ts:14` — unnecessary `as any` cast on `decodeStateUpdate`

```ts
// ws.ts:14
const update = decodeStateUpdate(evt.data) as any
```

Same pattern as V3-B.

**Severity:** Low — internal only.  
**Fix:** Cast to a typed interface.

### V3-D: `client.ts:15-17` — `(window as any)` casts for injected globals

```ts
const pageId   = (window as any).__SVELGO_PAGE_ID__ as string
const stateBlob = (window as any).__SVELGO_STATE__ as string
const manifest  = (window as any).__SVELGO_MANIFEST__ as ManifestEntry[]
```

These are the standard pattern for injected globals in TypeScript. Using a module augmentation for the `Window` interface is cleaner, but this is acceptable as-is because it is contained in `bootstrap()`. Noted but not a blocking issue.

---

## Standard 4 — Session Lifecycle

**Assessment: FAIL — sessions are never cleaned up**

### V4-A: `session.go` — `Delete()` is never called

`sessionStore.Delete()` is defined but called nowhere in the framework. Every `page.Render()` call registers a new session that persists forever in `globalSessionStore`. On WebSocket disconnect, `ws.go` sets `sess.conn = nil` but does not delete the session. 

This is a **memory leak**: every page load accumulates a `PageSession` (holding the full component map) that is never freed. In a long-running server with normal browser navigation (back button, refresh, multi-tab), sessions grow without bound.

**Severity:** Critical — unbounded memory growth in any real deployment.  
**Fix:** Call `globalSessionStore.Delete(pageID)` when the WebSocket disconnects (in the `defer` block in `ws.go`), and also before registering a replacement if re-render overwrites a session ID (currently impossible since IDs are UUIDs, but defensive).

### V4-B: `ws.go` — no session cleanup on disconnect

```go
// ws.go:37-41
defer func() {
    sess.mu.Lock()
    sess.conn = nil
    sess.mu.Unlock()
}()
```

The defer clears `conn` but does not delete the session from `globalSessionStore`. Combined with V4-A, this means the in-memory component tree for every disconnected tab is permanently retained.

**Fix:** Add `globalSessionStore.Delete(pageID)` to this defer block.

---

## Standard 5 — API Ergonomics

**Assessment: PASS with one noted concern**

The public Go API (`NewPage`, `Add`, `Render`, `Setup`, `SetStaticFS`, `WSHandler`) is minimal and consistent. `page.Render(w, r)` follows `net/http` conventions. `EventHandler` is an optional interface — correct use of Go's implicit interfaces.

**One concern:** `Component.Slot()` returns a fixed string `"root"` in all built-in components. The slot mechanism is exposed in the public `ComponentManifestEntry` struct but the framework provides no way to use multiple named slots without implementing custom components. This could create confusion when an app developer inspects the manifest JSON and sees `"slot": "root"` with no documented way to use other slot values. Not a bug but a friction point for discovery.

---

## Standard 6 — Component Registration

**Assessment: FAIL — silent overwrite in TypeScript registry**

### V6-A: `registry.ts:5-7` — silent overwrite

```ts
export function registerComponent(typeName: string, ctor: any) {
  ComponentRegistry[typeName] = ctor  // no duplicate check
}
```

Registering the same type name twice silently overwrites the first entry. In `builtins.ts`, `svelgo.Button` and `svelgo.Label` are registered on module import. If an app inadvertently calls `registerComponent('svelgo.Button', MyButton)`, it silently replaces the built-in with no warning.

**Severity:** Low-Medium — easy to trigger accidentally; silent failure.  
**Fix:** Warn (console.warn) if a type name is already registered and the constructor differs.

---

## Standard 7 — State Synchronization

**Assessment: FAIL — partial state update on marshal error**

### V7-A: `ws.go:79-107` — partial StateUpdate on marshal error

```go
// ws.go:82-86
stateBytes, err := proto.Marshal(c.ProtoState())
if err != nil {
    log.Printf("WS marshal error for component %s: %v", c.ComponentID(), err)
    continue  // this component is skipped; others are still included
}
```

If any component's `ProtoState()` fails to marshal, that component is skipped via `continue` and the remaining components are still sent. The client receives a `StateUpdate` that is missing one component's state — a partial update. The client will not re-render the failed component, leaving it frozen at its last-known state while other components advance.

The quality standard requires: "State diffs between Go server and browser must be atomic and ordered. No partial updates."

**Severity:** High — silent partial state divergence between server and client.  
**Fix:** Collect all marshal errors first; if any occur, abort the entire update and log a consolidated error rather than sending a partial one.

---

## Standard 8 — Build Reproducibility

**Assessment: PASS with one fragility**

The `Makefile` proto target is deterministic: `protoc` is pinned implicitly via `$PATH`, `pbjs` is run from `node_modules/.bin/` (version-locked by `package.json`). Output paths are explicit.

**One fragility:** The `mv gen/ui.pb.go gen/ui/ui.pb.go` step in the `Makefile` will fail silently or with a confusing error if the `gen/ui/` directory does not exist. A fresh checkout after a `make clean` that removed `gen/` would break `make proto`. The `clean` target does not touch `gen/`, so this is unlikely in practice, but is still a latent issue.

**Fix (minor):** Add `mkdir -p gen/ui` before the `mv` step.

---

## Summary of Violations

| ID | Standard | File | Severity | Status |
|---|---|---|---|---|
| V2-A | Error Propagation | `ws.go:105` | High | FIXED |
| V2-B | Error Propagation | `page.go:74` | Low | FIXED |
| V2-C | Error Propagation | `assets.go:73` | Medium | FIXED |
| V2-D | Error Propagation | `ws.go:71` | Medium | FIXED |
| V3-A | Type Safety | `registry.ts:3-6` | Medium | FIXED |
| V3-B | Type Safety | `client.ts:21` | Low | FIXED |
| V3-C | Type Safety | `ws.ts:14` | Low | FIXED |
| V4-A | Session Lifecycle | `session.go` | Critical | FIXED |
| V4-B | Session Lifecycle | `ws.go:37-41` | Critical | FIXED |
| V6-A | Component Registration | `registry.ts:5-7` | Low-Medium | FIXED |
| V7-A | State Synchronization | `ws.go:79-107` | High | FIXED |
| V8-1 | Build Reproducibility | `Makefile:10` | Low | FIXED |

---

## Fixes Applied

### FIX-V4-A+V4-B: Session cleanup on WebSocket disconnect (`ws.go`)

Added `globalSessionStore.Delete(pageID)` to the disconnect defer block so sessions are freed when the browser tab closes or navigates away.

### FIX-V2-A: WriteMessage error handling (`ws.go`)

`conn.WriteMessage()` return value is now checked. On error, `sess.conn` is cleared and the error is logged with full context (component count, page ID).

### FIX-V2-D: Improved HandleEvent error log (`ws.go`)

Error log now includes component type, component ID, page ID, and event type so developers can locate failures immediately.

### FIX-V7-A: Atomic StateUpdate — abort on partial marshal failure (`ws.go`)

Marshal errors now abort the entire update rather than sending a partial one. If any component fails to marshal, no `StateUpdate` is sent and the full error is logged.

### FIX-V2-B: json.Marshal error check (`page.go`)

`json.Marshal(manifest)` error is now checked and returns an HTTP 500 if it fails.

### FIX-V2-C: fs.Sub error check (`assets.go`)

`fs.Sub(staticFS, "static")` error is now checked with `log.Fatal` so a misconfigured embed fails at startup, not at the first asset request.

### FIX-V3-A: Remove `any` from registry public API (`registry.ts`)

`ComponentRegistry` and `registerComponent` now use `ComponentType` from `svelte` instead of `any`.

### FIX-V3-B+V3-C: Remove unnecessary `as any` casts (`client.ts`, `ws.ts`)

Replaced `as any` casts with typed interfaces (`DecodedPageState`, `DecodedStateUpdate`) for protobuf decoded objects.

### FIX-V6-A: Warn on duplicate component registration (`registry.ts`)

`registerComponent` now emits a `console.warn` if a type name is already registered with a different constructor.

### FIX-V8-1: mkdir -p in proto Makefile target (`Makefile`)

Added `mkdir -p gen/ui` before the `mv` step.
