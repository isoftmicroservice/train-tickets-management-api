.PHONY: proto build test run-server run-client clean

# Generate protobuf code
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/ticket.proto

# Generate API documentation (HTML) - requires protoc-gen-doc
docs:
	@echo "Generating HTML documentation..."
	@protoc --doc_out=docs --doc_opt=html,index.html api/ticket.proto || echo "Note: protoc-gen-doc not found. See docs/api.md for manual documentation."

# Generate API documentation (Markdown) - requires protoc-gen-doc
docs-md:
	@echo "Generating Markdown documentation..."
	@protoc --doc_out=docs --doc_opt=markdown,api_generated.md api/ticket.proto || echo "Note: protoc-gen-doc not found. See docs/api.md for manual documentation."

# Build server and client
build:
	go build -o bin/server ./cmd/server
	go build -o bin/client ./cmd/client

# Run tests
test:
	go test ./...

# Run server
run-server:
	go run ./cmd/server

# Run client
run-client:
	go run ./cmd/client

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf internal/proto/
	rm -rf docs/*.html docs/*.md





