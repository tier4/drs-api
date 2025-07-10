# Module Manager Service

A gRPC service for system management that runs on each DRS module (Sensing Module, Storage Module, etc.).
APIs can be individually enabled/disabled via configuration file, allowing customization per module requirements.

## Features

### System Control APIs
- **Reboot**: System restart (configurable delay)
- **Shutdown**: System shutdown (configurable delay)

### Service Management APIs
- **GetService**: Retrieve information for a specific service
- **ListServices**: List configured services
- **StartService**: Start a service
- **StopService**: Stop a service
- **RestartService**: Restart a service
- **EnableService**: Enable service auto-start
- **DisableService**: Disable service auto-start

### Resource Monitoring APIs
- **GetDiskUsage**: Get disk usage information (single path)
- **GetPTPStatus**: Get PTP (Precision Time Protocol) status information

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

### Get Disk Usage
```bash
# Get usage for path defined in configuration file
./tools/client -cmd=disk
```

### Service Management
```bash
# List services
./tools/client -cmd=list-services

# Get service information (resource name format)
./tools/client -cmd=service -name=services/drs_sensor -action=status

# Start/stop/restart services
./tools/client -cmd=service -name=services/drs_recorder -action=start
./tools/client -cmd=service -name=services/drs_recorder -action=stop
./tools/client -cmd=service -name=services/drs_recorder -action=restart

# Enable/disable service auto-start
./tools/client -cmd=service -name=services/drs_sensor -action=enable
./tools/client -cmd=service -name=services/drs_sensor -action=disable
```

### PTP Sync Check
```bash
# Check local PTP sync status
./tools/client -cmd=ptp

# Check all devices (local + remote) PTP sync status
./tools/client -cmd=ptp-all
```

### System Control
```bash
# Immediate reboot
./tools/client -cmd=reboot

# Reboot after 60 seconds
./tools/client -cmd=reboot -delay=60

# Immediate shutdown
./tools/client -cmd=shutdown

# Shutdown after 30 seconds
./tools/client -cmd=shutdown -delay=30
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