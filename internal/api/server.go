package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nadilas/zeitpolizei/internal/config"
	"github.com/nadilas/zeitpolizei/internal/enforcer"
	"github.com/nadilas/zeitpolizei/internal/storage"
	"github.com/nadilas/zeitpolizei/internal/unifi"
)

// Server represents the HTTP API server
type Server struct {
	config   *config.Config
	store    *storage.SQLite
	unifi    *unifi.Client
	enforcer *enforcer.Enforcer
	router   *gin.Engine
	server   *http.Server
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, store *storage.SQLite, unifiClient *unifi.Client, enf *enforcer.Enforcer) *Server {
	gin.SetMode(gin.ReleaseMode)

	s := &Server{
		config:   cfg,
		store:    store,
		unifi:    unifiClient,
		enforcer: enf,
		router:   gin.New(),
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Recovery middleware
	s.router.Use(gin.Recovery())

	// Logger middleware
	s.router.Use(gin.Logger())

	// CORS middleware
	s.router.Use(corsMiddleware())

	// Health check
	s.router.GET("/health", s.healthCheck)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// Authentication
		v1.POST("/auth/login", s.login)

		// Protected routes
		protected := v1.Group("")
		protected.Use(s.authMiddleware())
		{
			// Devices
			protected.GET("/devices", s.listDevices)
			protected.GET("/devices/managed", s.listManagedDevices)
			protected.POST("/devices/:mac/config", s.saveDeviceConfig)
			protected.GET("/devices/:mac/config", s.getDeviceConfig)
			protected.DELETE("/devices/:mac/config", s.deleteDeviceConfig)
			protected.POST("/devices/:mac/block", s.blockDevice)
			protected.POST("/devices/:mac/unblock", s.unblockDevice)
			protected.POST("/devices/:mac/add-time", s.addBonusTime)
			protected.POST("/devices/:mac/add-data", s.addBonusData)

			// Usage
			protected.GET("/usage", s.getAllUsage)
			protected.GET("/usage/:mac", s.getDeviceUsage)
			protected.GET("/usage/:mac/history", s.getUsageHistory)

			// Status
			protected.GET("/status", s.getStatus)
		}
	}

	// Serve static files for web UI (embedded)
	s.router.NoRoute(s.serveStaticFiles)
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// authMiddleware validates authentication
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simple token-based auth
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// Validate token (simple implementation - in production use JWT)
		expectedToken := "Bearer " + generateToken(s.config.Server.Username, s.config.Server.Password)
		if token != expectedToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// generateToken creates a simple auth token
func generateToken(username, password string) string {
	// Simple token generation (in production, use proper JWT)
	return username + ":" + password
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:         s.config.Server.Address,
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}

// healthCheck returns server health status
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// serveStaticFiles serves the embedded web UI
func (s *Server) serveStaticFiles(c *gin.Context) {
	// For now, return a simple HTML page
	// In production, this would serve the embedded Vue.js app
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, `<!DOCTYPE html>
<html>
<head>
    <title>Zeitpolizei</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 40px; }
        h1 { color: #333; }
    </style>
</head>
<body>
    <h1>Zeitpolizei</h1>
    <p>UniFi Parental Control Plugin</p>
    <p>API is available at <a href="/api/v1/status">/api/v1/status</a></p>
</body>
</html>`)
}
