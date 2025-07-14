# DRS Proto Generation Scripts

This directory contains scripts for generating gRPC code from protobuf definitions.

## generate-proto.sh

Generates gRPC code for both Go (module-manager) and C++ (ros2-bridge) services.

### Usage

```bash
# Generate both Go and C++ code (default)
./scripts/generate-proto.sh

# Generate only C++ code (useful when Go is not available)
./scripts/generate-proto.sh --cpp-only

# Generate only Go code
./scripts/generate-proto.sh --go-only

# Show help
./scripts/generate-proto.sh --help
```

### Requirements

**For C++ code generation (ros2-bridge):**
- `protoc` (Protocol Buffers compiler)
- `grpc_cpp_plugin` (gRPC C++ plugin)

Install on Ubuntu:
```bash
sudo apt-get install protobuf-compiler libgrpc++-dev protobuf-compiler-grpc
```

**For Go code generation (module-manager):**
- `protoc` (Protocol Buffers compiler)
- `protoc-gen-go` (Go protobuf plugin)
- `protoc-gen-go-grpc` (Go gRPC plugin)

Install Go plugins:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Examples

**Building ros2-bridge without Go:**
```bash
# Only generate C++ code (no Go tools required)
./scripts/generate-proto.sh --cpp-only

# Build ros2-bridge
cd services/ros2-bridge
source /path/to/proto_recorder/install/setup.bash
colcon build
```

**Building module-manager without C++ tools:**
```bash
# Only generate Go code (no C++ tools required)
./scripts/generate-proto.sh --go-only

# Build module-manager
cd services/module-manager
make build
```

### Generated Files

- **Go files**: `services/module-manager/drs/module/v1/*.pb.go`
- **C++ files**: `services/ros2-bridge/gen/drs/ros2bridge/v1/*{.pb.h,.pb.cc,.grpc.pb.h,.grpc.pb.cc}`

### Troubleshooting

If you get "missing tools" errors:

1. **For `--cpp-only` mode**: Only `protoc` and `grpc_cpp_plugin` are required
2. **For `--go-only` mode**: Only `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` are required
3. **For default mode**: All tools are required

The script will provide specific installation instructions based on the mode you're using.