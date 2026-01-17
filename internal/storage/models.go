package storage

import (
	"encoding/json"
	"time"
)

// DeviceConfig represents the configuration for a managed device
type DeviceConfig struct {
	MAC            string         `json:"mac"`
	Name           string         `json:"name"`
	Enabled        bool           `json:"enabled"`
	BlockOutside   bool           `json:"block_outside_time_blocks"`
	DailySchedules []DaySchedule  `json:"daily_schedules"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// DaySchedule defines time blocks for specific days
type DaySchedule struct {
	Days       []string    `json:"days"`        // ["monday","tuesday",...] or ["weekdays","weekends"]
	TimeBlocks []TimeBlock `json:"time_blocks"`
}

// TimeBlock represents a time window with limits
type TimeBlock struct {
	StartTime               string `json:"start_time"`                        // "HH:MM" format
	EndTime                 string `json:"end_time"`                          // "HH:MM" format
	LimitMinutes            *int   `json:"limit_minutes,omitempty"`           // nil = no time limit
	LimitBytes              *int64 `json:"limit_bytes,omitempty"`             // nil = no data limit
	WarningThresholdPercent int    `json:"warning_threshold_percent"`         // default 80
}

// BlockUsage tracks usage for a specific time block on a specific day
type BlockUsage struct {
	ID            int64     `json:"id"`
	MAC           string    `json:"mac"`
	Date          string    `json:"date"`           // YYYY-MM-DD
	BlockIndex    int       `json:"block_index"`    // Index of time block
	StartTime     string    `json:"start_time"`
	EndTime       string    `json:"end_time"`
	UsedBytes     int64     `json:"used_bytes"`
	UsedMinutes   int       `json:"used_minutes"`
	LimitBytes    *int64    `json:"limit_bytes"`
	LimitMinutes  *int      `json:"limit_minutes"`
	IsBlocked     bool      `json:"is_blocked"`
	BlockedReason string    `json:"blocked_reason"` // "time_limit", "data_limit", "outside_hours", "manual"
	BonusMinutes  int       `json:"bonus_minutes"`
	BonusBytes    int64     `json:"bonus_bytes"`
	LastTxBytes   int64     `json:"last_tx_bytes"`  // For delta calculation
	LastRxBytes   int64     `json:"last_rx_bytes"`
	LastUpdated   time.Time `json:"last_updated"`
}

// DeviceState tracks the current blocking state of a device
type DeviceState struct {
	MAC           string    `json:"mac"`
	IsBlocked     bool      `json:"is_blocked"`
	BlockedReason string    `json:"blocked_reason"`
	BlockedAt     time.Time `json:"blocked_at,omitempty"`
	UnblockedAt   time.Time `json:"unblocked_at,omitempty"`
}

// UsageSummary provides a summary of usage for a device
type UsageSummary struct {
	MAC              string          `json:"mac"`
	Name             string          `json:"name"`
	CurrentBlock     *CurrentBlock   `json:"current_time_block,omitempty"`
	TodayTotal       TodayTotal      `json:"today_total"`
	AllBlocksToday   []BlockSummary  `json:"all_blocks_today"`
}

// CurrentBlock represents the currently active time block with usage
type CurrentBlock struct {
	StartTime        string `json:"start_time"`
	EndTime          string `json:"end_time"`
	LimitMinutes     *int   `json:"limit_minutes,omitempty"`
	LimitBytes       *int64 `json:"limit_bytes,omitempty"`
	UsedMinutes      int    `json:"used_minutes"`
	UsedBytes        int64  `json:"used_bytes"`
	RemainingMinutes *int   `json:"remaining_minutes,omitempty"`
	RemainingBytes   *int64 `json:"remaining_bytes,omitempty"`
	IsBlocked        bool   `json:"is_blocked"`
	BlockedReason    string `json:"blocked_reason,omitempty"`
	BonusMinutes     int    `json:"bonus_minutes,omitempty"`
	BonusBytes       int64  `json:"bonus_bytes,omitempty"`
}

// TodayTotal summarizes total usage for the day
type TodayTotal struct {
	UsedMinutes int   `json:"used_minutes"`
	UsedBytes   int64 `json:"used_bytes"`
}

// BlockSummary provides a summary of a time block
type BlockSummary struct {
	StartTime    string `json:"start"`
	EndTime      string `json:"end"`
	UsedMinutes  int    `json:"used_minutes"`
	UsedBytes    int64  `json:"used_bytes"`
	LimitMinutes *int   `json:"limit_minutes,omitempty"`
	LimitBytes   *int64 `json:"limit_bytes,omitempty"`
	Active       bool   `json:"active,omitempty"`
	Completed    bool   `json:"completed,omitempty"`
}

// HistoryEntry represents a historical usage record
type HistoryEntry struct {
	Date         string         `json:"date"`
	TotalMinutes int            `json:"total_minutes"`
	TotalBytes   int64          `json:"total_bytes"`
	Blocks       []BlockSummary `json:"blocks"`
}

// UniFiClient represents a client device from UniFi
type UniFiClient struct {
	MAC       string `json:"mac"`
	Name      string `json:"name"`
	Hostname  string `json:"hostname"`
	IP        string `json:"ip"`
	TxBytes   int64  `json:"tx_bytes"`
	RxBytes   int64  `json:"rx_bytes"`
	IsBlocked bool   `json:"is_blocked"`
	IsOnline  bool   `json:"is_online"`
}

// MarshalSchedules converts schedules to JSON for storage
func MarshalSchedules(schedules []DaySchedule) (string, error) {
	data, err := json.Marshal(schedules)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalSchedules parses schedules from JSON storage
func UnmarshalSchedules(data string) ([]DaySchedule, error) {
	var schedules []DaySchedule
	if err := json.Unmarshal([]byte(data), &schedules); err != nil {
		return nil, err
	}
	return schedules, nil
}

// ByteLimit helper for human-readable byte limits
type ByteLimit struct {
	Value int64  `json:"value"`
	Unit  string `json:"unit"` // "bytes", "KB", "MB", "GB"
}

// ToBytes converts ByteLimit to raw bytes
func (b ByteLimit) ToBytes() int64 {
	switch b.Unit {
	case "KB":
		return b.Value * 1024
	case "MB":
		return b.Value * 1024 * 1024
	case "GB":
		return b.Value * 1024 * 1024 * 1024
	default:
		return b.Value
	}
}

// FormatBytes converts bytes to human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return string(rune(bytes)) + " B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return string(rune(bytes/div)) + " " + units[exp]
}
