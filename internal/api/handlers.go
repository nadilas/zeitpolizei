package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nadilas/zeitpolizei/internal/storage"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token string `json:"token"`
}

// login handles user authentication
func (s *Server) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Username != s.config.Server.Username || req.Password != s.config.Server.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token := generateToken(req.Username, req.Password)
	c.JSON(http.StatusOK, LoginResponse{Token: token})
}

// listDevices returns all known devices from UniFi
func (s *Server) listDevices(c *gin.Context) {
	clients, err := s.unifi.GetAllKnownClients()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get managed devices to mark them
	configs, _ := s.store.GetAllDeviceConfigs()
	managedMACs := make(map[string]bool)
	for _, cfg := range configs {
		managedMACs[strings.ToLower(cfg.MAC)] = true
	}

	type DeviceInfo struct {
		MAC       string `json:"mac"`
		Name      string `json:"name"`
		Hostname  string `json:"hostname"`
		IP        string `json:"ip"`
		IsBlocked bool   `json:"is_blocked"`
		IsManaged bool   `json:"is_managed"`
	}

	var devices []DeviceInfo
	for _, client := range clients {
		name := client.Name
		if name == "" {
			name = client.Hostname
		}

		devices = append(devices, DeviceInfo{
			MAC:       client.MAC,
			Name:      name,
			Hostname:  client.Hostname,
			IP:        client.IP,
			IsBlocked: client.Blocked,
			IsManaged: managedMACs[strings.ToLower(client.MAC)],
		})
	}

	c.JSON(http.StatusOK, devices)
}

// listManagedDevices returns all managed devices with their configs
func (s *Server) listManagedDevices(c *gin.Context) {
	configs, err := s.store.GetAllDeviceConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, configs)
}

// DeviceConfigRequest represents a device configuration request
type DeviceConfigRequest struct {
	Name           string                `json:"name"`
	Enabled        bool                  `json:"enabled"`
	BlockOutside   bool                  `json:"block_outside_time_blocks"`
	DailySchedules []storage.DaySchedule `json:"daily_schedules"`
}

// saveDeviceConfig creates or updates a device configuration
func (s *Server) saveDeviceConfig(c *gin.Context) {
	mac := strings.ToLower(c.Param("mac"))

	var req DeviceConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	config := &storage.DeviceConfig{
		MAC:            mac,
		Name:           req.Name,
		Enabled:        req.Enabled,
		BlockOutside:   req.BlockOutside,
		DailySchedules: req.DailySchedules,
	}

	if err := s.store.SaveDeviceConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// getDeviceConfig retrieves a device configuration
func (s *Server) getDeviceConfig(c *gin.Context) {
	mac := strings.ToLower(c.Param("mac"))

	config, err := s.store.GetDeviceConfig(mac)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// deleteDeviceConfig removes a device from management
func (s *Server) deleteDeviceConfig(c *gin.Context) {
	mac := strings.ToLower(c.Param("mac"))

	if err := s.store.DeleteDeviceConfig(mac); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Unblock the device if it was blocked
	s.enforcer.ManualUnblock(mac)

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// blockDevice manually blocks a device
func (s *Server) blockDevice(c *gin.Context) {
	mac := strings.ToLower(c.Param("mac"))

	if err := s.enforcer.ManualBlock(mac); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "blocked", "mac": mac})
}

// unblockDevice manually unblocks a device
func (s *Server) unblockDevice(c *gin.Context) {
	mac := strings.ToLower(c.Param("mac"))

	if err := s.enforcer.ManualUnblock(mac); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "unblocked", "mac": mac})
}

// AddTimeRequest represents a request to add bonus time
type AddTimeRequest struct {
	Minutes int `json:"minutes" binding:"required"`
}

// addBonusTime adds bonus minutes to the current time block
func (s *Server) addBonusTime(c *gin.Context) {
	mac := strings.ToLower(c.Param("mac"))

	var req AddTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Get device config
	config, err := s.store.GetDeviceConfig(mac)
	if err != nil || config == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	// Find current time block
	now := time.Now()
	activeBlock, blockIndex := s.enforcer.GetActiveTimeBlock(config, now)
	if activeBlock == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no active time block"})
		return
	}

	date := now.Format("2006-01-02")
	if err := s.store.AddBonusTime(mac, date, blockIndex, req.Minutes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Re-check enforcement to potentially unblock
	s.enforcer.CheckAndEnforce(mac, config, now)

	c.JSON(http.StatusOK, gin.H{"status": "added", "minutes": req.Minutes})
}

// AddDataRequest represents a request to add bonus data
type AddDataRequest struct {
	Amount int64  `json:"amount" binding:"required"`
	Unit   string `json:"unit" binding:"required"` // "bytes", "KB", "MB", "GB"
}

// addBonusData adds bonus bytes to the current time block
func (s *Server) addBonusData(c *gin.Context) {
	mac := strings.ToLower(c.Param("mac"))

	var req AddDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Convert to bytes
	byteLimit := storage.ByteLimit{Value: req.Amount, Unit: req.Unit}
	bytes := byteLimit.ToBytes()

	// Get device config
	config, err := s.store.GetDeviceConfig(mac)
	if err != nil || config == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	// Find current time block
	now := time.Now()
	activeBlock, blockIndex := s.enforcer.GetActiveTimeBlock(config, now)
	if activeBlock == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no active time block"})
		return
	}

	date := now.Format("2006-01-02")
	if err := s.store.AddBonusData(mac, date, blockIndex, bytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Re-check enforcement to potentially unblock
	s.enforcer.CheckAndEnforce(mac, config, now)

	c.JSON(http.StatusOK, gin.H{"status": "added", "bytes": bytes})
}

// getAllUsage returns today's usage for all managed devices
func (s *Server) getAllUsage(c *gin.Context) {
	date := time.Now().Format("2006-01-02")

	configs, err := s.store.GetAllDeviceConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var summaries []storage.UsageSummary
	for _, config := range configs {
		summary, err := s.buildUsageSummary(config, date)
		if err != nil {
			continue
		}
		summaries = append(summaries, *summary)
	}

	c.JSON(http.StatusOK, summaries)
}

// getDeviceUsage returns today's usage for a specific device
func (s *Server) getDeviceUsage(c *gin.Context) {
	mac := strings.ToLower(c.Param("mac"))
	date := time.Now().Format("2006-01-02")

	config, err := s.store.GetDeviceConfig(mac)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	summary, err := s.buildUsageSummary(config, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// buildUsageSummary builds a usage summary for a device
func (s *Server) buildUsageSummary(config *storage.DeviceConfig, date string) (*storage.UsageSummary, error) {
	now := time.Now()
	activeBlock, activeIndex := s.enforcer.GetActiveTimeBlock(config, now)

	usages, err := s.store.GetBlockUsageForDate(config.MAC, date)
	if err != nil {
		return nil, err
	}

	summary := &storage.UsageSummary{
		MAC:  config.MAC,
		Name: config.Name,
	}

	// Calculate totals and build block summaries
	for _, usage := range usages {
		summary.TodayTotal.UsedMinutes += usage.UsedMinutes
		summary.TodayTotal.UsedBytes += usage.UsedBytes

		blockSummary := storage.BlockSummary{
			StartTime:    usage.StartTime,
			EndTime:      usage.EndTime,
			UsedMinutes:  usage.UsedMinutes,
			UsedBytes:    usage.UsedBytes,
			LimitMinutes: usage.LimitMinutes,
			LimitBytes:   usage.LimitBytes,
		}

		// Check if this is the active block
		if activeBlock != nil && usage.BlockIndex == activeIndex {
			blockSummary.Active = true
		} else if now.Format("15:04") > usage.EndTime {
			blockSummary.Completed = true
		}

		summary.AllBlocksToday = append(summary.AllBlocksToday, blockSummary)
	}

	// Build current block info
	if activeBlock != nil {
		for _, usage := range usages {
			if usage.BlockIndex == activeIndex {
				currentBlock := &storage.CurrentBlock{
					StartTime:     usage.StartTime,
					EndTime:       usage.EndTime,
					LimitMinutes:  usage.LimitMinutes,
					LimitBytes:    usage.LimitBytes,
					UsedMinutes:   usage.UsedMinutes,
					UsedBytes:     usage.UsedBytes,
					IsBlocked:     usage.IsBlocked,
					BlockedReason: usage.BlockedReason,
					BonusMinutes:  usage.BonusMinutes,
					BonusBytes:    usage.BonusBytes,
				}

				// Calculate remaining
				if usage.LimitMinutes != nil {
					remaining := *usage.LimitMinutes + usage.BonusMinutes - usage.UsedMinutes
					if remaining < 0 {
						remaining = 0
					}
					currentBlock.RemainingMinutes = &remaining
				}
				if usage.LimitBytes != nil {
					remaining := *usage.LimitBytes + usage.BonusBytes - usage.UsedBytes
					if remaining < 0 {
						remaining = 0
					}
					currentBlock.RemainingBytes = &remaining
				}

				summary.CurrentBlock = currentBlock
				break
			}
		}
	}

	return summary, nil
}

// getUsageHistory returns historical usage for a device
func (s *Server) getUsageHistory(c *gin.Context) {
	mac := strings.ToLower(c.Param("mac"))

	// Default to 30 days
	days := 30
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	history, err := s.store.GetUsageHistory(mac, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// StatusResponse represents the system status
type StatusResponse struct {
	Status          string    `json:"status"`
	UniFiConnected  bool      `json:"unifi_connected"`
	ManagedDevices  int       `json:"managed_devices"`
	BlockedDevices  int       `json:"blocked_devices"`
	ServerTime      time.Time `json:"server_time"`
	Uptime          string    `json:"uptime,omitempty"`
}

// getStatus returns system health and status
func (s *Server) getStatus(c *gin.Context) {
	// Check UniFi connection
	_, unifiErr := s.unifi.GetClients()
	unifiConnected := unifiErr == nil

	// Count managed devices
	configs, _ := s.store.GetAllDeviceConfigs()
	managedCount := len(configs)

	// Count blocked devices
	blockedCount := 0
	for _, config := range configs {
		state, err := s.store.GetDeviceState(config.MAC)
		if err == nil && state.IsBlocked {
			blockedCount++
		}
	}

	status := "ok"
	if !unifiConnected {
		status = "degraded"
	}

	c.JSON(http.StatusOK, StatusResponse{
		Status:         status,
		UniFiConnected: unifiConnected,
		ManagedDevices: managedCount,
		BlockedDevices: blockedCount,
		ServerTime:     time.Now(),
	})
}
