.PHONY: proto dev build clean

# Regenerate Go and JS protobuf artifacts
proto:
	PATH="$$PATH:$$HOME/go/bin" protoc \
		--go_out=gen \
		--go_opt=paths=source_relative \
		--proto_path=proto \
		proto/ui.proto
	mv gen/ui.pb.go gen/ui/ui.pb.go
	cd frontend && ./node_modules/.bin/pbjs \
		-t json ../proto/ui.proto \
		-o src/ui_descriptor.json

# Development: Go server + Vite dev server
# Visit http://localhost:8080
dev:
	cd frontend && npm run dev &
	SVELGO_DEV=1 go run ./cmd/app/

# Production build: bundle frontend, then compile Go binary
build:
	cd frontend && npm run build
	go build -o dist/svelgo-app ./cmd/app/

clean:
	rm -rf dist/ static/assets static/.vite
