.PHONY: build build-static build-arm64 build-arm64-static clean all all-static

# Build for local architecture
build:
	go build -o client main.go

# Build for local architecture (static linking)
build-static:
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o client-static main.go

# Build for ARM64
build-arm64:
	GOOS=linux GOARCH=arm64 go build -o client-arm64 main.go

# Build for ARM64 (static linking for embedded systems)
build-arm64-static:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -ldflags '-extldflags "-static"' -o client-arm64-static main.go

# Build all versions
all: build build-arm64

# Build all including static
all-static: build build-static build-arm64 build-arm64-static

clean:
	rm -f client client-static client-arm64 client-arm64-static

# Show file information
info:
	@echo "Built binaries:"
	@ls -la client* 2>/dev/null || echo "No binaries found"
	@echo ""
	@echo "Architecture info:"
	@file client* 2>/dev/null || echo "No binaries to check"