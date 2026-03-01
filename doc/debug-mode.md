# Debug Mode (`SVELGO_DEBUG=1`)

SvelGo supports a `SVELGO_DEBUG` environment variable that enables verbose console
logging of the protobuf wire protocol in the browser. This makes it easy to inspect
decoded component state and WebSocket events without a binary protocol analyser.

---

## Usage

```bash
# Development — combine with SVELGO_DEV as needed
SVELGO_DEBUG=1 make dev

# Production binary — works independently of SVELGO_DEV
SVELGO_DEBUG=1 ./dist/svelgo-app
```

`SVELGO_DEBUG` and `SVELGO_DEV` are independent flags:

| Flag | Controls |
|---|---|
| `SVELGO_DEV=1` | Where assets are served from (Vite dev server vs embedded bundle) |
| `SVELGO_DEBUG=1` | Browser console logging of decoded protobuf frames |

You can use either flag alone or both together. In particular, `SVELGO_DEBUG=1` on a
production binary lets you inspect wire traffic without rebuilding or redeploying.

---

## What Gets Logged

Open browser DevTools → Console after starting with `SVELGO_DEBUG=1`:

**On page load** — decoded initial page state:
```
[svelgo init] page state  { pageId: "…", components: [ … ] }
```

**On event send** — outgoing `ClientEvent` frame:
```
[svelgo ws →]  { pageId: "…", componentId: "btn-1", eventType: "click" }
```

**On state update received** — decoded `StateUpdate` frame:
```
[svelgo ws ←]  { pageId: "…", updatedComponents: [ … ] }
```

Without the flag no debug output is produced.

---

## How It Works

When `SVELGO_DEBUG=1` is set, Go injects `window.__SVELGO_DEBUG__ = true` into every
HTML page response (via `template.go`). The JS runtime checks this flag in two places:

- **`client.ts` `bootstrap()`** — logs the decoded `PageState` after the initial
  base64 protobuf blob is parsed.
- **`ws.ts`** — logs outgoing `ClientEvent` objects before sending, and decoded
  `StateUpdate` objects after receiving each WebSocket frame.

The check is a simple boolean guard (`window.__SVELGO_DEBUG__ === true`), so there is
zero overhead when the flag is off.
