package grpc

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/tier4/drs-api/services/api-gateway/internal/config"
	modulev1 "github.com/tier4/drs-api/services/api-gateway/drs/module/v1"
	ros2bridgev1 "github.com/tier4/drs-api/services/api-gateway/drs/ros2bridge/v1"
)

// ClientManager manages gRPC connections to multiple modules
type ClientManager struct {
	config      *config.Config
	connections map[string]*grpc.ClientConn
	clients     map[string]*ModuleClients
	mu          sync.RWMutex
}

// ModuleClients holds all gRPC clients for a single module
type ModuleClients struct {
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
		clients:     make(map[string]*ModuleClients),
	}
}

// GetModuleClients returns the gRPC clients for a given module
func (cm *ClientManager) GetModuleClients(hostname string) (*ModuleClients, error) {
	cm.mu.RLock()
	clients, exists := cm.clients[hostname]
	cm.mu.RUnlock()

	if exists {
		return clients, nil
	}

	// Create new connection if not exists
	return cm.createModuleClients(hostname)
}

// createModuleClients creates gRPC clients for a given module
func (cm *ClientManager) createModuleClients(hostname string) (*ModuleClients, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Double-check if clients were created while waiting for lock
	if clients, exists := cm.clients[hostname]; exists {
		return clients, nil
	}

	// Get module configuration
	moduleConfig, exists := cm.config.Modules[hostname]
	if !exists {
		return nil, fmt.Errorf("Module %s not found in configuration", hostname)
	}

	// Create module-manager connection
	conn, err := grpc.Dial(
		moduleConfig.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(cm.config.GRPC.Timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to module %s at %s: %v", hostname, moduleConfig.Address, err)
	}

	clients := &ModuleClients{
		ServiceManager: modulev1.NewServiceManagerServiceClient(conn),
		SystemControl:  modulev1.NewSystemControlServiceClient(conn),
		Monitoring:     modulev1.NewMonitoringServiceClient(conn),
	}

	// Create ROS2 bridge connection if enabled
	if moduleConfig.HasROS2Bridge {
		ros2Conn, err := grpc.Dial(
			moduleConfig.ROS2BridgeAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithTimeout(cm.config.GRPC.Timeout),
		)
		if err != nil {
			// Log error but don't fail, as ROS2 bridge might not be available
			fmt.Printf("Warning: failed to connect to ROS2 bridge for module %s at %s: %v\n", hostname, moduleConfig.ROS2BridgeAddress, err)
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
	cm.clients = make(map[string]*ModuleClients)
}

// GetContext returns a context with timeout for gRPC calls
func (cm *ClientManager) GetContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), cm.config.GRPC.Timeout)
}

// IsServiceEnabled checks if a service is enabled for a given module
func (cm *ClientManager) IsServiceEnabled(hostname, service string) bool {
	return cm.config.IsServiceEnabled(hostname, service)
}

// GetModuleNames returns a list of all module hostnames
func (cm *ClientManager) GetModuleNames() []string {
	return cm.config.GetModuleNames()
}

// HealthCheck performs a health check on all modules
func (cm *ClientManager) HealthCheck() map[string]bool {
	results := make(map[string]bool)
	
	for _, hostname := range cm.GetModuleNames() {
		clients, err := cm.GetModuleClients(hostname)
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