# Module Manager Service

A gRPC service for system management that runs on each DRS module (Sensing Module, Storage Module, etc.).
APIs can be individually enabled/disabled via configuration file, allowing customization per module requirements.

## Service Architecture

The Module Manager now implements a service-oriented architecture with multiple gRPC services running on a single endpoint:

### Available Services

- **ServiceManagerService**: systemd service management (start/stop/restart/enable/disable)
- **SystemControlService**: system-level operations (reboot/shutdown)
- **MonitoringService**: resource monitoring (disk usage, PTP synchronization)

All services run on the same gRPC server and port (50051), allowing clients to use any service based on their needs.

## Features

### SystemControlService APIs

- **Reboot**: System restart (configurable delay)
- **Shutdown**: System shutdown (configurable delay)

### ServiceManagerService APIs

- **GetService**: Retrieve information for a specific service
- **ListServices**: List configured services
- **StartService**: Start a service
- **StopService**: Stop a service
- **RestartService**: Restart a service
- **EnableService**: Enable service auto-start
- **DisableService**: Disable service auto-start

### MonitoringService APIs

- **GetDiskUsage**: Get disk usage information (single path)
- **GetPTPStatus**: Get PTP (Precision Time Protocol) status information

### Individual API Control

Each API can be individually enabled/disabled through configuration:

- **MonitoringService**: disk.enabled and ptp.enabled settings control individual APIs
- **SystemControlService**: system.enable_reboot and system.enable_shutdown settings
- **ServiceManagerService**: services.enabled setting
- When disabled, APIs return `Unimplemented` error with descriptive message

## Configuration File

### Configuration File Location

Configuration files are searched in the following order:

1. `./config.yaml`
2. `./config.yml`
3. `<executable_directory>/config.yaml`
4. `<executable_directory>/config.yml`
5. `/etc/module-manager/config.yaml`
6. `/etc/module-manager/config.yml`

### Configuration Example (config.yaml)

```yaml
server:
  port: 50051

# Disk monitoring settings (single path per ECU)
disk:
  enabled: true                 # Enable disk usage API
  monitor_path: "/"             # Path to monitor

# Service management settings
services:
  enabled: true                 # Enable service management API
  services:
    drs_sensor:
      systemd_name: "drs-sensor.service"
      description: "DRS Sensor Service"
    drs_recorder:
      systemd_name: "drs-recorder.service"
      description: "DRS Recorder Service"

# System settings
system:
  enable_reboot: true           # Enable reboot API
  enable_shutdown: true         # Enable shutdown API
  max_delay_seconds: 300
  allow_reboot: true
  allow_shutdown: true

# PTP sync settings
ptp:
  enabled: true                 # Enable PTP sync check API
  remote_devices:               # Remote device PTP sync check
    - name: "sensor1"
      address: "192.168.1.101:50051"
    - name: "sensor2"
      address: "192.168.1.102:50051"
```

## Usage

### Run with default configuration

```bash
./bin/module-manager
```

### Run with custom configuration file

```bash
./bin/module-manager -config=custom-config.yaml
```

### Run with specific port (overrides configuration file)

```bash
./bin/module-manager -port=50052
```

## API Usage Examples

With the new service architecture, you can use gRPC clients to call specific services directly:

```bash
# Example using grpcurl for MonitoringService
grpcurl -plaintext localhost:50051 drs.module.v1.MonitoringService/GetDiskUsage

# Example using grpcurl for SystemControlService
grpcurl -plaintext -d '{"delay_seconds": 60}' localhost:50051 drs.module.v1.SystemControlService/Reboot

# Example using grpcurl for ServiceManagerService
grpcurl -plaintext localhost:50051 drs.module.v1.ServiceManagerService/ListServices
```

## Module-Specific Configuration Examples

### Sensing Module Configuration

```yaml
server:
  port: 50051

disk:
  enabled: true                    # Enable disk usage API
  monitor_path: "/data"            # Sensor data storage area

services:
  enabled: true                    # Enable service management API
  services:
    drs_sensor:
      systemd_name: "drs-sensor.service"
      description: "DRS Sensor Service"
    drs_recorder:
      systemd_name: "drs-recorder.service"
      description: "DRS Recorder Service"

system:
  enable_reboot: true              # Enable reboot API
  enable_shutdown: true            # Enable shutdown API
  max_delay_seconds: 300
  allow_reboot: true
  allow_shutdown: true

ptp:
  enabled: true                    # Enable PTP sync check API
```

### Storage Module Configuration

```yaml
server:
  port: 50051

disk:
  enabled: true                    # Enable disk usage API (main feature)
  monitor_path: "/storage"         # Data storage area

services:
  enabled: false                   # Disable service management API (not needed for storage module)

system:
  enable_reboot: true              # Enable reboot API
  enable_shutdown: true            # Enable shutdown API
  max_delay_seconds: 300
  allow_reboot: true
  allow_shutdown: true

ptp:
  enabled: false                   # Disable PTP sync check API (not needed for storage module)
```

### Control Module Configuration

```yaml
server:
  port: 50051

disk:
  enabled: true                    # Enable disk usage API
  monitor_path: "/data"

services:
  enabled: true                    # Enable service management API (for overall system management)
  services:
    drs_control:
      systemd_name: "drs-control.service"
      description: "DRS Control Service"

system:
  enable_reboot: true              # Enable reboot API
  enable_shutdown: true            # Enable shutdown API
  max_delay_seconds: 300
  allow_reboot: true
  allow_shutdown: true

ptp:
  enabled: true                    # Enable PTP sync check API (for sensor module sync check)
  remote_devices:                  # Monitor sensor module sync status
    - name: "sensor1"
      address: "192.168.1.101:50051"
    - name: "sensor2"
      address: "192.168.1.102:50051"
```

## Build

```bash
# Standard build
make build

# Static linked build (for older glibc environments)
make build-static

# ARM64 build
make build-arm64

# ARM64 static linked build
make build-arm64-static
```
