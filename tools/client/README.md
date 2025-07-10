# Module Manager Client

A minimal client for testing the module manager gRPC service APIs.

## Build

```bash
cd tools/client
export PATH=/usr/local/go/bin:$PATH

# For local architecture
make build

# For ARM64
make build-arm64

# Both
make all

# Or direct commands
mkdir -p bin && go build -o bin/client main.go                              # Local
mkdir -p bin && GOOS=linux GOARCH=arm64 go build -o bin/client-arm64 main.go  # ARM64
```

## Usage

### Basic Usage
```bash
# reboot (default)
./bin/client

# Execute specific command
./bin/client -cmd=<command>

# Specify different server
./bin/client -server=192.168.1.100:50051 -cmd=disk
```

### Available Commands

#### System Control
```bash
# Immediate reboot
./bin/client -cmd=reboot

# Reboot with 5 second delay
./bin/client -cmd=reboot -delay=5

# Immediate shutdown
./bin/client -cmd=shutdown

# Shutdown with 10 second delay
./bin/client -cmd=shutdown -delay=10
```

#### Service Management
```bash
# List all services
./bin/client -cmd=list-services

# Get service status (resource name format)
./bin/client -cmd=service -service=services/drs_sensor -action=status

# Start/stop/restart services
./bin/client -cmd=service -service=services/drs_recorder -action=start
./bin/client -cmd=service -service=services/drs_recorder -action=stop
./bin/client -cmd=service -service=services/drs_recorder -action=restart

# Enable/disable service auto-start
./bin/client -cmd=service -service=services/drs_sensor -action=enable
./bin/client -cmd=service -service=services/drs_sensor -action=disable
```

#### Legacy Service Management (Backward Compatibility)
```bash
# Stop DRS service
./bin/client -cmd=drs-stop

# Restart DRS service
./bin/client -cmd=drs-restart

# Check DRS service status
./bin/client -cmd=drs-status

# Stop Recorder service
./bin/client -cmd=recorder-stop

# Restart Recorder service
./bin/client -cmd=recorder-restart

# Check Recorder service status
./bin/client -cmd=recorder-status
```

#### Disk Usage Check
```bash
# Root disk usage
./bin/client -cmd=disk
```

#### PTP Time Synchronization Check
```bash
# Check local PTP sync status only
./bin/client -cmd=ptp

# Check PTP sync status for local and remote devices
./bin/client -cmd=ptp-all
```

## Options

- `-server`: Server address (default: localhost:50051)
- `-cmd`: Command to execute (default: reboot)
  - System control: `reboot`, `shutdown`
  - Service management: `service`, `list-services`
  - Legacy service shortcuts: `drs-stop`, `drs-restart`, `drs-status`, `recorder-stop`, `recorder-restart`, `recorder-status`
  - Disk: `disk`
  - PTP sync: `ptp` (local only), `ptp-all` (local + remote)
- `-delay`: Delay seconds before reboot/shutdown (default: 0)
- `-service`: Service name for service management (resource name format: services/{service_id})
- `-action`: Service action (start, stop, restart, status, enable, disable)

## Usage Examples

```bash
# Test server in ARM64 environment
./bin/client-arm64 -server=192.168.20.1:50051 -cmd=disk

# Test service management
./bin/client -server=localhost:50051 -cmd=service -service=services/drs_sensor -action=status

# Test all features on full version server
./bin/client -server=localhost:50051 -cmd=list-services

# Check PTP sync status
./bin/client -cmd=ptp
./bin/client -server=192.168.1.100:50051 -cmd=ptp-all
```

## Notes

- reboot/shutdown commands actually operate the system
- Use only in test environment
- Service management requires proper configuration in the service
- PTP sync check requires `pmc` command
- Remote device PTP check requires ARP table entries