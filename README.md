# DRS API - gRPC Microservices

This project provides gRPC-based microservice APIs for system management and control in the DRS (Data Recording System).

## Service Architecture

### 1. Module Manager Service
- **Language**: Go
- **Port**: 50051
- **Package**: `drs.module.v1`
- **Functions**:
  - System control (reboot/shutdown)
  - Service management (start/stop/restart/enable/disable)
  - Disk usage monitoring
  - PTP (Precision Time Protocol) synchronization status
  - Configurable API endpoints per module requirements

### 2. ROS2 Bridge Service
- **Language**: C++
- **Port**: 50052
- **Functions**:
  - Subscribe to ROS2 topics and retrieve values via gRPC
  - Call ROS2 services via gRPC

## Directory Structure

```
drs-api/
├── proto/                     # gRPC definition files
│   └── drs/
│       └── module/
│           └── v1/
│               └── module.proto
├── services/                  # Microservices
│   ├── module-manager/        # System management service
│   └── ros2-bridge/          # ROS2 bridge service
├── tools/                    # Client tools
│   └── client/              # Test client for module-manager
├── scripts/                  # Build and utility scripts
└── docs/                     # Documentation
```

## Setup

### 1. Required Tools Installation

```bash
# Install protobuf compiler
sudo apt-get install -y protobuf-compiler

# Install gRPC C++ plugin (for ROS2 bridge)
sudo apt-get install -y libgrpc++-dev protobuf-compiler-grpc

# Install Go protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Notes
- Module Manager service requires **native execution** (for shutdown/reboot/systemd operations)
- For ARM64 deployment, use cross-compiled binaries
- For older glibc environments, use static linked versions (`*-static`)

### 2. Generate Code from Proto Definitions

```bash
# From project root
cd drs-api

# Generate protobuf code for Module Manager
protoc --go_out=services/module-manager --go-grpc_out=services/module-manager proto/drs/module/v1/module.proto
```

## Build and Run

### Module Manager (Go)
Static linked versions work on older glibc environments.

```bash
cd services/module-manager

# Build for local architecture
make build

# Build static version
make build-static

# Build for ARM64
make build-arm64

# Build ARM64 static version
make build-arm64-static

# Build all versions
make all-static

# Run the service
./bin/module-manager -port=50051

# Run with custom config
./bin/module-manager -config=custom-config.yaml

# Run on ARM64
./bin/module-manager-arm64 -port=50051
```

### Configuration
The Module Manager service is highly configurable. Each API can be individually enabled/disabled based on module requirements:

```yaml
server:
  port: 50051

# Disk monitoring settings
disk:
  enabled: true
  monitor_path: "/"

# Service management settings  
services:
  enabled: true
  services:
    drs_sensor:
      systemd_name: "drs-sensor.service"
      description: "DRS Sensor Service"

# System settings
system:
  enable_reboot: true
  enable_shutdown: true

# PTP sync settings
ptp:
  enabled: true
  remote_devices:
    - name: "sensor1"
      address: "192.168.1.101:50051"
```

See `services/module-manager/README.md` for detailed configuration options.

### ROS2 Bridge (C++)
```bash
cd services/ros2-bridge
colcon build
source install/setup.bash
ros2 run ros2_bridge ros2_bridge_node
```

## Client Tools

### Module Manager Test Client
```bash
cd tools/client
make build

# Test system APIs
./bin/client -cmd=reboot -delay=60
./bin/client -cmd=shutdown

# Test service management
./bin/client -cmd=list-services
./bin/client -cmd=service -service=services/drs_sensor -action=status

# Test disk usage
./bin/client -cmd=disk

# Test PTP sync
./bin/client -cmd=ptp
./bin/client -cmd=ptp-all
```

## API Specifications

### Module Manager API
The Module Manager service follows Cloud API Design Guide principles:
- Resource-oriented design with standard methods
- Fire-and-forget pattern for system operations (reboot/shutdown)
- Direct resource returns without success/message wrappers

API details can be found in `proto/drs/module/v1/module.proto`.

### ROS2 Bridge API
API specifications for ROS2 Bridge are in the respective proto files.

## Development

### API Design Principles
- Follow [Google Cloud API Design Guide](https://cloud.google.com/apis/design)
- Use resource-oriented design
- Implement standard methods (Get, List, Create, Update, Delete) where applicable
- Use proper error handling with gRPC status codes

### Module-Specific Deployments
The Module Manager can be configured differently for each module type:
- **Sensing Module**: Enable all APIs including PTP sync monitoring
- **Storage Module**: Focus on disk monitoring, disable unnecessary APIs
- **Control Module**: Enable service management for system-wide control

See configuration examples in `services/module-manager/README.md`.