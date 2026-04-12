.PHONY: proto example dev build clean

# Regenerate framework Go and JS protobuf artifacts (framework messages only)
proto:
	mkdir -p gen/ui
	PATH="$$PATH:$$HOME/go/bin" protoc \
		--go_out=gen \
		--go_opt=paths=source_relative \
		--proto_path=proto \
		proto/ui.proto
	mv gen/ui.pb.go gen/ui/ui.pb.go
	cd frontend && ./node_modules/.bin/pbjs \
		-t json ../proto/ui.proto \
		-o src/runtime/ui_descriptor.json

# Build the example app (frontend + Go binary)
example:
	$(MAKE) -C example

# Development: run the example app against the Vite dev server
# Visit http://localhost:8080
dev:
	$(MAKE) -C example dev

# Production build of the example app
build:
	$(MAKE) -C example build

clean:
	rm -rf example/dist example/static/assets example/static/.vite
