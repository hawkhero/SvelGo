# Friction Log

Shared log between `go-web-dev` (app developer) and `svelgo-dev` (framework architect).

When building an app on SvelGo surfaces a pain point — verbose API, missing abstraction, confusing error, awkward generated code, dev/prod divergence — record it here as a concrete entry. The framework architect picks entries up, proposes fixes, and marks them resolved. History is never deleted: resolved entries stay so future work can see which tradeoffs were made and why.

## Entry format

```
### [OPEN] Short title
**Reported by:** go-web-dev · <date>
**Context:** What I was trying to build.
**Current pain:** What the framework made me do.
**Desired API:** What I wish I could write instead.
**Notes:** Anything else — workarounds tried, related entries, etc.
```

When resolved, change `[OPEN]` to `[RESOLVED]` and append:

```
**Resolution:** <summary> · <commit/PR ref> · <date>
```

---

## Resolved

### [RESOLVED] `make dev` panics on a fresh checkout — `make proto` must be run first
**Reported by:** go-web-dev · 2026-04-11
**Context:** Following the getting-started tutorial to run the example app for the first time (`make dev`).
**Current pain:** The Go server immediately panics with `slice bounds out of range [-1:]` inside `gen/ui/ui.pb.go` during protobuf descriptor initialization. The app never starts. The fix is to run `make proto` first, but the getting-started guide never mentions this step. The repo ships with a `ui.pb.go` that is either stale or uses `unsafe.StringData` on a const string in a way the current runtime rejects.
**Desired API:** Either the repo should ship a valid pre-generated `ui.pb.go` that works without running `make proto`, or the tutorial's very first step should say: "Run `make proto` once before `make dev`."
**Notes:** The panic stack is in `filedesc.(*File).unmarshalSeed` → `file_ui_proto_init()`. After running `make proto` (which regenerates `gen/ui/ui.pb.go` from `proto/ui.proto`), the panic disappears and the app starts normally. This is the single highest-friction moment on day 1 — a new developer would have no idea where to look.

**Resolution:** Added a prominent callout block to `doc/getting-started.md` in the Dev Mode section, immediately before the `make dev` instruction. The callout explains the `make proto` prerequisite, names the panic it prevents, and clarifies that it is a one-time step unless `proto/ui.proto` changes. The shipped `gen/ui/ui.pb.go` is not modified — regenerating it requires the local `protoc` toolchain, so fixing the doc is the correct minimal action. · 2026-04-11
**Verified:** Confirmed. `doc/getting-started.md` lines 154–162 contain a clearly marked `> **Important**` callout block immediately before the `make dev` heading. The callout names the `slice bounds out of range` panic and explains the one-time `make proto` step. Fix is correct and complete. · go-web-dev · 2026-04-11

---

### [RESOLVED] Label does not update when mutated from a Button's OnClick
**Reported by:** go-web-dev · 2026-04-11
**Context:** Building a click counter where a `component.Label` shows the current count and a `component.Button` shows it in the button face — the mandatory baseline app.
**Current pain:** The `WSHandler` only serializes and sends back the state of the component that received the event. Mutating `lbl.Text` inside `btn.OnClick` compiles and runs silently — the server-side value is updated — but the client never sees it. The Label stays frozen at "Count: 0" while the Button correctly counts up. The task requires both to update; only one can.
**Desired API:** Either (a) `WSHandler` should detect all dirty components and send a multi-component `StateUpdate` in one frame, or (b) provide an explicit API like `page.MarkDirty(lbl)` / `page.PushAll()` that the handler can call to flush additional component states.
**Notes:** The limitation is documented in `doc/built-in-components.md`, but only _after_ the full code example is shown. A reader who follows the example code expecting both to work will be confused. The workaround — embedding the count in the button's own label — defeats the purpose of having a separate Label component. The fix requires changing only `ws.go` (lines 77–92): iterate all components in the session and include any that were mutated.

**Resolution:** Fixed in `ws.go`. The state-response block now acquires `sess.mu`, iterates all components in `sess.components`, marshals each one's `ProtoState()`, and builds a single `StateUpdate` containing every component. The lock is held across the full marshal-and-write sequence to prevent races. Both framework and example app compile clean after this change. · 2026-04-11
**Verified:** Confirmed — with one regression noted separately. `ws.go` lines 79–107 correctly acquire `sess.mu` once, iterate all entries in `sess.components`, marshal each via `c.ProtoState()`, and send a single `StateUpdate` with all of them. The lock is held across the full marshal-and-write block; this is race-safe. The ws.go fix itself is correct and complete. However, `example/main.go` lines 31–33 still contain a stale comment (`// NOTE: lbl.Text mutation does not propagate to the client. // WSHandler only sends back the receiving component's state.`) that contradicts the fixed behaviour — see new OPEN entry below. · go-web-dev · 2026-04-11

---

### [RESOLVED] Tutorial shows Label mutation in OnClick but buries the "won't work" caveat
**Reported by:** go-web-dev · 2026-04-11
**Context:** Reading `doc/built-in-components.md` to understand how to wire Button + Label together.
**Current pain:** The docs show this code in the example: `lbl.Text = fmt.Sprintf("Count: %d", count)` inside `btn.OnClick`, immediately followed (in the same code block) by a comment `// lbl.Text update is set at render time only (see below for caveat)`. The caveat blockquote appears _after_ the full code listing. A developer reading linearly will copy the code, run it, and wonder why the Label never updates — because the most important constraint is presented after the code that silently violates it.
**Desired API:** The caveat should appear _before_ the code example, ideally in a callout box. The code example itself should not show `lbl.Text` mutation at all if it does not work — or it should be in a separate "what doesn't work yet" section.
**Notes:** Related to the single-component-update limitation above. This is a documentation structure issue, not a code issue.

**Resolution:** Rewrote the "Wiring Button and Label together" section in `doc/built-in-components.md`. The old caveat blockquote and "won't work" language are removed because the underlying bug is now fixed (see entry above). The section now leads with a clear statement that cross-component mutations work, followed by a clean code example without misleading inline comments. The Label's Behaviour section is also updated to reflect the fixed semantics. · 2026-04-11
**Verified:** Confirmed. `doc/built-in-components.md` contains no "won't work" caveat anywhere. The Label Behaviour section (lines 109–112) correctly states that `Text` can be mutated from any other component's callback and that the framework sends all component states in a single frame. The "Wiring Button and Label together" section (lines 115–178) leads with an affirmative statement, presents a clean code example with no misleading inline comments, and includes a step-by-step "What happens on each click" walkthrough. Fix is correct and complete. · go-web-dev · 2026-04-11

---

### [RESOLVED] `go.mod` declares `go 1.26` (unreleased version)
**Reported by:** go-web-dev · 2026-04-11
**Context:** Setting up the project, reading `go.mod`.
**Current pain:** The framework's `go.mod` says `go 1.26`. As of 2026-04-11, Go 1.26 is the current stable release, but any developer on Go 1.24 or 1.25 will get a toolchain mismatch error or unexpected behavior when using `replace` directives. There is no note in the tutorial about the minimum Go version for the framework vs. the example app.
**Desired API:** The `go.mod` should match the minimum version stated in the prerequisites table (`go 1.21`), or the prerequisites table should be updated to reflect the actual requirement. A mismatch between the doc's stated minimum ("Go 1.21+") and `go.mod`'s declared version creates confusion.
**Notes:** The getting-started guide's prerequisite table says "Go 1.21+". The actual `go.mod` says `go 1.26`. These are inconsistent. A developer with Go 1.21–1.25 would hit cryptic toolchain errors.

**Resolution:** Updated the prerequisites table in `doc/getting-started.md` to say "Go 1.26+" — matching `go.mod`. Lowering `go.mod` to `1.21` was rejected because it would silently permit use of a toolchain that cannot satisfy the module's actual dependency graph (the `google.golang.org/protobuf v1.36.11` transitive chain uses features from recent Go releases). The correct fix is to align the doc to the truth, not to weaken the module declaration. · 2026-04-11
**Verified:** Confirmed. `go.mod` line 3 declares `go 1.26`. `doc/getting-started.md` prerequisites table (line 13) states `Go | 1.26+`. Both agree. Fix is correct and complete. · go-web-dev · 2026-04-11

---

### [RESOLVED] `make dev` interleaves Go panic output with Vite startup output — hard to diagnose
**Reported by:** go-web-dev · 2026-04-11
**Context:** Running `make dev` for the first time and getting a panic.
**Current pain:** `make dev` runs `npm run dev` in the background and `go run .` in the foreground. When Go panics immediately, the output arrives interleaved with Vite's startup banner. The Vite "ready" message appears just before the Go panic stack trace, making it look like the frontend started successfully but something else failed. The make target also exits with a non-zero code but without a clear "the Go server crashed — look above for the panic" summary line.
**Desired API:** Either run Go first (and abort if it fails) before starting Vite, or prefix Go output with a `[go]` tag so failure is visually obvious. A simple health check — wait for Go's "Listening on :8080" before printing "dev ready" — would save minutes of debugging on day 1.
**Notes:** This compounds the proto regeneration friction above. Two things fail silently together.

**Resolution:** Rewrote the `dev` target in `example/Makefile`. Go now starts first in the background; the target then polls `http://localhost:8080/` in 0.5 s increments (up to 10 s) before launching Vite. If Go exits before the health check passes, the target prints a clear `[go] Go server exited unexpectedly` message and aborts with a non-zero exit code so Vite never starts. Output from each phase is prefixed with `[go]` / `[vite]` so it is visually distinct. Tradeoff: the 10 s poll window adds ~0.5 s to a clean startup; this is acceptable for a dev workflow. · 2026-04-11
**Verified:** Confirmed. `example/Makefile` `dev` target (lines 15–27) starts Go first in the background (`SVELGO_DEV=1 go run . & GO_PID=$$!`), polls `http://localhost:8080/` with `curl -sf` in 0.5 s increments up to 20 iterations (10 s), checks `kill -0 $$GO_PID` to detect early Go exit and prints `[go] Go server exited unexpectedly` before aborting. Vite starts only after the health-check loop completes. Fix is correct and complete. · go-web-dev · 2026-04-11

---

## Open

### [RESOLVED] `example/main.go` contains stale comment saying Label update "does not propagate to the client"
**Reported by:** go-web-dev · 2026-04-11
**Context:** Re-reading `example/main.go` after the `ws.go` fix landed (Entry #2).
**Current pain:** `example/main.go` lines 31–33 still read:
```go
// NOTE: lbl.Text mutation does not propagate to the client.
// WSHandler only sends back the receiving component's state.
// This is a known framework limitation — see FRICTION.md.
```
This comment was written to document the original ws.go bug. That bug is now fixed. The comment is now factually wrong — `lbl.Text` mutations *do* propagate to the client — but a developer reading the example app will conclude the feature still doesn't work. This directly contradicts the updated `doc/built-in-components.md` and will cause confusion on day 1 for anyone who reads the source rather than the docs.
**Desired API:** Remove lines 31–33 entirely. The example app should be the clean, working baseline — no apologetic comments referencing FRICTION.md entries that are already resolved.
**Notes:** The three stale comment lines are the only thing preventing `example/main.go` from being the canonical baseline app that the mandatory click-counter test requires. The fix is a two-line delete.

**Resolution:** Deleted lines 31–33 from `example/main.go` — the three-line `// NOTE:` comment block referencing the old framework limitation. The `lbl.Text` assignment remains; the code is self-explanatory. Example compiles clean. · 2026-04-11
**Verified:** Confirmed. `example/main.go` contains no `// NOTE:` comment block. The `OnClick` handler (lines 28–33) is clean: it increments `clickCount`, updates `btn.Label`, updates `lbl.Text`, and logs — no apologetic prose, no references to FRICTION.md, no workaround language. Both Button and Label are wired correctly and the file is the canonical baseline click-counter app. Fix is correct and complete. · go-web-dev · 2026-04-11
