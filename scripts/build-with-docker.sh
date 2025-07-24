#!/bin/bash

# Build Go binaries using Docker without requiring Go installation

set -e

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Parse arguments
TARGET="${1}"

# Show usage
if [ -z "${TARGET}" ]; then
    echo "Usage: $0 <target>"
    echo ""
    echo "Available targets:"
    echo "  module-manager  - Build module-manager static binaries"
    echo "  api-gateway    - Build api-gateway static binaries"
    echo "  drs-cli        - Build drs-cli static binaries"
    echo "  all            - Build all static binaries"
    exit 1
fi

# Build function
build_binaries() {
    local target_name="$1"
    local target_dir="$2"
    local build_cmd="$3"
    
    echo "Building ${target_name} static binaries for AMD64 and ARM64..."
    
    # Build using Docker
    docker run --rm \
        -v "${PROJECT_ROOT}:/workspace" \
        -w /workspace \
        golang:1.24-alpine \
        sh -c "
            echo 'Installing dependencies...' && \
            apk add --no-cache bash protobuf protobuf-dev git make && \
            go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
            go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
            
            echo 'Generating proto files...' && \
            cd /workspace && \
            chmod +x scripts/generate-proto.sh && \
            ./scripts/generate-proto.sh --go-only && \
            
            cd ${target_dir} && \
            go mod download && \
            
            echo 'Building ${target_name}-amd64-static...' && \
            CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -a -ldflags '-w -s -extldflags \"-static\"' -o bin/${target_name}-amd64-static ${build_cmd} && \
            
            echo 'Building ${target_name}-arm64-static...' && \
            CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -buildvcs=false -a -ldflags '-w -s -extldflags \"-static\"' -o bin/${target_name}-arm64-static ${build_cmd} && \
            
            echo 'Build complete!' && \
            
            echo 'Setting file ownership...' && \
            chown -R $(id -u):$(id -g) bin/
        "
    
    echo ""
    echo "Static binaries built successfully:"
    ls -la "${PROJECT_ROOT}/${target_dir}/bin/"*-static
}

# Execute based on target
case "${TARGET}" in
    module-manager)
        build_binaries "module-manager" "services/module-manager" "cmd/server/main.go"
        ;;
    api-gateway)
        build_binaries "api-gateway" "services/api-gateway" "cmd/server/main.go"
        ;;
    drs-cli)
        build_binaries "drs-cli" "tools/drs-cli" "."
        ;;
    all)
        build_binaries "module-manager" "services/module-manager" "cmd/server/main.go"
        echo ""
        build_binaries "api-gateway" "services/api-gateway" "cmd/server/main.go"
        echo ""
        build_binaries "drs-cli" "tools/drs-cli" "."
        ;;
    *)
        echo "Error: Unknown target '${TARGET}'"
        exit 1
        ;;
esac