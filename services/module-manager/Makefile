.PHONY: build build-static build-arm64 build-arm64-static clean all all-static test

# Build version (local architecture)
build:
	go build -o bin/module-manager cmd/server/main.go

# Build version (static linking for x64)
build-static:
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o bin/module-manager-static cmd/server/main.go

# Build version for ARM64
build-arm64:
	GOOS=linux GOARCH=arm64 go build -o bin/module-manager-arm64 cmd/server/main.go

# Build version for ARM64 (static linking for embedded systems)
build-arm64-static:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -ldflags '-extldflags "-static"' -o bin/module-manager-arm64-static cmd/server/main.go

# Build all versions
all: build build-arm64

# Build all versions including static
all-static: build build-static build-arm64 build-arm64-static

clean:
	rm -rf bin/

test:
	go test ./...
