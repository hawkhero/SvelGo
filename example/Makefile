.PHONY: proto dev build clean

# No app-specific proto messages — all built-ins are in the framework.
# Add app.proto here if you need custom component state messages.
proto:
	@echo "No app-specific proto messages to generate."

# Development: Go server + Vite dev server
# Visit http://localhost:8080
dev:
	cd frontend && npm run dev &
	SVELGO_DEV=1 go run .

# Production build: bundle frontend, then compile Go binary
build:
	cd frontend && npm run build
	go build -o dist/buttonapp .

clean:
	rm -rf dist/ static/assets static/.vite
