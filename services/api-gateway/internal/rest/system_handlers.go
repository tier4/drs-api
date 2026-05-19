package rest

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	modulev1 "github.com/tier4/drs-api/services/api-gateway/gen/drs/module/v1"
	"github.com/tier4/drs-api/services/api-gateway/internal/config"
	"github.com/tier4/drs-api/services/api-gateway/internal/grpc"
	"github.com/tier4/drs-api/services/api-gateway/internal/models"
)

// SystemHandler handles system control REST API endpoints
type SystemHandler struct {
	clientManager *grpc.ClientManager
	config        *config.Config
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(clientManager *grpc.ClientManager, cfg *config.Config) *SystemHandler {
	return &SystemHandler{
		clientManager: clientManager,
		config:        cfg,
	}
}

// SystemRestart handles POST /system/restart - restarts all modules
func (h *SystemHandler) SystemRestart(c *gin.Context) {
	var req models.SystemOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	h.performSystemOperation(c, "restart", req.DelaySeconds)
}

// SystemShutdown handles POST /system/shutdown - shuts down all modules
func (h *SystemHandler) SystemShutdown(c *gin.Context) {
	var req models.SystemOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	h.performSystemOperation(c, "shutdown", req.DelaySeconds)
}

// ModuleRestart handles POST /modules/{hostname}/restart - restarts a single module
func (h *SystemHandler) ModuleRestart(c *gin.Context) {
	hostname := c.Param("hostname")

	var req models.SystemOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	h.performModuleOperation(c, hostname, "restart", req.DelaySeconds)
}

// ModuleShutdown handles POST /modules/{hostname}/shutdown - shuts down a single module
func (h *SystemHandler) ModuleShutdown(c *gin.Context) {
	hostname := c.Param("hostname")

	var req models.SystemOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	h.performModuleOperation(c, hostname, "shutdown", req.DelaySeconds)
}

// ServicesRestart handles POST /modules/{hostname}/services/restart - restarts services on module
func (h *SystemHandler) ServicesRestart(c *gin.Context) {
	hostname := c.Param("hostname")

	clients, err := h.clientManager.GetModuleClients(hostname)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "module_not_found",
			Message: "Module not found or unreachable",
			Details: map[string]interface{}{
				"hostname": hostname,
			},
		})
		return
	}

	if !h.clientManager.IsServiceEnabled(hostname, "services") {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "service_disabled",
			Message: "Service management is disabled for this module",
		})
		return
	}

	ctx, cancel := h.clientManager.GetContext()
	defer cancel()

	// Restart only sensor service
	services := []string{"drs_sensor"}
	var wg sync.WaitGroup
	results := make(chan models.ServiceOperationResponse, len(services))

	for _, serviceName := range services {
		wg.Add(1)
		go func(svcName string) {
			defer wg.Done()

			// Try to restart the service
			_, err := clients.ServiceManager.RestartService(ctx, &modulev1.RestartServiceRequest{
				Name: "services/" + svcName,
			})

			if err != nil {
				results <- models.ServiceOperationResponse{
					Success: false,
					Message: err.Error(),
				}
			} else {
				results <- models.ServiceOperationResponse{
					Success: true,
					Message: "Sensor service restarted successfully",
				}
			}
		}(serviceName)
	}

	// Wait for all operations to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Check results
	allSuccess := true
	messages := []string{}
	for result := range results {
		if !result.Success {
			allSuccess = false
		}
		messages = append(messages, result.Message)
	}

	if allSuccess {
		c.JSON(http.StatusOK, models.ServiceOperationResponse{
			Success: true,
			Message: "Sensor service restarted successfully",
		})
	} else {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "restart_failed",
			Message: "Failed to restart sensor service",
			Details: map[string]interface{}{
				"messages": messages,
			},
		})
	}
}

// performSystemOperation performs a system operation on all modules
func (h *SystemHandler) performSystemOperation(c *gin.Context, operation string, delaySeconds int32) {
	moduleNames := h.clientManager.GetModuleNames()

	// Separate API Gateway host from other modules
	var apiGatewayHost string
	var otherModules []string

	for _, hostname := range moduleNames {
		if h.config.APIGatewayHost != "" && hostname == h.config.APIGatewayHost {
			apiGatewayHost = hostname
		} else {
			otherModules = append(otherModules, hostname)
		}
	}

	var wg sync.WaitGroup
	results := make(chan models.SystemOperationResponse, len(moduleNames))

	// First, send operation to all non-API Gateway modules
	for _, hostname := range otherModules {
		wg.Add(1)
		go func(hostname string) {
			defer wg.Done()

			clients, err := h.clientManager.GetModuleClients(hostname)
			if err != nil {
				results <- models.SystemOperationResponse{
					Success: false,
					Message: err.Error(),
				}
				return
			}

			ctx, cancel := h.clientManager.GetContext()
			defer cancel()

			switch operation {
			case "restart":
				resp, err := clients.SystemControl.Reboot(ctx, &modulev1.RebootRequest{
					DelaySeconds: delaySeconds,
				})
				if err != nil {
					results <- models.SystemOperationResponse{
						Success: false,
						Message: err.Error(),
					}
				} else {
					results <- models.SystemOperationResponse{
						Success:      resp.Accepted,
						Message:      resp.Message,
						DelaySeconds: resp.ScheduledDelay,
					}
				}
			case "shutdown":
				resp, err := clients.SystemControl.Shutdown(ctx, &modulev1.ShutdownRequest{
					DelaySeconds: delaySeconds,
				})
				if err != nil {
					results <- models.SystemOperationResponse{
						Success: false,
						Message: err.Error(),
					}
				} else {
					results <- models.SystemOperationResponse{
						Success:      resp.Accepted,
						Message:      resp.Message,
						DelaySeconds: resp.ScheduledDelay,
					}
				}
			}
		}(hostname)
	}

	// Wait for all non-API Gateway operations to complete
	wg.Wait()

	// Now send operation to API Gateway host if it exists
	if apiGatewayHost != "" {
		wg.Add(1)
		go func(hostname string) {
			defer wg.Done()

			clients, err := h.clientManager.GetModuleClients(hostname)
			if err != nil {
				results <- models.SystemOperationResponse{
					Success: false,
					Message: err.Error(),
				}
				return
			}

			ctx, cancel := h.clientManager.GetContext()
			defer cancel()

			// Add extra delay for API Gateway host to ensure other modules have received their commands
			apiGatewayDelay := delaySeconds
			if apiGatewayDelay < 5 {
				apiGatewayDelay = 5 // Minimum 5 seconds delay for API Gateway
			}

			switch operation {
			case "restart":
				resp, err := clients.SystemControl.Reboot(ctx, &modulev1.RebootRequest{
					DelaySeconds: apiGatewayDelay,
				})
				if err != nil {
					results <- models.SystemOperationResponse{
						Success: false,
						Message: err.Error(),
					}
				} else {
					results <- models.SystemOperationResponse{
						Success:      resp.Accepted,
						Message:      resp.Message + " (API Gateway host - delayed)",
						DelaySeconds: resp.ScheduledDelay,
					}
				}
			case "shutdown":
				resp, err := clients.SystemControl.Shutdown(ctx, &modulev1.ShutdownRequest{
					DelaySeconds: apiGatewayDelay,
				})
				if err != nil {
					results <- models.SystemOperationResponse{
						Success: false,
						Message: err.Error(),
					}
				} else {
					results <- models.SystemOperationResponse{
						Success:      resp.Accepted,
						Message:      resp.Message + " (API Gateway host - delayed)",
						DelaySeconds: resp.ScheduledDelay,
					}
				}
			}
		}(apiGatewayHost)

		wg.Wait()
	}

	// Close results channel
	close(results)

	// Check results
	allSuccess := true
	messages := []string{}
	for result := range results {
		if !result.Success {
			allSuccess = false
		}
		messages = append(messages, result.Message)
	}

	if allSuccess {
		c.JSON(http.StatusOK, models.SystemOperationResponse{
			Success:      true,
			Message:      "System operation completed successfully",
			DelaySeconds: delaySeconds,
		})
	} else {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "operation_failed",
			Message: "Some modules failed to perform the operation",
			Details: map[string]interface{}{
				"messages": messages,
			},
		})
	}
}

// performModuleOperation performs a system operation on a single module
func (h *SystemHandler) performModuleOperation(c *gin.Context, hostname, operation string, delaySeconds int32) {
	clients, err := h.clientManager.GetModuleClients(hostname)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "module_not_found",
			Message: "Module not found or unreachable",
			Details: map[string]interface{}{
				"hostname": hostname,
			},
		})
		return
	}

	ctx, cancel := h.clientManager.GetContext()
	defer cancel()

	switch operation {
	case "restart":
		resp, err := clients.SystemControl.Reboot(ctx, &modulev1.RebootRequest{
			DelaySeconds: delaySeconds,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "restart_failed",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, models.SystemOperationResponse{
			Success:      resp.Accepted,
			Message:      resp.Message,
			DelaySeconds: resp.ScheduledDelay,
		})
	case "shutdown":
		resp, err := clients.SystemControl.Shutdown(ctx, &modulev1.ShutdownRequest{
			DelaySeconds: delaySeconds,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "shutdown_failed",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, models.SystemOperationResponse{
			Success:      resp.Accepted,
			Message:      resp.Message,
			DelaySeconds: resp.ScheduledDelay,
		})
	}
}
