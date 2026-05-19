# DRS CLI

A command-line client for interacting with DRS gRPC services. Supports both module-manager service (port 50051) and ros2-bridge service (port 50052).

## Installation

### Build from source

```bash
make build
```

The binary will be created in `bin/drs-cli`.

### Install to system

```bash
make install
```

This will install the binary to `/usr/local/bin/drs-cli`.

## Usage

### Basic Usage

```bash
# Show help
./bin/drs-cli --help

# Show module commands
./bin/drs-cli module --help

# Connect to a different server
./bin/drs-cli --address=192.168.1.100:50051 module services list

# Set request timeout
./bin/drs-cli --timeout=60 module monitoring disk
```

## Commands

### Module Manager Service Commands

All module-manager related commands are under the `module` subcommand:

#### Service Management

```bash
# List all services
./bin/drs-cli module services list

# Get specific service information
./bin/drs-cli module services get services/drs_sensor

# Service control
./bin/drs-cli module services start services/drs_sensor
./bin/drs-cli module services stop services/drs_sensor
./bin/drs-cli module services restart services/drs_sensor

# Service enable/disable
./bin/drs-cli module services enable services/drs_sensor
./bin/drs-cli module services disable services/drs_sensor
```

#### System Control

```bash
# Immediate reboot/shutdown
./bin/drs-cli module system reboot
./bin/drs-cli module system shutdown

# Delayed reboot/shutdown
./bin/drs-cli module system reboot --delay=60
./bin/drs-cli module system shutdown --delay=30
```

#### Monitoring

```bash
# Get disk usage
./bin/drs-cli module monitoring disk

# Get PTP status (local only)
./bin/drs-cli module monitoring ptp

# Get PTP status including remote devices
./bin/drs-cli module monitoring ptp --include-remote
```

### ROS2-Bridge Service Commands

All ros2-bridge related commands are under the `ros2bridge` subcommand and require connecting to port 50052:

#### Sensing Commands

```bash
# Get current position from NavSatFix
./bin/drs-cli --address localhost:50052 ros2bridge sensing position

# List ROS2 nodes
./bin/drs-cli --address localhost:50052 ros2bridge sensing list-nodes

# Filter nodes by namespace
./bin/drs-cli --address localhost:50052 ros2bridge sensing list-nodes --filter "namespace:=/sensing"
```

#### Recording Commands

```bash
# Recording control (affects all ECUs)
./bin/drs-cli --address localhost:50052 ros2bridge recording start
./bin/drs-cli --address localhost:50052 ros2bridge recording stop
./bin/drs-cli --address localhost:50052 ros2bridge recording pause
./bin/drs-cli --address localhost:50052 ros2bridge recording resume

# Get recording status
./bin/drs-cli --address localhost:50052 ros2bridge recording list
./bin/drs-cli --address localhost:50052 ros2bridge recording get main_ecu

# Topic status monitoring
./bin/drs-cli --address localhost:50052 ros2bridge recording list-topics main_ecu
./bin/drs-cli --address localhost:50052 ros2bridge recording list-topics main_ecu --filter "topic_name:=/sensing/*"
```

## Configuration

### Command Line Flags

- `--address`: Server address (default: localhost:50051 for module-manager, use localhost:50052 for ros2-bridge)
- `--timeout`: Request timeout in seconds (default: 30)

### Service Endpoints

- **module-manager**: `localhost:50051` (default)
- **ros2-bridge**: `localhost:50052` (must specify with --address)

### Examples

```bash
# Connect to remote server with extended timeout
./bin/drs-cli --address=192.168.1.100:50051 --timeout=60 module services list

# Quick local monitoring
./bin/drs-cli module monitoring disk
./bin/drs-cli module monitoring ptp -r
```

## Service Name Format

When working with services, use the resource name format: `services/{service_id}`

Examples:

- `services/drs_sensor`
- `services/drs_recorder`
- `services/drs_control`

## Error Handling

The client will display appropriate error messages for common scenarios:

- **Service Disabled**: When a service (like disk monitoring or PTP) is disabled in the server configuration
- **Connection Failed**: When the server is unreachable
- **Invalid Service**: When requesting a non-existent service
- **Permission Denied**: When system operations are not permitted

## Development

### Dependencies

```bash
make deps
```

### Building for Multiple Platforms

```bash
make build-all
```

This creates binaries for:

- Linux AMD64
- Linux ARM64
- macOS AMD64

### Regenerating Proto Files

If the proto definitions change:

```bash
make proto-gen
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

## Examples in Different Scenarios

### Sensing Module Management

```bash
# Check sensor service status
./bin/drs-cli module services get services/drs_sensor

# Monitor disk usage on sensor data partition
./bin/drs-cli module monitoring disk

# Check PTP synchronization with other devices
./bin/drs-cli module monitoring ptp --include-remote
```

### Storage Module Management

```bash
# Check disk usage on storage partition
./bin/drs-cli module monitoring disk

# Restart storage service if needed
./bin/drs-cli module services restart services/drs_storage
```

### System Maintenance

```bash
# Schedule maintenance reboot
./bin/drs-cli module system reboot --delay=300

# Emergency shutdown
./bin/drs-cli module system shutdown
```

## Architecture

The CLI is structured to support multiple DRS services:

```text
drs-cli
├── module          # module-manager service (port 50051)
│   ├── services   # ServiceManagerService
│   ├── system     # SystemControlService
│   └── monitoring # MonitoringService
└── ros2bridge     # ros2-bridge service (port 50052)
    ├── sensing    # SensingService
    │   ├── position
    │   └── list-nodes
    └── recording  # RecordingService
        ├── start/stop/pause/resume
        ├── get/list
        └── list-topics
```

This design ensures clean separation of concerns and easy addition of new services as the DRS ecosystem grows.

## Port Reference

| Service        | Port  | Usage                                                    |
|----------------|-------|----------------------------------------------------------|
| module-manager | 50051 | `./bin/drs-cli module ...` (default)                     |
| ros2-bridge    | 50052 | `./bin/drs-cli --address localhost:50052 ros2bridge ...` |
