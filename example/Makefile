.PHONY: proto dev build clean

# Regenerate app-specific Go protobuf types and JS descriptor
proto:
	PATH="$$PATH:$$HOME/go/bin" protoc \
		--go_out=gen \
		--go_opt=paths=source_relative \
		--proto_path=proto \
		proto/app.proto
	mv gen/app.pb.go gen/app/app.pb.go
	../../frontend/node_modules/.bin/pbjs \
		-t json proto/app.proto \
		-o frontend/src/app_descriptor.json

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
