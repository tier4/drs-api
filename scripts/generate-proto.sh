#!/bin/bash

# Script to generate gRPC code from proto files

set -e  # Exit on error

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROTO_DIR="${SCRIPT_DIR}/../proto"
SERVICES_DIR="${SCRIPT_DIR}/../services"

# Parse command line arguments
GENERATE_GO=true
GENERATE_CPP=true

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Generate gRPC code from proto files"
    echo ""
    echo "Options:"
    echo "  --cpp-only      Generate only C++ code (for ros2-bridge)"
    echo "  --go-only       Generate only Go code (for module-manager)"
    echo "  --help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0              # Generate both Go and C++ code"
    echo "  $0 --cpp-only   # Generate only C++ code (no Go tools required)"
    echo "  $0 --go-only    # Generate only Go code"
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --cpp-only)
            GENERATE_GO=false
            GENERATE_CPP=true
            shift
            ;;
        --go-only)
            GENERATE_GO=true
            GENERATE_CPP=false
            shift
            ;;
        --help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Check if required tools are installed
check_tools() {
    local missing_tools=()
    
    # Always required
    if ! command -v protoc &> /dev/null; then
        missing_tools+=("protoc")
    fi
    
    # Go tools (only required if generating Go code)
    if [[ "$GENERATE_GO" == "true" ]]; then
        if ! command -v protoc-gen-go &> /dev/null; then
            missing_tools+=("protoc-gen-go")
        fi
        
        if ! command -v protoc-gen-go-grpc &> /dev/null; then
            missing_tools+=("protoc-gen-go-grpc")
        fi
    fi
    
    # C++ tools (only required if generating C++ code)
    if [[ "$GENERATE_CPP" == "true" ]]; then
        if ! command -v grpc_cpp_plugin &> /dev/null; then
            missing_tools+=("grpc_cpp_plugin")
        fi
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        echo "Error: Missing required tools: ${missing_tools[*]}"
        if [[ "$GENERATE_GO" == "true" && "$GENERATE_CPP" == "true" ]]; then
            echo "Please run: ./install-protoc-plugins.sh"
        elif [[ "$GENERATE_GO" == "true" ]]; then
            echo "For Go code generation, you need Go protoc plugins."
            echo "Install with: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
            echo "             go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
        elif [[ "$GENERATE_CPP" == "true" ]]; then
            echo "For C++ code generation, you need gRPC C++ plugin."
            echo "Install with: sudo apt-get install libgrpc++-dev protobuf-compiler-grpc"
        fi
        exit 1
    fi
}

# Create output directories based on what we're generating
if [[ "$GENERATE_GO" == "true" ]]; then
    mkdir -p ${SERVICES_DIR}/module-manager/gen/drs/module/v1
    # Also create directories for API Gateway
    mkdir -p ${SERVICES_DIR}/api-gateway/gen/drs/module/v1
    mkdir -p ${SERVICES_DIR}/api-gateway/gen/drs/ros2bridge/v1
fi

if [[ "$GENERATE_CPP" == "true" ]]; then
    mkdir -p ${SERVICES_DIR}/ros2-bridge/gen/drs/ros2bridge/v1
fi

check_tools

# Generate Go code if requested
if [[ "$GENERATE_GO" == "true" ]]; then
    echo "Generating Go code for module manager service..."
    protoc -I ${PROTO_DIR} \
        --go_out=${SERVICES_DIR}/module-manager/gen \
        --go_opt=paths=source_relative \
        --go-grpc_out=${SERVICES_DIR}/module-manager/gen \
        --go-grpc_opt=paths=source_relative \
        ${PROTO_DIR}/drs/module/v1/*.proto
    
    echo "Generating Go code for API gateway..."
    # Generate module proto files for API Gateway
    protoc -I ${PROTO_DIR} \
        --go_out=${SERVICES_DIR}/api-gateway/gen \
        --go_opt=paths=source_relative \
        --go-grpc_out=${SERVICES_DIR}/api-gateway/gen \
        --go-grpc_opt=paths=source_relative \
        ${PROTO_DIR}/drs/module/v1/*.proto
    
    # Generate ROS2 bridge proto files for API Gateway
    protoc -I ${PROTO_DIR} \
        --go_out=${SERVICES_DIR}/api-gateway/gen \
        --go_opt=paths=source_relative \
        --go-grpc_out=${SERVICES_DIR}/api-gateway/gen \
        --go-grpc_opt=paths=source_relative \
        ${PROTO_DIR}/drs/ros2bridge/v1/*.proto
fi

# Generate C++ code if requested
if [[ "$GENERATE_CPP" == "true" ]]; then
    echo "Generating C++ code for ROS2 bridge service..."
    protoc -I ${PROTO_DIR} \
        --cpp_out=${SERVICES_DIR}/ros2-bridge/gen \
        --grpc_out=${SERVICES_DIR}/ros2-bridge/gen \
        --plugin=protoc-gen-grpc=$(which grpc_cpp_plugin) \
        ${PROTO_DIR}/drs/ros2bridge/v1/*.proto
fi

echo "Proto generation complete!"
echo ""
echo "Generated files:"
if [[ "$GENERATE_GO" == "true" ]]; then
    echo "  - Go (module-manager): ${SERVICES_DIR}/module-manager/gen/"
    echo "  - Go (api-gateway): ${SERVICES_DIR}/api-gateway/gen/"
fi
if [[ "$GENERATE_CPP" == "true" ]]; then
    echo "  - C++ (ros2-bridge): ${SERVICES_DIR}/ros2-bridge/gen/"
fi