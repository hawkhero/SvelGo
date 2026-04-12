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

---

### [RESOLVED] `package.json` template uses a literal placeholder path for the `svelgo` npm dependency
**Reported by:** go-web-dev · 2026-04-12
**Context:** Scaffolding `demo/clickcounter/` from scratch by following `doc/getting-started.md` Step 3.
**Current pain:** The `package.json` template in the doc shows `"svelgo": "file:/path/to/svelgo/frontend"`. That `/path/to/svelgo/frontend` is not an example of a real relative path — it is a literal placeholder. A developer must figure out the correct relative path for their own directory layout without any guidance. For an app nested two levels under the framework repo (`demo/clickcounter/frontend/`), the correct path is `../../../frontend`, which is non-obvious and easy to get wrong.
**Desired API:** Replace the placeholder with a note explaining the convention: `"svelgo": "file:../../frontend"  // relative from <your-app>/frontend/ to <framework-root>/frontend/`. Alternatively, document the standard expected nesting depth (app sits one level under the framework root, as in `example/`) so the relative path is always `../../frontend`.
**Notes:** The `example/` app uses `"svelgo": "file:../../frontend"` in its `package.json`. That pattern is never stated in the tutorial. A developer who places their app two levels deep (e.g. `demo/clickcounter/`) must use `../../../frontend` and has no doc to confirm it. The workaround is to look at `example/frontend/package.json`, which the tutorial explicitly says not to copy from.

**Resolution:** Replaced the `file:/path/to/svelgo/frontend` placeholder in the `doc/getting-started.md` Step 3 `package.json` template with the concrete value `file:../../frontend` (the standard one-level-under-framework-root layout). Added a prose callout immediately after the code block that explains the path convention: relative from `<your-app>/frontend/` to `<framework-root>/frontend/`, with the rule "add one extra `../` per additional nesting level" and an explicit example for `demo/clickcounter/` (`../../../frontend`). · 2026-04-12
**Verified:** Confirmed. `doc/getting-started.md` Step 3 `package.json` template (line 220) uses `"svelgo": "file:../../frontend"` — no placeholder. The prose callout immediately after (line 228) states the path convention and names `../../../frontend` for the `demo/clickcounter/` two-level nesting. `demo/clickcounter/frontend/package.json` uses `"svelgo": "file:../../../frontend"` — correct for its depth and consistent with the doc example. Fix is correct and complete. · go-web-dev · 2026-04-12

---

### [RESOLVED] Vite emits "no Svelte config found" warning on every dev startup
**Reported by:** go-web-dev · 2026-04-12
**Context:** Running `npm run dev` in `demo/clickcounter/frontend/` after following the tutorial verbatim.
**Current pain:** Every Vite dev-server startup prints: `[vite-plugin-svelte] no Svelte config found at <path> — using default configuration.` The tutorial never mentions `svelte.config.js` and does not provide one in the scaffolded file list. The warning is harmless (defaults are fine for built-in-component-only apps), but it looks like an error to a new developer and clutters the output.
**Desired API:** Either add a minimal `svelte.config.ts` (even just `export default {}`) to the Step 3 file listing, or add a one-line note: "If you see a 'no Svelte config found' warning, it is safe to ignore — the defaults are correct for this app."
**Notes:** The `example/` app also lacks a `svelte.config.js` and emits the same warning. This is a cosmetic issue but produces needless friction on day 1.

**Resolution:** Added `svelte.config.ts` (content: `export default {}`) to the Step 3 scaffolded file listing in `doc/getting-started.md`, with a prose note explaining that this file suppresses the warning and that the empty export is correct. Also created `demo/clickcounter/frontend/svelte.config.ts` with the same content so the already-scaffolded demo app is consistent with the updated tutorial. The "safe to ignore" approach was rejected because showing the fix is strictly better — it eliminates the warning rather than normalising it. · 2026-04-12
**Verified:** Confirmed. `doc/getting-started.md` Step 3 directory listing (line 98) includes `svelte.config.ts` with the annotation "minimal Svelte config (prevents vite-plugin-svelte warning)". Lines 234–240 show the file content (`export default {}`) and prose explaining why it is needed. `demo/clickcounter/frontend/svelte.config.ts` exists on disk and contains exactly `export default {}`. Fix is correct and complete. · go-web-dev · 2026-04-12

---

### [RESOLVED] Tutorial is ambiguous about whether `proto.ts` and `registry.ts` should exist at all for built-in-only apps
**Reported by:** go-web-dev · 2026-04-12
**Context:** Scaffolding `demo/clickcounter/` which uses only `component.Button` and `component.Label` — no custom components.
**Current pain:** Step 3's directory layout lists `proto.ts` and `registry.ts` as `← optional` files under `src/`. Step 4 says these are "unnecessary" for built-in-only apps. But the custom-component section (Step 4 of the "Adding a custom component" workflow) shows `main.ts` importing both files (`import './proto'`; `import './registry'`) before `bootstrap()`. A developer following the full tutorial may create those import lines in `main.ts` proactively — which breaks the build with "cannot find module './proto'" because the files were never created. The two tutorial sections use inconsistent patterns.
**Desired API:** Step 3's directory listing should either (a) omit `proto.ts` and `registry.ts` entirely from the built-in-only scaffolding, or (b) provide their content (empty or comment-only) so the developer knows exactly what to create. The custom-component section's `main.ts` snippet should note that the `import './proto'` lines are only added when those files actually exist.
**Notes:** The working pattern (confirmed by running the app) is: `main.ts` contains only `import { bootstrap } from 'svelgo/runtime/client'; bootstrap()` — no other imports. The `proto.ts` and `registry.ts` files simply do not exist. The tutorial's directory layout creates the impression that they should always be present.

**Resolution:** Three coordinated changes: (1) Step 3 directory listing in `doc/getting-started.md` now annotates `proto.ts` and `registry.ts` as "only needed for custom components; do not create for built-in-only apps", and adds a callout block making the rule explicit with a build-fail warning. (2) The custom-component Step 6 `main.ts` snippet in `doc/getting-started.md` now carries a clear note that the `import './proto'` and `import './registry'` lines are only correct once those files exist, and must not be added to a built-in-only app. (3) `doc/built-in-components.md` Auto-registration section is rewritten: removed the misleading "can be empty" language and the placeholder file content snippets; now states definitively that `proto.ts` and `registry.ts` do not need to exist and `main.ts` should not import them for built-in-only apps. · 2026-04-12
**Verified:** Confirmed — all three coordinated changes are present and correct. (1) `doc/getting-started.md` Step 3 directory listing (lines 103–104) annotates `proto.ts` and `registry.ts` as "only needed for custom components; do not create for built-in-only apps"; the callout block (lines 108) explicitly warns that importing absent files breaks the build. (2) Step 6 `main.ts` snippet (lines 546–554) carries a clear note that the `import './proto'` and `import './registry'` lines must not be added to a built-in-only app and are only correct once those files exist. (3) `doc/built-in-components.md` Auto-registration section (lines 18–26) states definitively that `registry.ts` and `proto.ts` do not need to exist and `main.ts` should not import them for built-in-only apps. `demo/clickcounter/frontend/src/main.ts` contains only `import { bootstrap } from 'svelgo/runtime/client'` and `bootstrap()` — no `proto`/`registry` imports. Fix is correct and complete. · go-web-dev · 2026-04-12

---

### [RESOLVED] Tutorial recommends `svelte.config.ts` but vite-plugin-svelte cannot load `.ts` config files in ESM mode
**Reported by:** go-web-dev · 2026-04-12
**Context:** Scaffolding `demo/clickcounter/` following `doc/getting-started.md` Step 3, which says to create `frontend/svelte.config.ts` with `export default {}`.
**Current pain:** Creating `svelte.config.ts` exactly as documented causes the Vite dev server to abort with: `TypeError [ERR_UNKNOWN_FILE_EXTENSION]: Unknown file extension ".ts" for .../svelte.config.ts`. The `vite-plugin-svelte` plugin tries to `import()` the config file using Node's native ESM loader, which cannot handle `.ts` files without a TypeScript loader registered. The dev server never starts. The fix is to rename the file to `svelte.config.js` — a `.js` extension works because Node can load it as ESM without a transpiler. The example app does not ship a svelte config at all, so this issue only appears when following the tutorial's recommendation verbatim.
**Desired API:** The tutorial's Step 3 should recommend `svelte.config.js` instead of `svelte.config.ts`. The extension matters: `.ts` breaks, `.js` works. One character difference, complete failure vs. success.
**Notes:** Violates Quality Standard A1 ("simple problems must have simple solutions") — an instruction in the tutorial that is copied verbatim causes immediate dev-server failure with a cryptic Node error. The developer has no indication the tutorial is wrong. Workaround applied: created `svelte.config.js` with `export default {}` — this suppresses the warning and the dev server starts cleanly.

**Resolution:** Changed all three occurrences of `svelte.config.ts` in `doc/getting-started.md` to `svelte.config.js`: the directory listing (line 98), the section heading (line 234), and the code-block language tag (line 236). No framework Go files were touched. `demo/clickcounter/frontend/svelte.config.js` already existed with the correct extension and content — no change needed there. The `example/` app ships no svelte config at all and is unaffected. · 2026-04-12
**Verified:** Confirmed. `doc/getting-started.md` lines 98 and 234 reference `svelte.config.js` — no `.ts` anywhere in the file. `demo/clickcounter/frontend/svelte.config.js` exists on disk with `export default {}`. Dev server started with `node node_modules/vite/bin/vite.js`; Vite output contains no `ERR_UNKNOWN_FILE_EXTENSION` error and no "no svelte config found" warning. Go server HTML at port 8080 contains `__SVELGO_STATE__` with both `svelgo.Button` (id: `counter-btn`) and `svelgo.Label` (id: `counter-label`) in the manifest. Fix is correct and complete. · go-web-dev · 2026-04-12

---

## Open

### [RESOLVED] `npm run dev` fails — `.bin/vite` shim resolves `cli.js` relative to `.bin/` instead of the package directory
**Reported by:** go-web-dev · 2026-04-12
**Context:** Running `npm run dev` in `demo/clickcounter/frontend/` after `npm install`.
**Current pain:** `npm run dev` (which calls `.bin/vite`) immediately exits with: `Error [ERR_MODULE_NOT_FOUND]: Cannot find module '.../node_modules/dist/node/cli.js'`. The `.bin/vite` shim is a plain copy of `node_modules/vite/bin/vite.js`, not a symlink. Because the shim runs from the `.bin/` directory, the relative import `../dist/node/cli.js` resolves to `node_modules/dist/node/cli.js` instead of `node_modules/vite/dist/node/cli.js`. Running `node node_modules/vite/bin/vite.js` directly works correctly because that script runs with its own directory as the base, not `.bin/`. Re-running `npm install` does not fix it — it reports "up to date" and the shim remains a copied file.
**Desired API:** `npm run dev` should work out of the box. The `.bin/vite` entry should be a proper symlink to `../vite/bin/vite.js` so the relative path resolves correctly regardless of which directory npm uses as the working directory for the shim. Alternatively, the `package.json` `scripts.dev` entry should call `node node_modules/vite/bin/vite.js` directly as a workaround until the symlink issue is resolved.
**Notes:** This appears to be an npm/Node version interaction issue where npm creates a copied shim instead of a symlink on this platform (Node v22.16.0, macOS Darwin). The tutorial's `npm run dev` instruction silently fails with a cryptic Node ESM error that looks unrelated to Vite. A developer following the tutorial verbatim will be blocked. Workaround: run `node node_modules/vite/bin/vite.js` from the `frontend/` directory directly. This is the second npm toolchain issue encountered (after the `file:` path placeholder); the pattern suggests the tutorial's frontend setup section needs an explicit troubleshooting note for Node v22+ on macOS.

**Resolution:** Root cause identified: the `node_modules/` directory was migrated with `cp -r`, which dereferences symlinks into plain file copies. A fresh `rm -rf node_modules && npm install` in `demo/clickcounter/frontend/` creates the correct symlink (`node_modules/.bin/vite -> ../vite/bin/vite.js`) and `npm run dev` works. No framework or tutorial changes needed — this is an artifact of a `cp -r` migration, not a structural issue with npm on macOS. The demo app's `node_modules/` has been reinstalled from scratch and is now clean. · 2026-04-12
