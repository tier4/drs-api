package models

import "time"

// ModuleStatus represents the status of a single module for the dashboard
type ModuleStatus struct {
	Hostname     string             `json:"hostname"`
	Address      string             `json:"address"`
	Status       string             `json:"status"` // OK, WARN, ERROR
	StatusDetail ModuleStatusDetail `json:"status_detail"`
	// Disk is deprecated but kept for backward compatibility. Use Disks instead.
	Disk            DiskInfo        `json:"disk"`
	Disks           []DiskDetail    `json:"disks"`
	Environment     EnvironmentInfo `json:"environment"`
	EnabledServices []string        `json:"enabled_services"`
	LastUpdated     time.Time       `json:"last_updated"`
}

// ModuleStatusDetail provides detailed status information
type ModuleStatusDetail struct {
	Services  ServiceStatus `json:"services"`
	Recording RecordingInfo `json:"recording"`
	PTP       PTPInfo       `json:"ptp"`
}

// ServiceStatus represents the status of system services
type ServiceStatus struct {
	DRSSensor   string `json:"drs_sensor"`   // active, inactive, failed
	DRSRecorder string `json:"drs_recorder"` // active, inactive, failed
}

// RecordingInfo represents recording status
type RecordingInfo struct {
	Status     string `json:"status"`      // recording, stopped, paused
	DataStatus string `json:"data_status"` // OK, WARN, ERROR
}

// PTPInfo represents PTP synchronization information
type PTPInfo struct {
	OffsetNs int64 `json:"offset_ns"`
}

// DiskInfo represents disk usage information
type DiskInfo struct {
	UsagePercentage float64 `json:"usage_percentage"`
	FreeBytes       uint64  `json:"free_bytes"`
	TotalBytes      uint64  `json:"total_bytes"`
}

// EnvironmentInfo represents environment variables
type EnvironmentInfo struct {
	SensingSystemID string `json:"sensing_system_id"`
	ModuleID        string `json:"module_id"`
}

// ModuleListResponse represents the response for GET /modules
type ModuleListResponse struct {
	Modules []ModuleStatus `json:"modules"`
}

// RecordingStatusResponse represents the response for GET /recording/status
type RecordingStatusResponse struct {
	RecordingStatus []RecordingStatus `json:"recording_status"`
}

// RecordingStatus represents recording status for a single module
type RecordingStatus struct {
	Hostname        string `json:"hostname"`
	RecordingStatus string `json:"recording_status"` // "recording" | "stopped"
	DataStatus      string `json:"data_status"`      // "OK" | "WARN" | "ERROR"
	HardwareID      string `json:"hardware_id"`
}

// PTPStatusResponse represents the response for GET /ptp/status
type PTPStatusResponse struct {
	PTPStatus []PTPStatus `json:"ptp_status"`
}

// PTPStatus represents PTP status for a single module
type PTPStatus struct {
	Hostname     string          `json:"hostname"`
	LocalStatus  PTPLocalInfo    `json:"local_status"`
	RemoteStatus []PTPRemoteInfo `json:"remote_statuses"`
}

// PTPLocalInfo represents local PTP status
type PTPLocalInfo struct {
	ClockID        string `json:"clock_id"`
	MasterOffsetNs int64  `json:"master_offset_ns"`
	GMPresent      bool   `json:"gm_present"`
}

// PTPRemoteInfo represents remote device PTP status
type PTPRemoteInfo struct {
	DeviceName   string        `json:"device_name"`
	IPAddress    string        `json:"ip_address"`
	IsReachable  bool          `json:"is_reachable"`
	Status       *PTPLocalInfo `json:"status,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// TopicStatusResponse represents the response for GET /ecus/{hostname}/topics/status
type TopicStatusResponse struct {
	Topics []TopicStatus `json:"topics"`
}

// TopicStatus represents the status of a single topic
type TopicStatus struct {
	TopicName string  `json:"topic_name"`
	RateHz    float64 `json:"rate_hz"`
	Status    string  `json:"status"` // OK, WARN, ERROR
}

// SystemOperationRequest represents a system operation request
type SystemOperationRequest struct {
	DelaySeconds int32 `json:"delay_seconds"`
}

// SystemOperationResponse represents a system operation response
type SystemOperationResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	DelaySeconds int32  `json:"delay_seconds"`
}

// ServiceOperationResponse represents a service operation response
type ServiceOperationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// RecordingOperationResponse represents a recording operation response
type RecordingOperationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Status  string `json:"status"`
}
