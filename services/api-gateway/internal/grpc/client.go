package grpc

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/drs-api/services/api-gateway/internal/config"
	modulev1 "github.com/drs-api/services/api-gateway/drs/module/v1"
	ros2bridgev1 "github.com/drs-api/services/api-gateway/drs/ros2bridge/v1"
)

// ClientManager manages gRPC connections to multiple ECUs
type ClientManager struct {
	config      *config.Config
	connections map[string]*grpc.ClientConn
	clients     map[string]*ECUClients
	mu          sync.RWMutex
}

// ECUClients holds all gRPC clients for a single ECU
type ECUClients struct {
	ServiceManager modulev1.ServiceManagerServiceClient
	SystemControl  modulev1.SystemControlServiceClient
	Monitoring     modulev1.MonitoringServiceClient
	Recording      ros2bridgev1.RecordingServiceClient
	Sensing        ros2bridgev1.SensingServiceClient
}

// NewClientManager creates a new client manager
func NewClientManager(cfg *config.Config) *ClientManager {
	return &ClientManager{
		config:      cfg,
		connections: make(map[string]*grpc.ClientConn),
		clients:     make(map[string]*ECUClients),
	}
}

// GetECUClients returns the gRPC clients for a given ECU
func (cm *ClientManager) GetECUClients(hostname string) (*ECUClients, error) {
	cm.mu.RLock()
	clients, exists := cm.clients[hostname]
	cm.mu.RUnlock()

	if exists {
		return clients, nil
	}

	// Create new connection if not exists
	return cm.createECUClients(hostname)
}

// createECUClients creates gRPC clients for a given ECU
func (cm *ClientManager) createECUClients(hostname string) (*ECUClients, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Double-check if clients were created while waiting for lock
	if clients, exists := cm.clients[hostname]; exists {
		return clients, nil
	}

	// Get ECU configuration
	ecuConfig, exists := cm.config.ECUs[hostname]
	if !exists {
		return nil, fmt.Errorf("ECU %s not found in configuration", hostname)
	}

	// Create module-manager connection
	conn, err := grpc.Dial(
		ecuConfig.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(cm.config.GRPC.Timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ECU %s at %s: %v", hostname, ecuConfig.Address, err)
	}

	clients := &ECUClients{
		ServiceManager: modulev1.NewServiceManagerServiceClient(conn),
		SystemControl:  modulev1.NewSystemControlServiceClient(conn),
		Monitoring:     modulev1.NewMonitoringServiceClient(conn),
	}

	// Create ROS2 bridge connection if enabled
	if ecuConfig.HasROS2Bridge {
		ros2Conn, err := grpc.Dial(
			ecuConfig.ROS2BridgeAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithTimeout(cm.config.GRPC.Timeout),
		)
		if err != nil {
			// Log error but don't fail, as ROS2 bridge might not be available
			fmt.Printf("Warning: failed to connect to ROS2 bridge for ECU %s at %s: %v\n", hostname, ecuConfig.ROS2BridgeAddress, err)
		} else {
			clients.Recording = ros2bridgev1.NewRecordingServiceClient(ros2Conn)
			clients.Sensing = ros2bridgev1.NewSensingServiceClient(ros2Conn)
			cm.connections[hostname+"-ros2"] = ros2Conn
		}
	}

	cm.connections[hostname] = conn
	cm.clients[hostname] = clients

	return clients, nil
}

// Close closes all gRPC connections
func (cm *ClientManager) Close() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for hostname, conn := range cm.connections {
		if err := conn.Close(); err != nil {
			fmt.Printf("Error closing connection to %s: %v\n", hostname, err)
		}
	}

	cm.connections = make(map[string]*grpc.ClientConn)
	cm.clients = make(map[string]*ECUClients)
}

// GetContext returns a context with timeout for gRPC calls
func (cm *ClientManager) GetContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), cm.config.GRPC.Timeout)
}

// IsServiceEnabled checks if a service is enabled for a given ECU
func (cm *ClientManager) IsServiceEnabled(hostname, service string) bool {
	return cm.config.IsServiceEnabled(hostname, service)
}

// GetECUNames returns a list of all ECU hostnames
func (cm *ClientManager) GetECUNames() []string {
	return cm.config.GetECUNames()
}

// HealthCheck performs a health check on all ECUs
func (cm *ClientManager) HealthCheck() map[string]bool {
	results := make(map[string]bool)
	
	for _, hostname := range cm.GetECUNames() {
		clients, err := cm.GetECUClients(hostname)
		if err != nil {
			results[hostname] = false
			continue
		}

		// Try to get environment to check if service is responding
		ctx, cancel := cm.GetContext()
		_, err = clients.Monitoring.GetEnvironment(ctx, &modulev1.GetEnvironmentRequest{})
		cancel()

		results[hostname] = err == nil
	}

	return results
}