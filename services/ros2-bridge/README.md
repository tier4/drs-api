# ROS2 Bridge Service

Bridge service that combines ROS2 nodes with a gRPC server.

## Prerequisites

### ROS2 Environment
```bash
# Ensure ROS2 Humble is installed
source /opt/ros/humble/setup.bash
```

### gRPC C++ Libraries
```bash
sudo apt-get install -y libgrpc++-dev protobuf-compiler-grpc
```

## Build Steps

### 1. Generate C++ Code from Proto Definitions
```bash
cd ../../scripts
./generate-proto.sh  # Generate both Go and C++ code
```

### 2. Build as a ROS2 Package
```bash
cd ../services/ros2-bridge

# Set up the ROS2 environment
source /opt/ros/humble/setup.bash

# Build with colcon
colcon build --packages-select ros2_bridge

# Source the setup script
source install/setup.bash
```

### 3. Run
```bash
# Run as a ROS2 node
ros2 run ros2_bridge ros2_bridge_node

# Or run directly
./install/ros2_bridge/lib/ros2_bridge/ros2_bridge_node
```

## Configuration

### Change the gRPC Port
```bash
ros2 run ros2_bridge ros2_bridge_node --ros-args -p grpc_port:=50052
```

## API Specification

The gRPC service is defined in `proto/ros2bridge/v1/bridge.proto`.

### Main Features
- Subscribe to ROS2 topics and retrieve values through gRPC
- Call ROS2 services through gRPC
- Stream topics in real time

## Troubleshooting

### 1. gRPC Libraries Not Found
```bash
sudo apt-get install -y libgrpc++-dev
```

### 2. Proto Files Have Not Been Generated
```bash
cd ../../scripts
./generate-proto-go.sh
```

### 3. ROS2 Environment Not Found
```bash
source /opt/ros/humble/setup.bash
```
