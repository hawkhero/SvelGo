.PHONY: proto dev build clean

# No app-specific proto messages — all built-ins are in the framework.
# Add app.proto here if you need custom component state messages.
proto:
	@echo "No app-specific proto messages to generate."

# Development: Go server + Vite dev server
# Visit http://localhost:8080
#
# Strategy: start Go first and wait until it is listening on :8080.
# Only then start the Vite dev server so that startup output is not
# interleaved with a Go panic and so that a Go crash aborts the whole
# target instead of silently failing behind Vite's banner.
dev:
	@echo "[go] Starting Go server..."; \
	SVELGO_DEV=1 go run . & GO_PID=$$!; \
	for i in $$(seq 1 20); do \
		sleep 0.5; \
		curl -sf http://localhost:8080/ > /dev/null 2>&1 && break; \
		if ! kill -0 $$GO_PID 2>/dev/null; then \
			echo "[go] Go server exited unexpectedly — check the output above for errors."; \
			exit 1; \
		fi; \
	done; \
	echo "[vite] Starting Vite dev server..."; \
	cd frontend && npm run dev

# Production build: bundle frontend, then compile Go binary
build:
	cd frontend && npm run build
	go build -o dist/buttonapp .

clean:
	rm -rf dist/ static/assets static/.vite
