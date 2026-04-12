# Phase 1 Interaction Loop: svelgo-dev ↔ go-web-dev

## Context

Two subagents now exist — `svelgo-dev` (framework architect) and `go-web-dev` (senior app developer) — plus a shared `FRICTION.md` at the repo root. The agent files are done; what's missing is **a concrete workflow that makes them actually collaborate**.

Constraints for the first phase:
- Keep it small
- Use only the two existing built-in components (`component/button.go`, `component/label.go`)
- Validate that the basic architecture — *agent → task → friction log → fix → re-verify* — actually works
- Start from the simplest thing

Phase 1 is not about shipping features. **It is about proving the supplier/consumer loop itself runs end-to-end on a task small enough that any failure is clearly a loop failure, not a task-complexity failure.**

## Current baseline (observed, not planned)

- `example/main.go` mounts one `component.Button` only; clicking it mutates `btn.Label` in the handler.
- `component.Label` exists but is **not used anywhere** in the example app.
- `example/frontend/src/components/` is empty — the built-in Button/Label must be registered on the frontend from the framework side (or they are not, in which case that is already a friction point).
- `FRICTION.md` exists with an Open / Resolved structure and is empty.

## The Loop (Phase 1 version — human-routed, agent-driven reads)

Four steps (Step 0 is one-time setup, then Steps 1–3 form the repeating loop). The human (user) drives each handoff explicitly; agents do not auto-invoke each other. But agents read `FRICTION.md` and `doc/` themselves — the human does not summarize or filter.

### Step 0 — Write the tutorial (one-time prep)
Invoke `svelgo-dev` with: *"Before any app work begins, write the initial developer-facing tutorial under `doc/` that teaches a senior Go developer how to use SvelGo **as the framework exists right now**. Base it strictly on the current code — do not document APIs that don't exist. At minimum cover: project layout, dev/prod modes, the 6-step component workflow, and how to use the two built-in components (`component.Button`, `component.Label`) in an app. A reader who knows Go but has never seen SvelGo should be able to build a trivial working app from your doc alone. **Self-check before declaring done:** build a minimal example app by following only your own tutorial (no peeking at framework internals), run `make dev`, and verify it works in a browser. If a step in your tutorial is wrong, missing, or ambiguous, fix the doc before handing off."* Expected output:
- One or more tutorial files under `doc/` (e.g. `doc/getting-started.md`, `doc/built-in-components.md`)
- A self-verified minimal example proving the tutorial is runnable
- No framework code changes in this step — docs only
- A short summary of which files were written and what they cover

This is the **source of truth** `go-web-dev` will read in Step 1. It also acts as a baseline sanity check on `svelgo-dev`'s new documentation responsibility — if the doc cannot be written without hand-waving, that itself is the first friction signal.

### Step 1 — Build & log
Invoke `go-web-dev` with: *"Read `doc/` first. Then build the **mandatory baseline app** from scratch in a new directory `demo/clickcounter/` under the repo root (sibling of `example/`). Follow `doc/getting-started.md`'s "Creating a new application" section step by step — do NOT copy from `example/`. Start from a blank directory. Use only the tutorial and `doc/built-in-components.md` — no guessing at APIs, no reading framework internals to fill gaps. Every pain point, missing doc, awkward API, or dev/prod divergence must be logged as a concrete entry in `FRICTION.md`. Pain points about the tutorial itself (incorrect, incomplete, or misleading steps in `doc/`) count and must be logged."*

**Mandatory baseline app:** a click counter built in `demo/clickcounter/` — one `component.Button` and one `component.Label` wired together, where each click increments a counter whose value is shown in the Label and also reflected in the Button's own label. The app must be scaffolded from zero (blank directory, fresh `go mod init`, fresh `npm init`) following only the doc. This is fixed so that any loop failure is unambiguously a loop/framework/doc defect, not an app-choice problem.

**Optional additional apps (0–2):** go-web-dev picks any trivially simple app that uses only the two built-in components, each in its own subdirectory under `demo/`. Examples: a "label echoes the last button clicked" two-button demo, a two-button toggle that flips a label between two states. These are optional and exist only to broaden the surface area a little.

Constraints:
- No edits to `*.go` at repo root, `frontend/src/runtime/`, `component/`, or `proto/ui.proto`.
- No copying from `example/` — each demo app must be scaffolded from scratch.
- No new components, no new proto messages.
- Must work end-to-end in dev mode, verified in a browser.
- Every pain point — however small — must be written to `FRICTION.md` as a concrete entry.

Expected output:
- The mandatory baseline app at `demo/clickcounter/` + 0–2 optional apps under `demo/`, all working (or explicit reports of what failed and why)
- Zero-to-many new entries appended to `FRICTION.md` under **Open**
- A short summary of which apps were built and what felt wrong
- **Quality standards audit:** after completing the app(s), read `doc/quality-standards.md` end-to-end and produce a per-standard evaluation (PASS / FAIL / N/A with a one-line rationale for each). Every FAIL must have a corresponding `[OPEN]` entry in `FRICTION.md` that cites the standard by name. Write the audit to `demo/clickcounter/quality-audit.md`.

### Step 2 — Triage & fix
Invoke `svelgo-dev` with: *"Read `FRICTION.md`. For every Open entry, either resolve it in the framework/doc or close it with a written justification. Update `FRICTION.md` in place — mark each handled entry `[RESOLVED]` with a resolution note. Verify the existing example apps still build and run after your changes. **Justifications are not a lazy escape hatch** — if you close an entry without a code/doc change, the resolution note must name a specific tradeoff (API stability, scope boundary, intentional design choice, deferred to Phase 2+) and explain *why* the friction is acceptable. 'Works as intended' alone is not a valid justification."* Expected output:
- Framework/doc changes **or** justifications with explicit tradeoffs appended to entries
- Every previously Open entry now marked `[RESOLVED]` with a short resolution note
- Verification that the existing example apps still build and run

The agent reads `FRICTION.md` directly; the human does not pre-filter or pick entries.

### Step 3 — Re-verify & check loop exit
Invoke `go-web-dev` again with: *"Re-run the Phase 1 task on the updated framework. Read `FRICTION.md` — for each entry svelgo-dev marked `[RESOLVED]`, confirm the fix actually lands in your experience. If the original friction is gone, append a short confirmation note to the entry. If it persists or a new problem appeared, add a new `[OPEN]` entry."* Expected output:
- Confirmation notes appended to resolved entries, or
- New `[OPEN]` entries for regressions / missed fixes

**Loop exit condition:** After Step 3, if `FRICTION.md` has **zero `[OPEN]` entries**, the loop stops — Phase 1 is complete. If any `[OPEN]` entries remain (including new ones raised in Step 3), return to Step 2 and run another round.

**Hard iteration cap: 3 full rounds of Steps 1→2→3.** If FRICTION.md still has Open entries after round 3, the loop stops anyway and the remaining entries are reclassified as Phase 2 backlog. A loop that cannot converge in 3 rounds on a task this trivial is itself a signal — about the framework, the docs, or the loop mechanism — and that signal is more valuable than grinding through a fourth round.

**Phase 1 exits successfully when:** the mandatory baseline app runs end-to-end with Button + Label cross-wired, *and* `FRICTION.md` has no remaining Open entries after at least one full loop pass (or, alternatively, the 3-round cap is reached with remaining entries explicitly deferred), *and* `go-web-dev`'s quality standards audit exists with every FAIL item resolved or explicitly deferred in `FRICTION.md`.

## What Phase 1 intentionally does NOT include

- No new built-in components (no Input, no Form, no Container)
- No routing, auth, middleware, or multi-page
- No new proto messages
- No production build verification (dev mode only)
- No automated test infrastructure
- No automatic agent-to-agent handoff — the human is the router

These are deferred to Phase 2+, after we know the loop works.

## Critical files for Phase 1

For `svelgo-dev` in Step 0, editable scope is `doc/` only — no framework code changes. The tutorial must be grounded in the current state of `*.go` at repo root, `component/`, `proto/ui.proto`, and `frontend/src/runtime/`.

Read-only for `go-web-dev` in Step 1 except where noted:
- `doc/` — the tutorial written in Step 0; **the primary source of truth for app development**
- `CLAUDE.md` — secondary reference only if `doc/` is silent on something
- `FRICTION.md` — shared log, append-only during Step 1
- `demo/clickcounter/` — the directory go-web-dev creates from scratch; must not pre-exist
- `example/` — **do not copy or modify**; may be referenced to verify framework structure if doc is silent, but must not be the source of scaffolding

For `svelgo-dev` in Step 2, editable scope is the full framework core + `doc/` as defined in its agent file.

## Verification

Phase 1 is verified by the human, not by automated tests:

1. Confirm `doc/` contains a tutorial that a Go developer can actually follow — skim it as a human sanity check.
2. Run `make dev`, open each small app go-web-dev built, and confirm it behaves as the tutorial claims.
3. Open `FRICTION.md` and confirm **zero `[OPEN]` entries remain** — every entry raised during the loop has been resolved (framework fix, doc update, or written justification) and confirmed by go-web-dev in Step 3.
4. Open `demo/clickcounter/quality-audit.md` and confirm every standard in `doc/quality-standards.md` has been evaluated. Every FAIL must have a corresponding resolved (or deferred) entry in `FRICTION.md`.
5. Run `make build` once at the end to confirm the production path still compiles (sanity check, not the focus).
6. If any step in the loop required the human to manually rescue an agent (re-explain the task, fix something the agent couldn't), that is itself a loop defect. Since `FRICTION.md`'s entry format is framework-pain-oriented and doesn't fit loop-mechanism defects cleanly, record meta-friction informally at the top of `FRICTION.md` under a `## Meta` heading (or in session notes) and address it before starting Phase 2.

## Out of scope / open for later phases

- Defining how multiple friction entries get prioritized when the log grows
- Whether `svelgo-dev` should proactively scan `FRICTION.md` at the start of every invocation (currently yes per its agent file, but untested)
- Automation: hooks, scheduled loops, or agent-to-agent invocation
- Expanding the task surface (new components, new pages, real enterprise features)
- **Dev/prod parity testing.** Phase 1 only runs `make dev`; the `SVELGO_DEV=1` vs embedded-static divergence goes untested even though both agents carry a principle about it. Phase 2 should include a production-build pass as part of the loop, otherwise the principle stays aspirational.
