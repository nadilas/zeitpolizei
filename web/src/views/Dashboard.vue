<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1>Dashboard</h1>
      <button @click="fetchData" class="btn btn-secondary">
        <span v-if="loading">Refreshing...</span>
        <span v-else>Refresh</span>
      </button>
    </div>

    <!-- Status Cards -->
    <div class="status-cards">
      <div class="status-card">
        <div class="status-icon connected" v-if="status?.unifi_connected">
          <svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/></svg>
        </div>
        <div class="status-icon disconnected" v-else>
          <svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/></svg>
        </div>
        <div class="status-info">
          <div class="status-label">UniFi Controller</div>
          <div class="status-value">{{ status?.unifi_connected ? 'Connected' : 'Disconnected' }}</div>
        </div>
      </div>

      <div class="status-card">
        <div class="status-icon devices">
          <svg viewBox="0 0 24 24" fill="currentColor"><path d="M4 6h18V4H4c-1.1 0-2 .9-2 2v11H0v3h14v-3H4V6zm19 2h-6c-.55 0-1 .45-1 1v10c0 .55.45 1 1 1h6c.55 0 1-.45 1-1V9c0-.55-.45-1-1-1zm-1 9h-4v-7h4v7z"/></svg>
        </div>
        <div class="status-info">
          <div class="status-label">Managed Devices</div>
          <div class="status-value">{{ status?.managed_devices || 0 }}</div>
        </div>
      </div>

      <div class="status-card">
        <div class="status-icon blocked">
          <svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zM4 12c0-4.42 3.58-8 8-8 1.85 0 3.55.63 4.9 1.69L5.69 16.9C4.63 15.55 4 13.85 4 12zm8 8c-1.85 0-3.55-.63-4.9-1.69L18.31 7.1C19.37 8.45 20 10.15 20 12c0 4.42-3.58 8-8 8z"/></svg>
        </div>
        <div class="status-info">
          <div class="status-label">Blocked Devices</div>
          <div class="status-value">{{ status?.blocked_devices || 0 }}</div>
        </div>
      </div>
    </div>

    <!-- Device Usage -->
    <div class="card">
      <div class="card-header">
        <h2 class="card-title">Today's Usage</h2>
        <router-link to="/devices" class="btn btn-secondary btn-sm">Manage Devices</router-link>
      </div>

      <div v-if="loading" class="loading">Loading...</div>

      <div v-else-if="usage.length === 0" class="empty-state">
        <p>No managed devices yet.</p>
        <router-link to="/devices" class="btn btn-primary">Add Device</router-link>
      </div>

      <div v-else class="device-list">
        <div v-for="device in usage" :key="device.mac" class="device-card">
          <div class="device-header">
            <div class="device-info">
              <h3 class="device-name">{{ device.name || device.mac }}</h3>
              <span class="device-mac">{{ device.mac }}</span>
            </div>
            <div class="device-status">
              <span v-if="device.current_time_block?.is_blocked" class="badge badge-danger">
                Blocked
              </span>
              <span v-else-if="device.current_time_block" class="badge badge-success">
                Active
              </span>
              <span v-else class="badge badge-info">
                Outside Hours
              </span>
            </div>
          </div>

          <div v-if="device.current_time_block" class="current-block">
            <div class="block-time">
              {{ device.current_time_block.start_time }} - {{ device.current_time_block.end_time }}
            </div>

            <!-- Time Usage -->
            <div v-if="device.current_time_block.limit_minutes" class="usage-item">
              <div class="usage-header">
                <span class="usage-label">Time</span>
                <span class="usage-value">
                  {{ formatMinutes(device.current_time_block.used_minutes) }} /
                  {{ formatMinutes(device.current_time_block.limit_minutes + (device.current_time_block.bonus_minutes || 0)) }}
                </span>
              </div>
              <div class="progress-bar">
                <div
                  class="progress-fill"
                  :class="getProgressClass(device.current_time_block.used_minutes, device.current_time_block.limit_minutes + (device.current_time_block.bonus_minutes || 0))"
                  :style="{ width: getProgressPercent(device.current_time_block.used_minutes, device.current_time_block.limit_minutes + (device.current_time_block.bonus_minutes || 0)) + '%' }"
                ></div>
              </div>
            </div>

            <!-- Data Usage -->
            <div v-if="device.current_time_block.limit_bytes" class="usage-item">
              <div class="usage-header">
                <span class="usage-label">Data</span>
                <span class="usage-value">
                  {{ formatBytes(device.current_time_block.used_bytes) }} /
                  {{ formatBytes(device.current_time_block.limit_bytes + (device.current_time_block.bonus_bytes || 0)) }}
                </span>
              </div>
              <div class="progress-bar">
                <div
                  class="progress-fill"
                  :class="getProgressClass(device.current_time_block.used_bytes, device.current_time_block.limit_bytes + (device.current_time_block.bonus_bytes || 0))"
                  :style="{ width: getProgressPercent(device.current_time_block.used_bytes, device.current_time_block.limit_bytes + (device.current_time_block.bonus_bytes || 0)) + '%' }"
                ></div>
              </div>
            </div>
          </div>

          <div class="device-actions">
            <button
              v-if="device.current_time_block?.is_blocked"
              @click="addTime(device.mac)"
              class="btn btn-success btn-sm"
            >
              +15 min
            </button>
            <button
              v-if="!device.current_time_block?.is_blocked"
              @click="blockDevice(device.mac)"
              class="btn btn-danger btn-sm"
            >
              Block Now
            </button>
            <button
              v-else
              @click="unblockDevice(device.mac)"
              class="btn btn-secondary btn-sm"
            >
              Unblock
            </button>
            <router-link :to="`/devices/${device.mac}`" class="btn btn-secondary btn-sm">
              Configure
            </router-link>
          </div>

          <!-- Today's Blocks Summary -->
          <div v-if="device.all_blocks_today?.length > 1" class="blocks-summary">
            <div class="blocks-label">Today's Blocks:</div>
            <div class="blocks-list">
              <div
                v-for="(block, index) in device.all_blocks_today"
                :key="index"
                class="block-pill"
                :class="{ active: block.active, completed: block.completed }"
              >
                {{ block.start }} - {{ block.end }}
                <span v-if="block.limit_minutes">({{ formatMinutes(block.used_minutes) }}/{{ formatMinutes(block.limit_minutes) }})</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { api, formatBytes, formatMinutes } from '../api'

export default {
  name: 'Dashboard',
  data() {
    return {
      status: null,
      usage: [],
      loading: true
    }
  },
  mounted() {
    this.fetchData()
    // Auto-refresh every 30 seconds
    this.refreshInterval = setInterval(this.fetchData, 30000)
  },
  beforeUnmount() {
    clearInterval(this.refreshInterval)
  },
  methods: {
    formatBytes,
    formatMinutes,
    async fetchData() {
      this.loading = true
      try {
        const [status, usage] = await Promise.all([
          api.getStatus(),
          api.getAllUsage()
        ])
        this.status = status
        this.usage = usage || []
      } catch (err) {
        console.error('Failed to fetch data:', err)
      } finally {
        this.loading = false
      }
    },
    getProgressPercent(used, limit) {
      if (!limit) return 0
      return Math.min(100, Math.round((used / limit) * 100))
    },
    getProgressClass(used, limit) {
      const percent = this.getProgressPercent(used, limit)
      if (percent >= 100) return 'red'
      if (percent >= 80) return 'yellow'
      return 'green'
    },
    async blockDevice(mac) {
      try {
        await api.blockDevice(mac)
        this.fetchData()
      } catch (err) {
        alert('Failed to block device: ' + err.message)
      }
    },
    async unblockDevice(mac) {
      try {
        await api.unblockDevice(mac)
        this.fetchData()
      } catch (err) {
        alert('Failed to unblock device: ' + err.message)
      }
    },
    async addTime(mac) {
      try {
        await api.addBonusTime(mac, 15)
        this.fetchData()
      } catch (err) {
        alert('Failed to add time: ' + err.message)
      }
    }
  }
}
</script>

<style scoped>
.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.dashboard-header h1 {
  font-size: 1.75rem;
  color: #333;
}

.status-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.status-card {
  background: white;
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.status-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.status-icon svg {
  width: 24px;
  height: 24px;
}

.status-icon.connected {
  background: #d4edda;
  color: #28a745;
}

.status-icon.disconnected {
  background: #f8d7da;
  color: #dc3545;
}

.status-icon.devices {
  background: #d1ecf1;
  color: #17a2b8;
}

.status-icon.blocked {
  background: #fff3cd;
  color: #856404;
}

.status-label {
  font-size: 0.875rem;
  color: #6c757d;
}

.status-value {
  font-size: 1.25rem;
  font-weight: 600;
  color: #333;
}

.device-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.device-card {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 1rem;
}

.device-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.device-name {
  font-size: 1.1rem;
  font-weight: 600;
  color: #333;
  margin-bottom: 0.25rem;
}

.device-mac {
  font-size: 0.875rem;
  color: #6c757d;
  font-family: monospace;
}

.current-block {
  margin-bottom: 1rem;
}

.block-time {
  font-size: 0.875rem;
  color: #6c757d;
  margin-bottom: 0.75rem;
}

.usage-item {
  margin-bottom: 0.5rem;
}

.usage-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.25rem;
}

.usage-label {
  font-size: 0.875rem;
  color: #495057;
}

.usage-value {
  font-size: 0.875rem;
  font-weight: 500;
  color: #333;
}

.device-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.blocks-summary {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #e9ecef;
}

.blocks-label {
  font-size: 0.75rem;
  color: #6c757d;
  margin-bottom: 0.5rem;
}

.blocks-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.block-pill {
  font-size: 0.75rem;
  padding: 0.25rem 0.5rem;
  background: #e9ecef;
  border-radius: 4px;
  color: #495057;
}

.block-pill.active {
  background: #d4edda;
  color: #155724;
}

.block-pill.completed {
  background: #d1ecf1;
  color: #0c5460;
}

.loading, .empty-state {
  text-align: center;
  padding: 2rem;
  color: #6c757d;
}

.empty-state .btn {
  margin-top: 1rem;
}
</style>
