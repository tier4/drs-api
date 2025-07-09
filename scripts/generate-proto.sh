#!/bin/bash

# Script to generate gRPC code from proto files

set -e  # Exit on error

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROTO_DIR="${SCRIPT_DIR}/../proto"
SERVICES_DIR="${SCRIPT_DIR}/../services"

# Check if required tools are installed
check_tools() {
    local missing_tools=()
    
    if ! command -v protoc &> /dev/null; then
        missing_tools+=("protoc")
    fi
    
    if ! command -v protoc-gen-go &> /dev/null; then
        missing_tools+=("protoc-gen-go")
    fi
    
    if ! command -v protoc-gen-go-grpc &> /dev/null; then
        missing_tools+=("protoc-gen-go-grpc")
    fi
    
    if ! command -v grpc_cpp_plugin &> /dev/null; then
        missing_tools+=("grpc_cpp_plugin")
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        echo "Error: Missing required tools: ${missing_tools[*]}"
        echo "Please run: ./install-protoc-plugins.sh"
        exit 1
    fi
}

# Create output directories
mkdir -p ${SERVICES_DIR}/module-agent/gen/system/v1
mkdir -p ${SERVICES_DIR}/ros2-bridge/gen/ros2bridge/v1

check_tools

echo "Generating Go code for system service..."
protoc -I ${PROTO_DIR} \
    --go_out=${SERVICES_DIR}/module-agent/gen \
    --go_opt=module=github.com/drs-api/services/module-agent/gen \
    --go-grpc_out=${SERVICES_DIR}/module-agent/gen \
    --go-grpc_opt=module=github.com/drs-api/services/module-agent/gen \
    ${PROTO_DIR}/system/v1/system.proto

echo "Generating C++ code for ROS2 bridge service..."
protoc -I ${PROTO_DIR} \
    --cpp_out=${SERVICES_DIR}/ros2-bridge/gen \
    --grpc_out=${SERVICES_DIR}/ros2-bridge/gen \
    --plugin=protoc-gen-grpc=$(which grpc_cpp_plugin) \
    ${PROTO_DIR}/ros2bridge/v1/bridge.proto

echo "Proto generation complete!"
echo ""
echo "Generated files:"
echo "  - Go: ${SERVICES_DIR}/module-agent/gen/"
echo "  - C++: ${SERVICES_DIR}/ros2-bridge/gen/"