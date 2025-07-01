#!/bin/bash

echo "Installing Protocol Buffer compiler plugins..."

# Install Go plugins
echo "Installing Go protoc plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Add Go bin to PATH if not already there
if ! echo $PATH | grep -q "$(go env GOPATH)/bin"; then
    echo "Adding Go bin to PATH..."
    echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
    export PATH=$PATH:$(go env GOPATH)/bin
fi

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "protoc is not installed. Please install it first:"
    echo "  Ubuntu/Debian: sudo apt-get install -y protobuf-compiler"
    echo "  macOS: brew install protobuf"
    exit 1
fi

# Check if grpc C++ plugin is installed
if ! command -v grpc_cpp_plugin &> /dev/null; then
    echo "gRPC C++ plugin is not installed. Installing..."
    echo "  Ubuntu/Debian: sudo apt-get install -y libgrpc++-dev protobuf-compiler-grpc"
    echo "  macOS: brew install grpc"
fi

echo "Installation complete!"
echo ""
echo "Installed tools:"
command -v protoc-gen-go &> /dev/null && echo "✓ protoc-gen-go: $(protoc-gen-go --version)"
command -v protoc-gen-go-grpc &> /dev/null && echo "✓ protoc-gen-go-grpc: installed"
command -v grpc_cpp_plugin &> /dev/null && echo "✓ grpc_cpp_plugin: installed"