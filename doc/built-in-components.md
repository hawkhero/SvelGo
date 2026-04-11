# Built-in Components

SvelGo ships two built-in components in the `github.com/hawkhero/svelgo/component` package: `Button` and `Label`. They cover the most common UI primitives and require zero frontend registration — no entries in `registry.ts` or `proto.ts` are needed.

---

## Auto-registration

The framework's TypeScript runtime (`frontend/src/runtime/builtins.ts`) registers both components automatically when `bootstrap()` runs:

```ts
registerComponent('svelgo.Button', Button)
registerComponent('svelgo.Label', Label)
registerComponentDecoder('svelgo.Button', root.lookupType('ui.ButtonState'))
registerComponentDecoder('svelgo.Label', root.lookupType('ui.LabelState'))
```

This happens before any app code runs. For a built-in-only app, `registry.ts` and `proto.ts` do **not** need to exist at all — and `main.ts` should **not** import them. The minimal `main.ts` is:

```ts
import { bootstrap } from 'svelgo/runtime/client'

bootstrap()
```

`registry.ts` and `proto.ts` are only created (and only imported from `main.ts`) when your app adds custom Svelte components. See the "Adding a custom component" section in `doc/getting-started.md` for that workflow.

---

## component.Button

**Package:** `github.com/hawkhero/svelgo/component`

**Go struct:**

```go
type Button struct {
    ID       string   // required — must be unique on the page
    Label    string   // text shown on the button face
    Disabled bool     // when true, the button is greyed out and ignores clicks
    OnClick  func()   // called on the server when the user clicks the button
}
```

**ComponentType string:** `"svelgo.Button"`

**Proto state (ui.proto):**

```proto
message ButtonState {
  string label    = 1;
  bool   disabled = 2;
}
```

Only `Label` and `Disabled` are sent over the wire. `ID` stays on the server; `OnClick` is a Go closure.

**Behaviour:**
- A click in the browser sends a `"click"` event over the WebSocket.
- The framework calls `HandleEvent("click", nil)` on the Go struct.
- If `OnClick` is non-nil, it is invoked.
- After `OnClick` returns, the framework serializes the struct's current state (`Label`, `Disabled`) and sends a `StateUpdate` to the browser.
- The Svelte component re-renders with the new label text and disabled state.

**Example — disable after first click:**

```go
btn := &component.Button{
    ID:    "submit-btn",
    Label: "Submit",
}
btn.OnClick = func() {
    btn.Label    = "Submitted"
    btn.Disabled = true
}
page.Add(btn)
```

---

## component.Label

**Package:** `github.com/hawkhero/svelgo/component`

**Go struct:**

```go
type Label struct {
    ID   string   // required — must be unique on the page
    Text string   // text content to display
}
```

**ComponentType string:** `"svelgo.Label"`

**Proto state (ui.proto):**

```proto
message LabelState {
  string text = 1;
}
```

**Behaviour:**
- Label is a display-only component. It implements `Component` but not `EventHandler`.
- Its initial text is set at render time and injected into the page as initial state.
- A Label's `Text` can be mutated from any other component's callback (e.g., a Button's `OnClick`). The framework sends updated states for all components on the page in a single WebSocket frame after each event, so the client sees the new text immediately.

---

## Wiring Button and Label together

A Button's `OnClick` callback can mutate any component's state — including a separate Label. The framework sends back updated states for **all** components in the session in one WebSocket frame, so cross-component mutations work as expected.

### Example: click counter

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
    svelgo.Setup()

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Allocate state variables that live for the duration of this page session.
        count := 0

        btn := &component.Button{
            ID:    "counter-btn",
            Label: "Click me (0 clicks)",
        }
        lbl := &component.Label{
            ID:   "counter-label",
            Text: "Count: 0",
        }

        // Wire the callback: update both the button label and the separate Label.
        btn.OnClick = func() {
            count++
            btn.Label = fmt.Sprintf("Click me (%d clicks)", count)
            lbl.Text  = fmt.Sprintf("Count: %d", count)
        }

        page := svelgo.NewPage()
        page.Add(btn).Add(lbl)
        page.Render(w, r)
    })

    log.Println("Listening on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**What happens on each click:**

1. User clicks the button in the browser.
2. Browser sends a binary `ClientEvent{page_id, component_id: "counter-btn", event_type: "click"}` over the WebSocket.
3. Go's `WSHandler` looks up `"counter-btn"` in the `PageSession` and calls `HandleEvent("click", nil)`.
4. `Button.HandleEvent` calls `btn.OnClick()`.
5. `OnClick` increments `count`, updates `btn.Label`, and updates `lbl.Text`.
6. The framework serializes the current state of **every component on the page** and sends a single `StateUpdate` frame containing all of them.
7. The browser receives the frame, decodes it, and patches the Svelte stores for both `"counter-btn"` and `"counter-label"`. Both components re-render.

### Important: state lives in the handler closure

Notice that `count`, `btn`, and `lbl` are all local variables inside the `http.HandleFunc` closure. The framework's `PageSession` holds live pointers to the `btn` and `lbl` structs. Each HTTP request gets its own independent set of variables and its own `PageSession`. Two browser tabs open to the same URL have completely separate state.

---

## Component IDs

Component IDs must be:
- **Unique within a page** — two components with the same ID on the same page will collide in the browser's store map.
- **Stable within a session** — the ID in the initial HTML must match the ID in every subsequent `StateUpdate`.

String literals like `"counter-btn"` are fine. If you render a list of components dynamically, use a deterministic scheme such as `fmt.Sprintf("item-%d", i)`.

---

## Slot

Both built-in components return `"root"` from their `Slot()` method. The framework mounts them in order inside the `<div id="svelgo-root">` element. Slot-based layout (grid, sidebar, etc.) is not yet implemented — all components render in a single vertical stack.
