package system

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Reboot(delaySeconds int) error {
	if delaySeconds < 0 {
		return fmt.Errorf("delay cannot be negative")
	}

	var cmd *exec.Cmd
	if delaySeconds == 0 {
		cmd = exec.Command("reboot")
	} else {
		cmd = exec.Command("shutdown", "-r", fmt.Sprintf("+%d", delaySeconds/60))
	}

	return cmd.Run()
}

func (m *Manager) Shutdown(delaySeconds int) error {
	if delaySeconds < 0 {
		return fmt.Errorf("delay cannot be negative")
	}

	var cmd *exec.Cmd
	if delaySeconds == 0 {
		cmd = exec.Command("poweroff")
	} else {
		cmd = exec.Command("shutdown", fmt.Sprintf("+%d", delaySeconds/60))
	}

	return cmd.Run()
}

const (
	DrsServiceName      = "drs.service"
	RecorderServiceName = "recorder.service"
)

// DRS Service Management
func (m *Manager) StopDrsService() error {
	cmd := exec.Command("systemctl", "stop", DrsServiceName)
	return cmd.Run()
}

func (m *Manager) RestartDrsService() error {
	cmd := exec.Command("systemctl", "restart", DrsServiceName)
	return cmd.Run()
}

func (m *Manager) GetDrsServiceStatus() (string, error) {
	return m.getServiceStatus(DrsServiceName)
}

// Recorder Service Management
func (m *Manager) StopRecorderService() error {
	cmd := exec.Command("systemctl", "stop", RecorderServiceName)
	return cmd.Run()
}

func (m *Manager) RestartRecorderService() error {
	cmd := exec.Command("systemctl", "restart", RecorderServiceName)
	return cmd.Run()
}

func (m *Manager) GetRecorderServiceStatus() (string, error) {
	return m.getServiceStatus(RecorderServiceName)
}

// ServiceInfo represents detailed service information
type ServiceInfo struct {
	Name          string
	Status        string
	Enabled       bool
	Description   string
	UptimeSeconds int64
}

// Generic service management
func (m *Manager) ManageService(serviceName, action string) (string, error) {
	switch action {
	case "start":
		return m.startService(serviceName)
	case "stop":
		return m.stopService(serviceName)
	case "restart":
		return m.restartService(serviceName)
	case "status":
		return m.getServiceStatus(serviceName)
	case "enable":
		return m.enableService(serviceName)
	case "disable":
		return m.disableService(serviceName)
	default:
		return "", fmt.Errorf("unsupported action: %s", action)
	}
}

func (m *Manager) GetServiceInfo(serviceName string) (*ServiceInfo, error) {
	info := &ServiceInfo{
		Name: serviceName,
	}

	// Get status
	status, err := m.getServiceStatus(serviceName)
	if err != nil {
		return nil, err
	}
	info.Status = strings.TrimSpace(status)

	// Get enabled status
	info.Enabled, _ = m.isServiceEnabled(serviceName)

	// Get description
	info.Description, _ = m.getServiceDescription(serviceName)

	// Get uptime if active
	if info.Status == "active" {
		info.UptimeSeconds, _ = m.getServiceUptime(serviceName)
	}

	return info, nil
}

func (m *Manager) startService(serviceName string) (string, error) {
	cmd := exec.Command("systemctl", "start", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to start service %s: %s (exit code: %v)", serviceName, string(output), err)
	}
	return "started", nil
}

func (m *Manager) stopService(serviceName string) (string, error) {
	cmd := exec.Command("systemctl", "stop", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to stop service %s: %s (exit code: %v)", serviceName, string(output), err)
	}
	return "stopped", nil
}

func (m *Manager) restartService(serviceName string) (string, error) {
	cmd := exec.Command("systemctl", "restart", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to restart service %s: %s (exit code: %v)", serviceName, string(output), err)
	}
	return "restarted", nil
}

func (m *Manager) enableService(serviceName string) (string, error) {
	cmd := exec.Command("systemctl", "enable", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to enable service %s: %s (exit code: %v)", serviceName, string(output), err)
	}
	return "enabled", nil
}

func (m *Manager) disableService(serviceName string) (string, error) {
	cmd := exec.Command("systemctl", "disable", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to disable service %s: %s (exit code: %v)", serviceName, string(output), err)
	}
	return "disabled", nil
}

func (m *Manager) isServiceEnabled(serviceName string) (bool, error) {
	cmd := exec.Command("systemctl", "is-enabled", serviceName)
	output, err := cmd.Output()
	if err != nil {
		return false, nil // Service might not exist or be disabled
	}
	status := strings.TrimSpace(string(output))
	return status == "enabled", nil
}

func (m *Manager) getServiceDescription(serviceName string) (string, error) {
	cmd := exec.Command("systemctl", "show", serviceName, "--property=Description", "--value")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (m *Manager) getServiceUptime(serviceName string) (int64, error) {
	cmd := exec.Command("systemctl", "show", serviceName, "--property=ActiveEnterTimestamp", "--value")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	timestampStr := strings.TrimSpace(string(output))
	if timestampStr == "" {
		return 0, nil
	}

	// Parse systemd timestamp (format: "Mon 2021-01-01 12:00:00 UTC")
	layouts := []string{
		"Mon 2006-01-02 15:04:05 MST",
		time.RFC3339,
		"2006-01-02 15:04:05",
	}

	var startTime time.Time
	for _, layout := range layouts {
		if t, err := time.Parse(layout, timestampStr); err == nil {
			startTime = t
			break
		}
	}

	if startTime.IsZero() {
		return 0, fmt.Errorf("could not parse timestamp: %s", timestampStr)
	}

	uptime := time.Since(startTime)
	return int64(uptime.Seconds()), nil
}

// Common helper function
func (m *Manager) getServiceStatus(serviceName string) (string, error) {
	cmd := exec.Command("systemctl", "is-active", serviceName)
	// CombinedOutput captures stderr so we can distinguish unit-missing from
	// execution failures (systemctl unavailable, DBus error, permission denied).
	output, err := cmd.CombinedOutput()
	if err != nil {
		// systemctl is-active exits non-zero for non-active states (inactive, failed,
		// activating, deactivating) but still prints the actual state to stdout.
		// "unknown" means the unit does not exist.
		status := strings.TrimSpace(string(output))
		switch status {
		case "inactive", "failed", "activating", "deactivating":
			return status, nil
		case "unknown":
			return "not-found", fmt.Errorf("service not found: %s", serviceName)
		default:
			// Empty output or unrecognised text means the command itself failed
			// (e.g. systemctl unavailable, DBus error, permission denied).
			return "unknown", fmt.Errorf("getServiceStatus %s: %w (output: %q)", serviceName, err, status)
		}
	}

	return strings.TrimSpace(string(output)), nil
}
