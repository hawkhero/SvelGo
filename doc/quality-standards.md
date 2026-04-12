# SvelGo Framework Quality Standards

This document is SvelGo's contract with itself — the bar every design decision, API, error, doc, and tool must clear. It is not a style guide and not a wishlist. It is the minimum a serious Go UI framework must meet to be worth adopting over the existing GoTH stack (Go + templ + HTMX + Tailwind), grounded in the community pain points documented in [`golang-ui-framework-demand.md`](./golang-ui-framework-demand.md).

The standards are grouped into four axes:

- **Onboarding DX** — "how easy is it to start?"
- **Go-native composition** — "does this feel like Go, or a port of Rails?"
- **Runtime quality** — "when things break or get deployed, does it still hold up?"
- **Evolution & testability** — "can I live with this framework for two years?"

A serious framework must clear the bar on all four — not just the first. Framework contributors use this document as a design checklist before shipping; application developers use it as an evaluation rubric when reporting friction. The two audiences apply the same bar from different vantage points.

---

## Group A — Onboarding DX (first impression)

### Simple problems must have simple solutions

A developer who wants a button that increments a counter should be running in **one command** after a **one command** scaffold. That is the bar Rails, Django, and Buffalo set. If the simplest use case requires 6 manual steps and 60 lines of boilerplate, the framework has failed the first-impression test — and the Go community, which is already skeptical of frameworks, will not give a second chance.

Ask at every step: **"Is this the simplest thing that could work?"** If a beginner would have to stop and think about where to put this file, it's too complex. Complexity is only acceptable when the problem genuinely requires it.

### Convention over configuration

Go's standard library is opinionated: `net/http`, flat package layout, `go test`, `go build`. SvelGo follows the same philosophy. The common case needs zero configuration. Sensible defaults:

- Static assets go in `static/`
- Frontend source goes in `frontend/`
- Dev server is at `:5173`, app server at `:8080`
- `make dev` starts everything; `make build` produces a single binary

If something every app configures the same way shows up in configuration, it is not configuration — it is a missing convention.

### Scaffolding is not optional

Every mature framework has a `new` command: `rails new`, `django-admin startproject`, `buffalo new`, `create-react-app`. A "Creating a new application" section in the docs with 6 manual steps is direct evidence that the framework is missing its CLI. A developer should be able to run:

```bash
svelgo new myapp
cd myapp
make dev
```

...and see a running app. Anything more is friction that will kill adoption. The Go community pain point is real: developers report fragile Makefiles, multiple watchers, and complex bootstrap as the top reason they don't adopt Go UI frameworks. SvelGo must not repeat that mistake.

---

## Group B — Go-native composition

### Go philosophy applies to framework design

Go was designed with explicit principles: clarity over cleverness, composition over inheritance, no magic. Apply the same lens to SvelGo's API:

- If a feature requires understanding hidden framework internals to use it, it violates "clarity over cleverness"
- If wiring up a page requires importing from 4 different packages, that's a composition failure
- If the framework does something unexpected that isn't obvious from the call site, that's magic — and Go developers hate magic

The Go community has a documented anti-framework culture. The way to win them over is to build something that feels like well-designed Go, not a port of Rails patterns into Go syntax. Think `net/http` — minimal interface, composable, no surprises.

### Compose with `net/http`, don't replace it

`page.Render(w, r)` exists because Go developers already know that signature. Middleware must be `func(http.Handler) http.Handler` — the same shape as every other Go web library. If SvelGo ever invents its own request or response type for the common path, that is a smell. The framework is a library of handlers, not a parallel universe. A developer bringing their own router (chi, gorilla, stdlib `http.ServeMux`) should be able to mount SvelGo pages as handlers without ceremony.

### `context.Context` flows through every boundary

Event handlers receive a `context.Context`. Cancellation, deadlines, and request-scoped values propagate from HTTP → WebSocket → component without framework-specific escape hatches. If a developer has to reach around the framework to get a request's context into a handler — or worse, if the framework drops the context at the WebSocket boundary — that is a bug. Every Go backend already uses context for timeouts, tracing, and auth; a UI framework that ignores this breaks integration with everything the developer already has.

### The Go type system is the API

Event payloads and component state are typed Go structs, not `map[string]any`. Codegen from protobuf exists precisely to make this free at the call site. If user code ever has to type-assert, string-key into framework values, or reach for `interface{}`, the framework has failed to use the type system it was built on top of. The win here is direct: Go's type checker becomes the first line of defense against UI bugs, which is exactly why a Go developer chose Go for the backend in the first place.

---

## Group C — Runtime quality

### Errors must be debuggable in place

Every framework-originated error names the component type, the page ID, and the event that triggered it. No generic "something went wrong" toasts. No errors swallowed at the WebSocket boundary. Go stack traces are preserved end-to-end. A developer staring at a failure should know within 10 seconds where in *their* code to look — not "somewhere in the framework." Given that SvelGo has at least three dispatch layers (HTTP → WebSocket → component event), this is not automatic; it is a commitment the framework has to honor deliberately.

### Dev and prod are the same app

`SVELGO_DEV=1` (Vite proxy) and embedded mode must behave identically for any user-visible feature. A bug that appears in only one mode is a P0 — the worst kind, because it hides until deployment. This is not merely a tactical "watch out"; it shapes every design decision, because any feature that could diverge between the two modes should be redesigned so it cannot.

### Single binary is a load-bearing promise

No runtime Node process. No sidecar. No external asset server in production. This is SvelGo's headline differentiator per the community research — the whole reason a Go shop would pick this over Next.js or Remix. Every design decision must preserve it. If a new feature would require a runtime JS process to work, the feature needs to be redesigned, not the promise weakened. The moment SvelGo ships with "you also need to run X alongside the binary," it has become another GoTH-stack variant.

### The two-codebase tax is real

The market research is clear: developers switched to Go-native solutions specifically to eliminate "the maintenance tax of two separate codebases." Every time SvelGo forces a developer to touch both Go and TypeScript to accomplish one logical task, it's charging that tax. The goal is: **Go struct change → UI change, no TypeScript required** for common cases. Custom components are the escape hatch, not the default path.

---

## Group D — Evolution & testability

### Components are unit-testable without a browser

A standard `go test` can instantiate a `Component`, call `HandleEvent`, and assert the resulting state. If that requires spinning up a WebSocket, a headless browser, or any framework lifecycle beyond constructing the struct, the design is wrong. This is the wedge none of the Go LiveView competitors have cracked — and the Go community will not adopt a UI framework they cannot test with the same tools they test everything else with. Testability here is not a nice-to-have; it is how Go developers decide whether a library is serious.

### Generated code is part of the public API

The `.pb.go` outputs and TypeScript decoder files must read as idiomatic, hand-written-looking code. Awkward names, strange import paths, or generated files a developer has to open to figure out how to call them are framework bugs, not codegen artifacts. From the application developer's seat, the generated code *is* the framework's API surface — it is what they import, read, and call every day. Treat it with the same care as hand-written framework code.

### No hidden global state

No package-level `init()` that silently configures routing. No struct tags that change behavior at a distance. No magic environment variables — `SVELGO_DEV` is a documented exception, not a pattern to repeat. Everything wired is wired visibly at the call site. This is Go philosophy applied directly: if a reader cannot trace the app's behavior from `main.go` downward, the framework has broken the "no magic" rule that Go developers care about most. "Convention over configuration" must mean *visible* conventions, not invisible ones.

---

## How this document is used

**Framework contributors (`svelgo-dev`):** these standards are the bar every design must clear. Before shipping an API, a doc, or a tool change, check it against every relevant standard. When a standard conflicts with implementation simplicity or a deadline, flag the tradeoff explicitly in the response and in `FRICTION.md` — do not ship and hope no one notices. If you believe a standard is wrong, say so and propose an amendment; do not silently violate it.

**Application developers (`go-web-dev`):** these standards are the bar you hold the framework to. When something falls short — even if it "works" — log a concrete `FRICTION.md` entry that cites the standard it violates. Do not rationalize failures away. Your fresh-eyes view from the consumer side is the ground truth that catches gaps the architect could not see from the inside.

**Both roles share the same standards but apply them from different vantage points.** The framework architect asks "does my design meet this?" — the application developer asks "does the framework I'm using meet this?" The gap between those two answers is exactly what the `FRICTION.md` loop exists to surface.
