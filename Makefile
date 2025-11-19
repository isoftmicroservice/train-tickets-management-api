.PHONY: proto build test run-server run-client clean

# Generate protobuf code
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/ticket.proto

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





