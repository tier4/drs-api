.PHONY: build build-lite build-static build-lite-static build-arm64 build-lite-arm64 build-arm64-static build-lite-arm64-static clean all all-static

# Build full version (local architecture)
build:
	go build -o bin/system-manager cmd/server/main.go

# Build lite version with build tag (local architecture)
build-lite:
	go build -tags lite -o bin/system-manager-lite cmd/server/main.go

# Build full version (static linking for x64)
build-static:
	CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o bin/system-manager-static cmd/server/main.go

# Build lite version (static linking for x64)
build-lite-static:
	CGO_ENABLED=0 go build -tags lite -a -ldflags '-extldflags "-static"' -o bin/system-manager-lite-static cmd/server/main.go

# Build full version for ARM64
build-arm64:
	GOOS=linux GOARCH=arm64 go build -o bin/system-manager-arm64 cmd/server/main.go

# Build lite version for ARM64
build-lite-arm64:
	GOOS=linux GOARCH=arm64 go build -tags lite -o bin/system-manager-lite-arm64 cmd/server/main.go

# Build full version for ARM64 (static linking for embedded systems)
build-arm64-static:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -ldflags '-extldflags "-static"' -o bin/system-manager-arm64-static cmd/server/main.go

# Build lite version for ARM64 (static linking for embedded systems)
build-lite-arm64-static:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lite -a -ldflags '-extldflags "-static"' -o bin/system-manager-lite-arm64-static cmd/server/main.go

# Build all versions
all: build build-lite build-arm64 build-lite-arm64

# Build all versions including static
all-static: build build-lite build-static build-lite-static build-arm64 build-lite-arm64 build-arm64-static build-lite-arm64-static

clean:
	rm -rf bin/

test:
	go test ./...

test-lite:
	go test -tags lite ./...