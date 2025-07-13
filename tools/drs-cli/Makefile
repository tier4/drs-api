.PHONY: build clean test lint deps install

# Binary name
BINARY_NAME := drs-cli
OUTPUT_DIR := bin

# Build the binary
build:
	go build -o $(OUTPUT_DIR)/$(BINARY_NAME) .

# Build for multiple platforms
build-all: build-linux build-arm64 build-darwin

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64 .

build-arm64:
	GOOS=linux GOARCH=arm64 go build -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-arm64 .

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64 .

# Install dependencies
deps:
	go mod tidy
	go mod download

# Test
test:
	go test -v ./...

# Lint
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf $(OUTPUT_DIR)

# Install to local bin
install: build
	sudo cp $(OUTPUT_DIR)/$(BINARY_NAME) /usr/local/bin/

# Generate proto files (if needed)
proto-gen:
	protoc --proto_path=../../proto \
		--go_out=. \
		--go-grpc_out=. \
		../../proto/drs/module/v1/common.proto \
		../../proto/drs/module/v1/service_manager.proto \
		../../proto/drs/module/v1/system_control.proto \
		../../proto/drs/module/v1/monitoring.proto

# Development help
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  deps       - Install/update dependencies"
	@echo "  test       - Run tests"
	@echo "  lint       - Run linter"
	@echo "  clean      - Clean build artifacts"
	@echo "  install    - Install to /usr/local/bin"
	@echo "  proto-gen  - Regenerate proto files"
	@echo "  help       - Show this help"