package system

import (
	"fmt"
	"os/exec"
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
		cmd = exec.Command("sudo", "reboot", "now")
	} else {
		cmd = exec.Command("sudo", "shutdown", "-r", fmt.Sprintf("+%d", delaySeconds/60))
	}

	return cmd.Run()
}

func (m *Manager) Shutdown(delaySeconds int) error {
	if delaySeconds < 0 {
		return fmt.Errorf("delay cannot be negative")
	}

	var cmd *exec.Cmd
	if delaySeconds == 0 {
		cmd = exec.Command("sudo", "shutdown", "now")
	} else {
		cmd = exec.Command("sudo", "shutdown", fmt.Sprintf("+%d", delaySeconds/60))
	}

	return cmd.Run()
}

const (
	DrsServiceName      = "drs.service"
	RecorderServiceName = "recorder.service"
)

// DRS Service Management
func (m *Manager) StopDrsService() error {
	cmd := exec.Command("sudo", "systemctl", "stop", DrsServiceName)
	return cmd.Run()
}

func (m *Manager) RestartDrsService() error {
	cmd := exec.Command("sudo", "systemctl", "restart", DrsServiceName)
	return cmd.Run()
}

func (m *Manager) GetDrsServiceStatus() (string, error) {
	return m.getServiceStatus(DrsServiceName)
}

// Recorder Service Management
func (m *Manager) StopRecorderService() error {
	cmd := exec.Command("sudo", "systemctl", "stop", RecorderServiceName)
	return cmd.Run()
}

func (m *Manager) RestartRecorderService() error {
	cmd := exec.Command("sudo", "systemctl", "restart", RecorderServiceName)
	return cmd.Run()
}

func (m *Manager) GetRecorderServiceStatus() (string, error) {
	return m.getServiceStatus(RecorderServiceName)
}

// Common helper function
func (m *Manager) getServiceStatus(serviceName string) (string, error) {
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.Output()
	if err != nil {
		// Check if service exists
		checkCmd := exec.Command("systemctl", "list-units", "--all", serviceName)
		if _, checkErr := checkCmd.Output(); checkErr != nil {
			return "not-found", fmt.Errorf("service not found: %s", serviceName)
		}
		return "inactive", nil
	}

	return string(output), nil
}