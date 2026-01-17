package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLite implements the storage interface using SQLite
type SQLite struct {
	db *sql.DB
}

// NewSQLite creates a new SQLite storage instance
func NewSQLite(path string) (*SQLite, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	s := &SQLite{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return s, nil
}

// Close closes the database connection
func (s *SQLite) Close() error {
	return s.db.Close()
}

// migrate runs database migrations
func (s *SQLite) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS device_configs (
			mac TEXT PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			enabled BOOLEAN NOT NULL DEFAULT 1,
			block_outside BOOLEAN NOT NULL DEFAULT 0,
			schedules TEXT NOT NULL DEFAULT '[]',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS block_usage (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			mac TEXT NOT NULL,
			date TEXT NOT NULL,
			block_index INTEGER NOT NULL,
			start_time TEXT NOT NULL,
			end_time TEXT NOT NULL,
			used_bytes INTEGER NOT NULL DEFAULT 0,
			used_minutes INTEGER NOT NULL DEFAULT 0,
			limit_bytes INTEGER,
			limit_minutes INTEGER,
			is_blocked BOOLEAN NOT NULL DEFAULT 0,
			blocked_reason TEXT NOT NULL DEFAULT '',
			bonus_minutes INTEGER NOT NULL DEFAULT 0,
			bonus_bytes INTEGER NOT NULL DEFAULT 0,
			last_tx_bytes INTEGER NOT NULL DEFAULT 0,
			last_rx_bytes INTEGER NOT NULL DEFAULT 0,
			last_updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(mac, date, block_index)
		)`,
		`CREATE TABLE IF NOT EXISTS device_states (
			mac TEXT PRIMARY KEY,
			is_blocked BOOLEAN NOT NULL DEFAULT 0,
			blocked_reason TEXT NOT NULL DEFAULT '',
			blocked_at DATETIME,
			unblocked_at DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_block_usage_mac_date ON block_usage(mac, date)`,
		`CREATE INDEX IF NOT EXISTS idx_block_usage_date ON block_usage(date)`,
	}

	for _, migration := range migrations {
		if _, err := s.db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// SaveDeviceConfig saves or updates a device configuration
func (s *SQLite) SaveDeviceConfig(config *DeviceConfig) error {
	schedules, err := MarshalSchedules(config.DailySchedules)
	if err != nil {
		return fmt.Errorf("failed to marshal schedules: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO device_configs (mac, name, enabled, block_outside, schedules, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(mac) DO UPDATE SET
			name = excluded.name,
			enabled = excluded.enabled,
			block_outside = excluded.block_outside,
			schedules = excluded.schedules,
			updated_at = CURRENT_TIMESTAMP
	`, config.MAC, config.Name, config.Enabled, config.BlockOutside, schedules)

	return err
}

// GetDeviceConfig retrieves a device configuration by MAC
func (s *SQLite) GetDeviceConfig(mac string) (*DeviceConfig, error) {
	var config DeviceConfig
	var schedules string

	err := s.db.QueryRow(`
		SELECT mac, name, enabled, block_outside, schedules, created_at, updated_at
		FROM device_configs WHERE mac = ?
	`, mac).Scan(&config.MAC, &config.Name, &config.Enabled, &config.BlockOutside, &schedules, &config.CreatedAt, &config.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	config.DailySchedules, err = UnmarshalSchedules(schedules)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal schedules: %w", err)
	}

	return &config, nil
}

// GetAllDeviceConfigs retrieves all device configurations
func (s *SQLite) GetAllDeviceConfigs() ([]*DeviceConfig, error) {
	rows, err := s.db.Query(`
		SELECT mac, name, enabled, block_outside, schedules, created_at, updated_at
		FROM device_configs
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*DeviceConfig
	for rows.Next() {
		var config DeviceConfig
		var schedules string

		if err := rows.Scan(&config.MAC, &config.Name, &config.Enabled, &config.BlockOutside, &schedules, &config.CreatedAt, &config.UpdatedAt); err != nil {
			return nil, err
		}

		config.DailySchedules, err = UnmarshalSchedules(schedules)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal schedules: %w", err)
		}

		configs = append(configs, &config)
	}

	return configs, rows.Err()
}

// DeleteDeviceConfig removes a device configuration
func (s *SQLite) DeleteDeviceConfig(mac string) error {
	_, err := s.db.Exec("DELETE FROM device_configs WHERE mac = ?", mac)
	return err
}

// GetOrCreateBlockUsage gets or creates a usage record for a time block
func (s *SQLite) GetOrCreateBlockUsage(mac, date string, blockIndex int, startTime, endTime string, limitMinutes *int, limitBytes *int64) (*BlockUsage, error) {
	var usage BlockUsage

	err := s.db.QueryRow(`
		SELECT id, mac, date, block_index, start_time, end_time, used_bytes, used_minutes,
			   limit_bytes, limit_minutes, is_blocked, blocked_reason, bonus_minutes, bonus_bytes,
			   last_tx_bytes, last_rx_bytes, last_updated
		FROM block_usage WHERE mac = ? AND date = ? AND block_index = ?
	`, mac, date, blockIndex).Scan(
		&usage.ID, &usage.MAC, &usage.Date, &usage.BlockIndex, &usage.StartTime, &usage.EndTime,
		&usage.UsedBytes, &usage.UsedMinutes, &usage.LimitBytes, &usage.LimitMinutes,
		&usage.IsBlocked, &usage.BlockedReason, &usage.BonusMinutes, &usage.BonusBytes,
		&usage.LastTxBytes, &usage.LastRxBytes, &usage.LastUpdated,
	)

	if err == sql.ErrNoRows {
		// Create new record
		result, err := s.db.Exec(`
			INSERT INTO block_usage (mac, date, block_index, start_time, end_time, limit_minutes, limit_bytes)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, mac, date, blockIndex, startTime, endTime, limitMinutes, limitBytes)
		if err != nil {
			return nil, err
		}

		id, _ := result.LastInsertId()
		usage = BlockUsage{
			ID:           id,
			MAC:          mac,
			Date:         date,
			BlockIndex:   blockIndex,
			StartTime:    startTime,
			EndTime:      endTime,
			LimitMinutes: limitMinutes,
			LimitBytes:   limitBytes,
			LastUpdated:  time.Now(),
		}
		return &usage, nil
	}

	if err != nil {
		return nil, err
	}

	return &usage, nil
}

// UpdateBlockUsage updates a usage record
func (s *SQLite) UpdateBlockUsage(usage *BlockUsage) error {
	_, err := s.db.Exec(`
		UPDATE block_usage SET
			used_bytes = ?, used_minutes = ?, is_blocked = ?, blocked_reason = ?,
			bonus_minutes = ?, bonus_bytes = ?, last_tx_bytes = ?, last_rx_bytes = ?,
			last_updated = CURRENT_TIMESTAMP
		WHERE id = ?
	`, usage.UsedBytes, usage.UsedMinutes, usage.IsBlocked, usage.BlockedReason,
		usage.BonusMinutes, usage.BonusBytes, usage.LastTxBytes, usage.LastRxBytes, usage.ID)
	return err
}

// GetBlockUsageForDate retrieves all usage records for a device on a date
func (s *SQLite) GetBlockUsageForDate(mac, date string) ([]*BlockUsage, error) {
	rows, err := s.db.Query(`
		SELECT id, mac, date, block_index, start_time, end_time, used_bytes, used_minutes,
			   limit_bytes, limit_minutes, is_blocked, blocked_reason, bonus_minutes, bonus_bytes,
			   last_tx_bytes, last_rx_bytes, last_updated
		FROM block_usage WHERE mac = ? AND date = ? ORDER BY block_index
	`, mac, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usages []*BlockUsage
	for rows.Next() {
		var usage BlockUsage
		if err := rows.Scan(
			&usage.ID, &usage.MAC, &usage.Date, &usage.BlockIndex, &usage.StartTime, &usage.EndTime,
			&usage.UsedBytes, &usage.UsedMinutes, &usage.LimitBytes, &usage.LimitMinutes,
			&usage.IsBlocked, &usage.BlockedReason, &usage.BonusMinutes, &usage.BonusBytes,
			&usage.LastTxBytes, &usage.LastRxBytes, &usage.LastUpdated,
		); err != nil {
			return nil, err
		}
		usages = append(usages, &usage)
	}

	return usages, rows.Err()
}

// GetUsageHistory retrieves historical usage for a device
func (s *SQLite) GetUsageHistory(mac string, days int) ([]*HistoryEntry, error) {
	rows, err := s.db.Query(`
		SELECT date, SUM(used_minutes) as total_minutes, SUM(used_bytes) as total_bytes
		FROM block_usage
		WHERE mac = ? AND date >= date('now', '-' || ? || ' days')
		GROUP BY date
		ORDER BY date DESC
	`, mac, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*HistoryEntry
	for rows.Next() {
		var entry HistoryEntry
		if err := rows.Scan(&entry.Date, &entry.TotalMinutes, &entry.TotalBytes); err != nil {
			return nil, err
		}

		// Get block details for this date
		blocks, err := s.GetBlockUsageForDate(mac, entry.Date)
		if err != nil {
			return nil, err
		}

		for _, b := range blocks {
			entry.Blocks = append(entry.Blocks, BlockSummary{
				StartTime:    b.StartTime,
				EndTime:      b.EndTime,
				UsedMinutes:  b.UsedMinutes,
				UsedBytes:    b.UsedBytes,
				LimitMinutes: b.LimitMinutes,
				LimitBytes:   b.LimitBytes,
			})
		}

		history = append(history, &entry)
	}

	return history, rows.Err()
}

// SaveDeviceState saves the current blocking state of a device
func (s *SQLite) SaveDeviceState(state *DeviceState) error {
	_, err := s.db.Exec(`
		INSERT INTO device_states (mac, is_blocked, blocked_reason, blocked_at, unblocked_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(mac) DO UPDATE SET
			is_blocked = excluded.is_blocked,
			blocked_reason = excluded.blocked_reason,
			blocked_at = excluded.blocked_at,
			unblocked_at = excluded.unblocked_at
	`, state.MAC, state.IsBlocked, state.BlockedReason, state.BlockedAt, state.UnblockedAt)
	return err
}

// GetDeviceState retrieves the current blocking state of a device
func (s *SQLite) GetDeviceState(mac string) (*DeviceState, error) {
	var state DeviceState
	var blockedAt, unblockedAt sql.NullTime

	err := s.db.QueryRow(`
		SELECT mac, is_blocked, blocked_reason, blocked_at, unblocked_at
		FROM device_states WHERE mac = ?
	`, mac).Scan(&state.MAC, &state.IsBlocked, &state.BlockedReason, &blockedAt, &unblockedAt)

	if err == sql.ErrNoRows {
		return &DeviceState{MAC: mac}, nil
	}
	if err != nil {
		return nil, err
	}

	if blockedAt.Valid {
		state.BlockedAt = blockedAt.Time
	}
	if unblockedAt.Valid {
		state.UnblockedAt = unblockedAt.Time
	}

	return &state, nil
}

// GetAllUsageForDate retrieves usage for all managed devices on a date
func (s *SQLite) GetAllUsageForDate(date string) (map[string][]*BlockUsage, error) {
	rows, err := s.db.Query(`
		SELECT id, mac, date, block_index, start_time, end_time, used_bytes, used_minutes,
			   limit_bytes, limit_minutes, is_blocked, blocked_reason, bonus_minutes, bonus_bytes,
			   last_tx_bytes, last_rx_bytes, last_updated
		FROM block_usage WHERE date = ? ORDER BY mac, block_index
	`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usages := make(map[string][]*BlockUsage)
	for rows.Next() {
		var usage BlockUsage
		if err := rows.Scan(
			&usage.ID, &usage.MAC, &usage.Date, &usage.BlockIndex, &usage.StartTime, &usage.EndTime,
			&usage.UsedBytes, &usage.UsedMinutes, &usage.LimitBytes, &usage.LimitMinutes,
			&usage.IsBlocked, &usage.BlockedReason, &usage.BonusMinutes, &usage.BonusBytes,
			&usage.LastTxBytes, &usage.LastRxBytes, &usage.LastUpdated,
		); err != nil {
			return nil, err
		}
		usages[usage.MAC] = append(usages[usage.MAC], &usage)
	}

	return usages, rows.Err()
}

// AddBonusTime adds bonus minutes to the current time block
func (s *SQLite) AddBonusTime(mac string, date string, blockIndex int, minutes int) error {
	_, err := s.db.Exec(`
		UPDATE block_usage SET bonus_minutes = bonus_minutes + ?, last_updated = CURRENT_TIMESTAMP
		WHERE mac = ? AND date = ? AND block_index = ?
	`, minutes, mac, date, blockIndex)
	return err
}

// AddBonusData adds bonus bytes to the current time block
func (s *SQLite) AddBonusData(mac string, date string, blockIndex int, bytes int64) error {
	_, err := s.db.Exec(`
		UPDATE block_usage SET bonus_bytes = bonus_bytes + ?, last_updated = CURRENT_TIMESTAMP
		WHERE mac = ? AND date = ? AND block_index = ?
	`, bytes, mac, date, blockIndex)
	return err
}
