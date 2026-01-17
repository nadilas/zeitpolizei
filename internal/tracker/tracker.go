package tracker

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/nadilas/zeitpolizei/internal/enforcer"
	"github.com/nadilas/zeitpolizei/internal/storage"
	"github.com/nadilas/zeitpolizei/internal/unifi"
)

// Tracker handles polling UniFi for client stats and tracking usage
type Tracker struct {
	store        *storage.SQLite
	unifi        *unifi.Client
	enforcer     *enforcer.Enforcer
	pollInterval time.Duration
	accumulator  *Accumulator
}

// New creates a new Tracker instance
func New(store *storage.SQLite, unifiClient *unifi.Client, enf *enforcer.Enforcer, pollInterval time.Duration) *Tracker {
	return &Tracker{
		store:        store,
		unifi:        unifiClient,
		enforcer:     enf,
		pollInterval: pollInterval,
		accumulator:  NewAccumulator(store, pollInterval),
	}
}

// Start begins the tracking loop
func (t *Tracker) Start(ctx context.Context) {
	log.Printf("Starting tracker with %v poll interval", t.pollInterval)

	ticker := time.NewTicker(t.pollInterval)
	defer ticker.Stop()

	// Run immediately on start
	t.poll()

	for {
		select {
		case <-ctx.Done():
			log.Println("Tracker stopping...")
			return
		case <-ticker.C:
			t.poll()
		}
	}
}

// poll fetches client stats and processes them
func (t *Tracker) poll() {
	// Get all managed device configs
	configs, err := t.store.GetAllDeviceConfigs()
	if err != nil {
		log.Printf("Error getting device configs: %v", err)
		return
	}

	if len(configs) == 0 {
		return // No managed devices
	}

	// Create a map for quick lookup
	managedMACs := make(map[string]*storage.DeviceConfig)
	for _, cfg := range configs {
		if cfg.Enabled {
			managedMACs[strings.ToLower(cfg.MAC)] = cfg
		}
	}

	// Get current client stats from UniFi
	clients, err := t.unifi.GetClients()
	if err != nil {
		log.Printf("Error getting clients from UniFi: %v", err)
		return
	}

	now := time.Now()

	// Process each connected client that we're managing
	for _, client := range clients {
		mac := strings.ToLower(client.MAC)
		config, managed := managedMACs[mac]
		if !managed {
			continue
		}

		// Get the active time block for this device
		activeBlock, blockIndex := t.enforcer.GetActiveTimeBlock(config, now)

		if activeBlock == nil {
			// No active time block - check if we need to block
			if config.BlockOutside {
				if err := t.enforcer.BlockDevice(mac, "outside_hours"); err != nil {
					log.Printf("Error blocking device %s: %v", mac, err)
				}
			}
			continue
		}

		// Accumulate traffic for this time block
		if err := t.accumulator.ProcessClientStats(mac, &client, now, activeBlock, blockIndex); err != nil {
			log.Printf("Error accumulating stats for %s: %v", mac, err)
			continue
		}

		// Check limits and enforce
		if err := t.enforcer.CheckAndEnforce(mac, config, now); err != nil {
			log.Printf("Error enforcing limits for %s: %v", mac, err)
		}
	}

	// Check for devices that should be blocked because they're outside time blocks
	// (even if they're not currently connected)
	for mac, config := range managedMACs {
		activeBlock, _ := t.enforcer.GetActiveTimeBlock(config, now)
		if activeBlock == nil && config.BlockOutside {
			if err := t.enforcer.BlockDevice(mac, "outside_hours"); err != nil {
				log.Printf("Error blocking device %s: %v", mac, err)
			}
		}
	}
}
