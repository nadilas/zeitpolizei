package tracker

import (
	"time"

	"github.com/nadilas/zeitpolizei/internal/storage"
	"github.com/nadilas/zeitpolizei/internal/unifi"
)

// ActivityThreshold is the minimum bytes to consider a device "active"
const ActivityThreshold = 1024 // 1 KB

// Accumulator tracks traffic accumulation for devices
type Accumulator struct {
	store             *storage.SQLite
	pollIntervalSecs  int
}

// NewAccumulator creates a new Accumulator instance
func NewAccumulator(store *storage.SQLite, pollInterval time.Duration) *Accumulator {
	return &Accumulator{
		store:            store,
		pollIntervalSecs: int(pollInterval.Seconds()),
	}
}

// ProcessClientStats processes client statistics and accumulates usage
func (a *Accumulator) ProcessClientStats(mac string, client *unifi.ClientInfo, now time.Time, block *storage.TimeBlock, blockIndex int) error {
	date := now.Format("2006-01-02")

	// Get or create usage record for this time block
	usage, err := a.store.GetOrCreateBlockUsage(
		mac,
		date,
		blockIndex,
		block.StartTime,
		block.EndTime,
		block.LimitMinutes,
		block.LimitBytes,
	)
	if err != nil {
		return err
	}

	// Calculate traffic delta
	currentTotal := client.TxBytes + client.RxBytes
	lastTotal := usage.LastTxBytes + usage.LastRxBytes

	var delta int64
	if lastTotal == 0 {
		// First poll for this block - just record current values
		usage.LastTxBytes = client.TxBytes
		usage.LastRxBytes = client.RxBytes
		return a.store.UpdateBlockUsage(usage)
	}

	if currentTotal < lastTotal {
		// Counter reset (client reconnected) - take full current value as delta
		// This handles the case where a client disconnects and reconnects,
		// causing the UniFi controller to reset the byte counters
		delta = currentTotal
	} else {
		delta = currentTotal - lastTotal
	}

	// Update usage
	usage.UsedBytes += delta

	// Count active minutes if there was significant traffic
	// We use the poll interval to determine how many "active minutes" to add
	if delta > ActivityThreshold {
		// Add the poll interval as active time (in minutes, rounded up)
		activeMinutes := (a.pollIntervalSecs + 59) / 60
		if activeMinutes < 1 {
			activeMinutes = 1
		}
		usage.UsedMinutes += activeMinutes
	}

	// Update last seen values
	usage.LastTxBytes = client.TxBytes
	usage.LastRxBytes = client.RxBytes
	usage.LastUpdated = now

	return a.store.UpdateBlockUsage(usage)
}

// ResetForNewBlock resets tracking state for a new time block
// This is called when transitioning to a new time block
func (a *Accumulator) ResetForNewBlock(mac string, date string, blockIndex int, block *storage.TimeBlock) error {
	// Create fresh usage record for the new block
	_, err := a.store.GetOrCreateBlockUsage(
		mac,
		date,
		blockIndex,
		block.StartTime,
		block.EndTime,
		block.LimitMinutes,
		block.LimitBytes,
	)
	return err
}
