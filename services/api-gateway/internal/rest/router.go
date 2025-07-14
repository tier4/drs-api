package rest

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tier4/drs-api/services/api-gateway/internal/config"
	"github.com/tier4/drs-api/services/api-gateway/internal/grpc"
	modulev1 "github.com/tier4/drs-api/services/api-gateway/drs/module/v1"
)

// Router sets up the HTTP router with all endpoints
func NewRouter(cfg *config.Config, clientManager *grpc.ClientManager) *gin.Engine {
	// Set gin mode
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Setup CORS middleware
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		for _, allowedOrigin := range cfg.CORS.AllowedOrigins {
			if origin == allowedOrigin {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Initialize handlers
		moduleHandler := NewModuleHandler(clientManager)
		systemHandler := NewSystemHandler(clientManager)
		recordingHandler := NewRecordingHandler(clientManager)

		// Module endpoints
		v1.GET("/modules", moduleHandler.GetAllModules)
		v1.GET("/modules/:hostname", moduleHandler.GetModule)

		// System control endpoints
		v1.POST("/system/restart", systemHandler.SystemRestart)
		v1.POST("/system/shutdown", systemHandler.SystemShutdown)
		
		// Per-module system control
		v1.POST("/modules/:hostname/restart", systemHandler.ModuleRestart)
		v1.POST("/modules/:hostname/shutdown", systemHandler.ModuleShutdown)
		v1.POST("/modules/:hostname/services/restart", systemHandler.ServicesRestart)

		// Recording endpoints
		v1.GET("/recording/status", recordingHandler.GetRecordingStatus)
		v1.POST("/recording/start", recordingHandler.StartRecording)
		v1.POST("/recording/stop", recordingHandler.StopRecording)
		v1.POST("/recording/pause", recordingHandler.PauseRecording)
		v1.POST("/recording/resume", recordingHandler.ResumeRecording)

		// PTP status endpoint
		v1.GET("/ptp/status", recordingHandler.GetPTPStatus)

		// Topic status endpoint
		v1.GET("/modules/:hostname/topics/status", recordingHandler.GetTopicStatus)

		// Service management endpoints
		v1.GET("/modules/:hostname/services", func(c *gin.Context) {
			hostname := c.Param("hostname")
			
			clients, err := clientManager.GetModuleClients(hostname)
			if err != nil {
				c.JSON(404, gin.H{"error": "Module not found"})
				return
			}

			if !clientManager.IsServiceEnabled(hostname, "services") {
				c.JSON(403, gin.H{"error": "Service management disabled"})
				return
			}

			ctx, cancel := clientManager.GetContext()
			defer cancel()

			resp, err := clients.ServiceManager.ListServices(ctx, &modulev1.ListServicesRequest{})
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{"services": resp.Services})
		})

		v1.POST("/modules/:hostname/services/:service_name/start", func(c *gin.Context) {
			hostname := c.Param("hostname")
			serviceName := c.Param("service_name")
			
			clients, err := clientManager.GetModuleClients(hostname)
			if err != nil {
				c.JSON(404, gin.H{"error": "Module not found"})
				return
			}

			if !clientManager.IsServiceEnabled(hostname, "services") {
				c.JSON(403, gin.H{"error": "Service management disabled"})
				return
			}

			ctx, cancel := clientManager.GetContext()
			defer cancel()

			resp, err := clients.ServiceManager.StartService(ctx, &modulev1.StartServiceRequest{
				Name: "services/" + serviceName,
			})
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{
				"success": true,
				"message": "Service operation completed successfully",
				"service": resp.Service,
			})
		})

		v1.POST("/modules/:hostname/services/:service_name/stop", func(c *gin.Context) {
			hostname := c.Param("hostname")
			serviceName := c.Param("service_name")
			
			clients, err := clientManager.GetModuleClients(hostname)
			if err != nil {
				c.JSON(404, gin.H{"error": "Module not found"})
				return
			}

			if !clientManager.IsServiceEnabled(hostname, "services") {
				c.JSON(403, gin.H{"error": "Service management disabled"})
				return
			}

			ctx, cancel := clientManager.GetContext()
			defer cancel()

			resp, err := clients.ServiceManager.StopService(ctx, &modulev1.StopServiceRequest{
				Name: "services/" + serviceName,
			})
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{
				"success": true,
				"message": "Service operation completed successfully",
				"service": resp.Service,
			})
		})

		v1.POST("/modules/:hostname/services/:service_name/restart", func(c *gin.Context) {
			hostname := c.Param("hostname")
			serviceName := c.Param("service_name")
			
			clients, err := clientManager.GetModuleClients(hostname)
			if err != nil {
				c.JSON(404, gin.H{"error": "Module not found"})
				return
			}

			if !clientManager.IsServiceEnabled(hostname, "services") {
				c.JSON(403, gin.H{"error": "Service management disabled"})
				return
			}

			ctx, cancel := clientManager.GetContext()
			defer cancel()

			resp, err := clients.ServiceManager.RestartService(ctx, &modulev1.RestartServiceRequest{
				Name: "services/" + serviceName,
			})
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, gin.H{
				"success": true,
				"message": "Service operation completed successfully",
				"service": resp.Service,
			})
		})
	}

	return router
}