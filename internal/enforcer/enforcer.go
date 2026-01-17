package enforcer

import (
	"log"
	"strings"
	"time"

	"github.com/nadilas/zeitpolizei/internal/storage"
	"github.com/nadilas/zeitpolizei/internal/unifi"
)

// Enforcer handles checking limits and blocking/unblocking devices
type Enforcer struct {
	store *storage.SQLite
	unifi *unifi.Client
}

// New creates a new Enforcer instance
func New(store *storage.SQLite, unifiClient *unifi.Client) *Enforcer {
	return &Enforcer{
		store: store,
		unifi: unifiClient,
	}
}

// GetActiveTimeBlock finds the currently active time block for a device
func (e *Enforcer) GetActiveTimeBlock(config *storage.DeviceConfig, now time.Time) (*storage.TimeBlock, int) {
	dayName := strings.ToLower(now.Weekday().String())
	currentTime := now.Format("15:04")

	for _, schedule := range config.DailySchedules {
		if !containsDay(schedule.Days, dayName) {
			continue
		}
		for i, block := range schedule.TimeBlocks {
			if currentTime >= block.StartTime && currentTime < block.EndTime {
				return &block, i
			}
		}
	}
	return nil, -1 // No active time block
}

// containsDay checks if a day is in the schedule days list
func containsDay(days []string, day string) bool {
	for _, d := range days {
		d = strings.ToLower(d)
		switch d {
		case "weekdays":
			if day != "saturday" && day != "sunday" {
				return true
			}
		case "weekends":
			if day == "saturday" || day == "sunday" {
				return true
			}
		default:
			if d == day {
				return true
			}
		}
	}
	return false
}

// CheckAndEnforce checks limits and enforces blocking if needed
func (e *Enforcer) CheckAndEnforce(mac string, config *storage.DeviceConfig, now time.Time) error {
	activeBlock, blockIndex := e.GetActiveTimeBlock(config, now)

	// Handle outside time blocks
	if activeBlock == nil {
		if config.BlockOutside {
			return e.BlockDevice(mac, "outside_hours")
		}
		return nil
	}

	// Get usage for current time block
	date := now.Format("2006-01-02")
	usage, err := e.store.GetOrCreateBlockUsage(
		mac, date, blockIndex,
		activeBlock.StartTime, activeBlock.EndTime,
		activeBlock.LimitMinutes, activeBlock.LimitBytes,
	)
	if err != nil {
		return err
	}

	// Calculate effective limits (base + bonus)
	effectiveLimitMinutes := addBonusInt(activeBlock.LimitMinutes, usage.BonusMinutes)
	effectiveLimitBytes := addBonusInt64(activeBlock.LimitBytes, usage.BonusBytes)

	// Check time limit
	if effectiveLimitMinutes != nil && usage.UsedMinutes >= *effectiveLimitMinutes {
		if !usage.IsBlocked || usage.BlockedReason != "time_limit" {
			log.Printf("Device %s reached time limit (%d/%d minutes)", mac, usage.UsedMinutes, *effectiveLimitMinutes)
			if err := e.BlockDevice(mac, "time_limit"); err != nil {
				return err
			}
			usage.IsBlocked = true
			usage.BlockedReason = "time_limit"
			return e.store.UpdateBlockUsage(usage)
		}
		return nil
	}

	// Check data limit
	if effectiveLimitBytes != nil && usage.UsedBytes >= *effectiveLimitBytes {
		if !usage.IsBlocked || usage.BlockedReason != "data_limit" {
			log.Printf("Device %s reached data limit (%d/%d bytes)", mac, usage.UsedBytes, *effectiveLimitBytes)
			if err := e.BlockDevice(mac, "data_limit"); err != nil {
				return err
			}
			usage.IsBlocked = true
			usage.BlockedReason = "data_limit"
			return e.store.UpdateBlockUsage(usage)
		}
		return nil
	}

	// Unblock if was blocked but now has remaining quota
	// (e.g., bonus time/data was added, or we're in a new time block)
	if usage.IsBlocked && usage.BlockedReason != "manual" {
		hasTimeQuota := effectiveLimitMinutes == nil || usage.UsedMinutes < *effectiveLimitMinutes
		hasDataQuota := effectiveLimitBytes == nil || usage.UsedBytes < *effectiveLimitBytes

		if hasTimeQuota && hasDataQuota {
			log.Printf("Device %s unblocked (quota available)", mac)
			if err := e.UnblockDevice(mac); err != nil {
				return err
			}
			usage.IsBlocked = false
			usage.BlockedReason = ""
			return e.store.UpdateBlockUsage(usage)
		}
	}

	return nil
}

// BlockDevice blocks a device via UniFi and updates state
func (e *Enforcer) BlockDevice(mac string, reason string) error {
	state, err := e.store.GetDeviceState(mac)
	if err != nil {
		return err
	}

	// Already blocked with same reason - no action needed
	if state.IsBlocked && state.BlockedReason == reason {
		return nil
	}

	// Block via UniFi
	if err := e.unifi.BlockClient(mac); err != nil {
		return err
	}

	// Update state
	state.IsBlocked = true
	state.BlockedReason = reason
	state.BlockedAt = time.Now()

	return e.store.SaveDeviceState(state)
}

// UnblockDevice unblocks a device via UniFi and updates state
func (e *Enforcer) UnblockDevice(mac string) error {
	state, err := e.store.GetDeviceState(mac)
	if err != nil {
		return err
	}

	// Already unblocked - no action needed
	if !state.IsBlocked {
		return nil
	}

	// Unblock via UniFi
	if err := e.unifi.UnblockClient(mac); err != nil {
		return err
	}

	// Update state
	state.IsBlocked = false
	state.BlockedReason = ""
	state.UnblockedAt = time.Now()

	return e.store.SaveDeviceState(state)
}

// ManualBlock manually blocks a device
func (e *Enforcer) ManualBlock(mac string) error {
	return e.BlockDevice(mac, "manual")
}

// ManualUnblock manually unblocks a device
func (e *Enforcer) ManualUnblock(mac string) error {
	// Get current usage to update blocked status
	now := time.Now()
	date := now.Format("2006-01-02")

	config, err := e.store.GetDeviceConfig(mac)
	if err != nil {
		return err
	}

	if config != nil {
		activeBlock, blockIndex := e.GetActiveTimeBlock(config, now)
		if activeBlock != nil {
			usage, err := e.store.GetOrCreateBlockUsage(
				mac, date, blockIndex,
				activeBlock.StartTime, activeBlock.EndTime,
				activeBlock.LimitMinutes, activeBlock.LimitBytes,
			)
			if err != nil {
				return err
			}
			usage.IsBlocked = false
			usage.BlockedReason = ""
			if err := e.store.UpdateBlockUsage(usage); err != nil {
				return err
			}
		}
	}

	return e.UnblockDevice(mac)
}

// addBonusInt adds bonus to a limit, returning nil if base is nil
func addBonusInt(base *int, bonus int) *int {
	if base == nil {
		return nil
	}
	result := *base + bonus
	return &result
}

// addBonusInt64 adds bonus to a limit, returning nil if base is nil
func addBonusInt64(base *int64, bonus int64) *int64 {
	if base == nil {
		return nil
	}
	result := *base + bonus
	return &result
}

// IsDeviceBlocked checks if a device is currently blocked
func (e *Enforcer) IsDeviceBlocked(mac string) (bool, string, error) {
	state, err := e.store.GetDeviceState(mac)
	if err != nil {
		return false, "", err
	}
	return state.IsBlocked, state.BlockedReason, nil
}
