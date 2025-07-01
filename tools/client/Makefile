.PHONY: build build-static build-arm64 build-arm64-static clean all all-static

# Build for local architecture
build:
	mkdir -p bin
	go build -o bin/client main.go

# Build for local architecture (static linking)
build-static:
	mkdir -p bin
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o bin/client-static main.go

# Build for ARM64
build-arm64:
	mkdir -p bin
	GOOS=linux GOARCH=arm64 go build -o bin/client-arm64 main.go

# Build for ARM64 (static linking for embedded systems)
build-arm64-static:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -ldflags '-extldflags "-static"' -o bin/client-arm64-static main.go

# Build all versions
all: build build-arm64

# Build all including static
all-static: build build-static build-arm64 build-arm64-static

clean:
	rm -rf bin/

# Show file information
info:
	@echo "Built binaries:"
	@ls -la bin/* 2>/dev/null || echo "No binaries found"
	@echo ""
	@echo "Architecture info:"
	@file bin/* 2>/dev/null || echo "No binaries to check"