package rest

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/drs-api/services/api-gateway/internal/grpc"
	"github.com/drs-api/services/api-gateway/internal/models"
	modulev1 "github.com/drs-api/services/api-gateway/drs/module/v1"
)

// SystemHandler handles system control REST API endpoints
type SystemHandler struct {
	clientManager *grpc.ClientManager
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(clientManager *grpc.ClientManager) *SystemHandler {
	return &SystemHandler{
		clientManager: clientManager,
	}
}

// SystemRestart handles POST /system/restart - restarts all ECUs
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

// SystemShutdown handles POST /system/shutdown - shuts down all ECUs
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

// ECURestart handles POST /ecus/{hostname}/restart - restarts a single ECU
func (h *SystemHandler) ECURestart(c *gin.Context) {
	hostname := c.Param("hostname")
	
	var req models.SystemOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	h.performECUOperation(c, hostname, "restart", req.DelaySeconds)
}

// ECUShutdown handles POST /ecus/{hostname}/shutdown - shuts down a single ECU
func (h *SystemHandler) ECUShutdown(c *gin.Context) {
	hostname := c.Param("hostname")
	
	var req models.SystemOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
		return
	}

	h.performECUOperation(c, hostname, "shutdown", req.DelaySeconds)
}

// ServicesRestart handles POST /ecus/{hostname}/services/restart - restarts services on ECU
func (h *SystemHandler) ServicesRestart(c *gin.Context) {
	hostname := c.Param("hostname")
	
	clients, err := h.clientManager.GetECUClients(hostname)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "ecu_not_found",
			Message: "ECU not found or unreachable",
			Details: map[string]interface{}{
				"hostname": hostname,
			},
		})
		return
	}

	if !h.clientManager.IsServiceEnabled(hostname, "services") {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "service_disabled",
			Message: "Service management is disabled for this ECU",
		})
		return
	}

	ctx, cancel := h.clientManager.GetContext()
	defer cancel()

	// Restart critical services
	services := []string{"drs_sensor", "drs_recorder"}
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
					Message: "Service restarted successfully",
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
			Message: "All services restarted successfully",
		})
	} else {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "restart_failed",
			Message: "Some services failed to restart",
			Details: map[string]interface{}{
				"messages": messages,
			},
		})
	}
}

// performSystemOperation performs a system operation on all ECUs
func (h *SystemHandler) performSystemOperation(c *gin.Context, operation string, delaySeconds int32) {
	ecuNames := h.clientManager.GetECUNames()
	
	var wg sync.WaitGroup
	results := make(chan models.SystemOperationResponse, len(ecuNames))

	for _, hostname := range ecuNames {
		wg.Add(1)
		go func(hostname string) {
			defer wg.Done()
			
			clients, err := h.clientManager.GetECUClients(hostname)
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
		c.JSON(http.StatusOK, models.SystemOperationResponse{
			Success:      true,
			Message:      "System operation completed successfully",
			DelaySeconds: delaySeconds,
		})
	} else {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "operation_failed",
			Message: "Some ECUs failed to perform the operation",
			Details: map[string]interface{}{
				"messages": messages,
			},
		})
	}
}

// performECUOperation performs a system operation on a single ECU
func (h *SystemHandler) performECUOperation(c *gin.Context, hostname, operation string, delaySeconds int32) {
	clients, err := h.clientManager.GetECUClients(hostname)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "ecu_not_found",
			Message: "ECU not found or unreachable",
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