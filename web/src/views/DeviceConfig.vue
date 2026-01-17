<template>
  <div class="device-config">
    <div class="page-header">
      <div>
        <router-link to="/devices" class="back-link">&larr; Back to Devices</router-link>
        <h1>{{ config?.name || mac }}</h1>
        <p class="device-mac">{{ mac }}</p>
      </div>
      <div class="header-actions">
        <button @click="toggleEnabled" class="btn" :class="config?.enabled ? 'btn-secondary' : 'btn-success'">
          {{ config?.enabled ? 'Disable' : 'Enable' }}
        </button>
        <button @click="saveConfig" class="btn btn-primary" :disabled="saving">
          {{ saving ? 'Saving...' : 'Save Changes' }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading">Loading configuration...</div>

    <div v-else-if="config" class="config-content">
      <!-- Basic Settings -->
      <div class="card">
        <h2 class="card-title">Basic Settings</h2>

        <div class="form-group">
          <label class="form-label">Device Name</label>
          <input v-model="config.name" type="text" class="input" placeholder="Enter device name" />
        </div>

        <div class="form-group">
          <label class="checkbox-label">
            <input v-model="config.block_outside_time_blocks" type="checkbox" />
            <span>Block device outside of defined time blocks</span>
          </label>
          <p class="form-hint">When enabled, the device will be blocked when not in an active time block.</p>
        </div>
      </div>

      <!-- Schedules -->
      <div class="card">
        <div class="card-header">
          <h2 class="card-title">Schedules</h2>
          <button @click="addSchedule" class="btn btn-secondary btn-sm">Add Schedule</button>
        </div>

        <div v-if="config.daily_schedules.length === 0" class="empty-state">
          <p>No schedules defined. Add a schedule to set time limits.</p>
        </div>

        <div v-for="(schedule, scheduleIndex) in config.daily_schedules" :key="scheduleIndex" class="schedule-card">
          <div class="schedule-header">
            <h3>Schedule {{ scheduleIndex + 1 }}</h3>
            <button @click="removeSchedule(scheduleIndex)" class="btn btn-danger btn-sm">Remove</button>
          </div>

          <!-- Day Selection -->
          <div class="form-group">
            <label class="form-label">Days</label>
            <div class="day-selector">
              <label v-for="day in dayOptions" :key="day.value" class="day-chip" :class="{ selected: schedule.days.includes(day.value) }">
                <input type="checkbox" :value="day.value" v-model="schedule.days" hidden />
                {{ day.label }}
              </label>
            </div>
          </div>

          <!-- Time Blocks -->
          <div class="time-blocks">
            <div class="form-group">
              <div class="time-blocks-header">
                <label class="form-label">Time Blocks</label>
                <button @click="addTimeBlock(scheduleIndex)" class="btn btn-secondary btn-sm">Add Block</button>
              </div>
            </div>

            <div v-for="(block, blockIndex) in schedule.time_blocks" :key="blockIndex" class="time-block">
              <div class="time-block-row">
                <div class="time-inputs">
                  <div class="form-group">
                    <label class="form-label">Start</label>
                    <input v-model="block.start_time" type="time" class="input" />
                  </div>
                  <div class="form-group">
                    <label class="form-label">End</label>
                    <input v-model="block.end_time" type="time" class="input" />
                  </div>
                </div>

                <div class="limit-inputs">
                  <div class="form-group">
                    <label class="form-label">Time Limit (minutes)</label>
                    <input
                      :value="block.limit_minutes"
                      @input="block.limit_minutes = $event.target.value ? parseInt($event.target.value) : null"
                      type="number"
                      class="input"
                      placeholder="No limit"
                      min="0"
                    />
                  </div>
                  <div class="form-group">
                    <label class="form-label">Data Limit (MB)</label>
                    <input
                      :value="block.limit_bytes ? block.limit_bytes / (1024 * 1024) : ''"
                      @input="block.limit_bytes = $event.target.value ? parseInt($event.target.value) * 1024 * 1024 : null"
                      type="number"
                      class="input"
                      placeholder="No limit"
                      min="0"
                    />
                  </div>
                </div>

                <button @click="removeTimeBlock(scheduleIndex, blockIndex)" class="btn btn-danger btn-sm remove-block-btn">
                  &times;
                </button>
              </div>
            </div>

            <div v-if="schedule.time_blocks.length === 0" class="empty-state small">
              <p>No time blocks. Add one to define usage limits.</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Current Usage -->
      <div v-if="usage" class="card">
        <h2 class="card-title">Current Usage</h2>

        <div v-if="usage.current_time_block" class="current-usage">
          <div class="usage-block-header">
            <span class="block-time">{{ usage.current_time_block.start_time }} - {{ usage.current_time_block.end_time }}</span>
            <span v-if="usage.current_time_block.is_blocked" class="badge badge-danger">Blocked</span>
          </div>

          <div class="usage-stats">
            <div v-if="usage.current_time_block.limit_minutes" class="stat">
              <div class="stat-label">Time Used</div>
              <div class="stat-value">
                {{ formatMinutes(usage.current_time_block.used_minutes) }} /
                {{ formatMinutes(usage.current_time_block.limit_minutes + (usage.current_time_block.bonus_minutes || 0)) }}
              </div>
              <div class="progress-bar">
                <div
                  class="progress-fill"
                  :class="getProgressClass(usage.current_time_block.used_minutes, usage.current_time_block.limit_minutes)"
                  :style="{ width: getProgressPercent(usage.current_time_block.used_minutes, usage.current_time_block.limit_minutes) + '%' }"
                ></div>
              </div>
            </div>

            <div v-if="usage.current_time_block.limit_bytes" class="stat">
              <div class="stat-label">Data Used</div>
              <div class="stat-value">
                {{ formatBytes(usage.current_time_block.used_bytes) }} /
                {{ formatBytes(usage.current_time_block.limit_bytes + (usage.current_time_block.bonus_bytes || 0)) }}
              </div>
              <div class="progress-bar">
                <div
                  class="progress-fill"
                  :class="getProgressClass(usage.current_time_block.used_bytes, usage.current_time_block.limit_bytes)"
                  :style="{ width: getProgressPercent(usage.current_time_block.used_bytes, usage.current_time_block.limit_bytes) + '%' }"
                ></div>
              </div>
            </div>
          </div>

          <div class="bonus-actions">
            <button @click="addBonus('time', 15)" class="btn btn-success btn-sm">+15 min</button>
            <button @click="addBonus('time', 30)" class="btn btn-success btn-sm">+30 min</button>
            <button @click="addBonus('data', 100)" class="btn btn-success btn-sm">+100 MB</button>
            <button @click="addBonus('data', 500)" class="btn btn-success btn-sm">+500 MB</button>
          </div>
        </div>

        <div v-else class="empty-state small">
          <p>Device is not in an active time block.</p>
        </div>
      </div>

      <!-- Quick Actions -->
      <div class="card">
        <h2 class="card-title">Quick Actions</h2>
        <div class="quick-actions">
          <button @click="blockNow" class="btn btn-danger">Block Now</button>
          <button @click="unblockNow" class="btn btn-success">Unblock Now</button>
        </div>
      </div>
    </div>

    <div v-if="message" class="toast" :class="messageType">{{ message }}</div>
  </div>
</template>

<script>
import { api, formatBytes, formatMinutes } from '../api'

export default {
  name: 'DeviceConfig',
  data() {
    return {
      mac: '',
      config: null,
      usage: null,
      loading: true,
      saving: false,
      message: '',
      messageType: 'success',
      dayOptions: [
        { value: 'monday', label: 'Mon' },
        { value: 'tuesday', label: 'Tue' },
        { value: 'wednesday', label: 'Wed' },
        { value: 'thursday', label: 'Thu' },
        { value: 'friday', label: 'Fri' },
        { value: 'saturday', label: 'Sat' },
        { value: 'sunday', label: 'Sun' }
      ]
    }
  },
  mounted() {
    this.mac = this.$route.params.mac
    this.fetchConfig()
  },
  methods: {
    formatBytes,
    formatMinutes,
    async fetchConfig() {
      this.loading = true
      try {
        const [config, usage] = await Promise.all([
          api.getDeviceConfig(this.mac),
          api.getDeviceUsage(this.mac).catch(() => null)
        ])
        this.config = config
        this.usage = usage
      } catch (err) {
        this.showMessage('Failed to load configuration', 'error')
      } finally {
        this.loading = false
      }
    },
    async saveConfig() {
      this.saving = true
      try {
        await api.saveDeviceConfig(this.mac, this.config)
        this.showMessage('Configuration saved successfully', 'success')
      } catch (err) {
        this.showMessage('Failed to save: ' + err.message, 'error')
      } finally {
        this.saving = false
      }
    },
    toggleEnabled() {
      this.config.enabled = !this.config.enabled
    },
    addSchedule() {
      this.config.daily_schedules.push({
        days: ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'],
        time_blocks: [
          { start_time: '08:00', end_time: '20:00', limit_minutes: 60 }
        ]
      })
    },
    removeSchedule(index) {
      this.config.daily_schedules.splice(index, 1)
    },
    addTimeBlock(scheduleIndex) {
      this.config.daily_schedules[scheduleIndex].time_blocks.push({
        start_time: '08:00',
        end_time: '20:00',
        limit_minutes: 60,
        limit_bytes: null
      })
    },
    removeTimeBlock(scheduleIndex, blockIndex) {
      this.config.daily_schedules[scheduleIndex].time_blocks.splice(blockIndex, 1)
    },
    async addBonus(type, amount) {
      try {
        if (type === 'time') {
          await api.addBonusTime(this.mac, amount)
        } else {
          await api.addBonusData(this.mac, amount, 'MB')
        }
        this.showMessage(`Added ${amount} ${type === 'time' ? 'minutes' : 'MB'}`, 'success')
        this.fetchConfig()
      } catch (err) {
        this.showMessage('Failed to add bonus: ' + err.message, 'error')
      }
    },
    async blockNow() {
      try {
        await api.blockDevice(this.mac)
        this.showMessage('Device blocked', 'success')
        this.fetchConfig()
      } catch (err) {
        this.showMessage('Failed to block: ' + err.message, 'error')
      }
    },
    async unblockNow() {
      try {
        await api.unblockDevice(this.mac)
        this.showMessage('Device unblocked', 'success')
        this.fetchConfig()
      } catch (err) {
        this.showMessage('Failed to unblock: ' + err.message, 'error')
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
    showMessage(text, type) {
      this.message = text
      this.messageType = type
      setTimeout(() => {
        this.message = ''
      }, 3000)
    }
  }
}
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
}

.back-link {
  color: #667eea;
  text-decoration: none;
  font-size: 0.875rem;
  display: block;
  margin-bottom: 0.5rem;
}

.back-link:hover {
  text-decoration: underline;
}

.page-header h1 {
  font-size: 1.75rem;
  color: #333;
  margin-bottom: 0.25rem;
}

.device-mac {
  font-family: monospace;
  color: #6c757d;
}

.header-actions {
  display: flex;
  gap: 0.5rem;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}

.form-hint {
  font-size: 0.875rem;
  color: #6c757d;
  margin-top: 0.25rem;
}

.schedule-card {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 1rem;
}

.schedule-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.schedule-header h3 {
  font-size: 1rem;
  font-weight: 600;
}

.day-selector {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.day-chip {
  padding: 0.5rem 1rem;
  background: #e9ecef;
  border-radius: 20px;
  cursor: pointer;
  transition: all 0.2s;
}

.day-chip.selected {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.time-blocks-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.time-block {
  background: white;
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 0.5rem;
}

.time-block-row {
  display: flex;
  gap: 1rem;
  align-items: flex-end;
}

.time-inputs,
.limit-inputs {
  display: flex;
  gap: 1rem;
  flex: 1;
}

.time-inputs .form-group,
.limit-inputs .form-group {
  flex: 1;
  margin-bottom: 0;
}

.remove-block-btn {
  height: 38px;
  width: 38px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
}

.current-usage {
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 8px;
}

.usage-block-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.block-time {
  font-weight: 600;
}

.usage-stats {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-bottom: 1rem;
}

.stat-label {
  font-size: 0.875rem;
  color: #6c757d;
  margin-bottom: 0.25rem;
}

.stat-value {
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.bonus-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.quick-actions {
  display: flex;
  gap: 1rem;
}

.loading, .empty-state {
  text-align: center;
  padding: 2rem;
  color: #6c757d;
}

.empty-state.small {
  padding: 1rem;
}

.toast {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  padding: 1rem 1.5rem;
  border-radius: 8px;
  color: white;
  font-weight: 500;
  animation: slideIn 0.3s ease;
}

.toast.success {
  background: #28a745;
}

.toast.error {
  background: #dc3545;
}

@keyframes slideIn {
  from {
    transform: translateY(100%);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

@media (max-width: 768px) {
  .time-block-row {
    flex-direction: column;
  }

  .time-inputs,
  .limit-inputs {
    width: 100%;
  }

  .remove-block-btn {
    width: 100%;
  }
}
</style>
