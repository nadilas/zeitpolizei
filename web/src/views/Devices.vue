<template>
  <div class="devices">
    <div class="page-header">
      <h1>Devices</h1>
      <button @click="fetchDevices" class="btn btn-secondary">Refresh</button>
    </div>

    <div v-if="loading" class="loading">Loading devices...</div>

    <div v-else>
      <!-- Managed Devices -->
      <div class="card">
        <h2 class="card-title">Managed Devices</h2>
        <div v-if="managedDevices.length === 0" class="empty-state">
          <p>No managed devices yet. Select a device below to start managing it.</p>
        </div>
        <table v-else class="table">
          <thead>
            <tr>
              <th>Name</th>
              <th>MAC Address</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="device in managedDevices" :key="device.mac">
              <td>
                <strong>{{ device.name }}</strong>
                <span v-if="!device.enabled" class="badge badge-warning ml-1">Disabled</span>
              </td>
              <td><code>{{ device.mac }}</code></td>
              <td>
                <span v-if="getDeviceStatus(device.mac) === 'blocked'" class="badge badge-danger">Blocked</span>
                <span v-else-if="getDeviceStatus(device.mac) === 'online'" class="badge badge-success">Online</span>
                <span v-else class="badge badge-info">Offline</span>
              </td>
              <td>
                <div class="action-buttons">
                  <router-link :to="`/devices/${device.mac}`" class="btn btn-primary btn-sm">Configure</router-link>
                  <button @click="confirmRemove(device)" class="btn btn-danger btn-sm">Remove</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Available Devices -->
      <div class="card">
        <h2 class="card-title">Available Devices</h2>
        <p class="card-description">Select a device to manage its internet access.</p>

        <div class="search-box">
          <input
            v-model="searchQuery"
            type="text"
            class="input"
            placeholder="Search by name, hostname, or MAC..."
          />
        </div>

        <table class="table">
          <thead>
            <tr>
              <th>Name</th>
              <th>MAC Address</th>
              <th>IP Address</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="device in filteredDevices" :key="device.mac">
              <td>
                {{ device.name || device.hostname || '(Unknown)' }}
              </td>
              <td><code>{{ device.mac }}</code></td>
              <td>{{ device.ip || '-' }}</td>
              <td>
                <button
                  v-if="!device.is_managed"
                  @click="addDevice(device)"
                  class="btn btn-primary btn-sm"
                >
                  Add
                </button>
                <span v-else class="badge badge-info">Managed</span>
              </td>
            </tr>
          </tbody>
        </table>

        <div v-if="filteredDevices.length === 0" class="empty-state">
          <p v-if="searchQuery">No devices match your search.</p>
          <p v-else>No available devices found.</p>
        </div>
      </div>
    </div>

    <!-- Remove Confirmation Modal -->
    <div v-if="deviceToRemove" class="modal-overlay" @click.self="deviceToRemove = null">
      <div class="modal">
        <h3>Remove Device</h3>
        <p>Are you sure you want to remove <strong>{{ deviceToRemove.name }}</strong> from management?</p>
        <p class="modal-note">This will unblock the device and remove all limits.</p>
        <div class="modal-actions">
          <button @click="deviceToRemove = null" class="btn btn-secondary">Cancel</button>
          <button @click="removeDevice" class="btn btn-danger">Remove</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { api } from '../api'

export default {
  name: 'Devices',
  data() {
    return {
      allDevices: [],
      managedDevices: [],
      loading: true,
      searchQuery: '',
      deviceToRemove: null
    }
  },
  computed: {
    filteredDevices() {
      const query = this.searchQuery.toLowerCase()
      return this.allDevices.filter(d => {
        const searchText = [d.name, d.hostname, d.mac, d.ip].filter(Boolean).join(' ').toLowerCase()
        return searchText.includes(query)
      })
    }
  },
  mounted() {
    this.fetchDevices()
  },
  methods: {
    async fetchDevices() {
      this.loading = true
      try {
        const [all, managed] = await Promise.all([
          api.getDevices(),
          api.getManagedDevices()
        ])
        this.allDevices = all || []
        this.managedDevices = managed || []
      } catch (err) {
        console.error('Failed to fetch devices:', err)
      } finally {
        this.loading = false
      }
    },
    getDeviceStatus(mac) {
      const device = this.allDevices.find(d => d.mac.toLowerCase() === mac.toLowerCase())
      if (!device) return 'offline'
      if (device.is_blocked) return 'blocked'
      return 'online'
    },
    async addDevice(device) {
      const config = {
        name: device.name || device.hostname || device.mac,
        enabled: true,
        block_outside_time_blocks: false,
        daily_schedules: [
          {
            days: ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'],
            time_blocks: [
              { start_time: '06:00', end_time: '22:00', limit_minutes: 120 }
            ]
          },
          {
            days: ['saturday', 'sunday'],
            time_blocks: [
              { start_time: '08:00', end_time: '22:00', limit_minutes: 180 }
            ]
          }
        ]
      }

      try {
        await api.saveDeviceConfig(device.mac, config)
        this.$router.push(`/devices/${device.mac}`)
      } catch (err) {
        alert('Failed to add device: ' + err.message)
      }
    },
    confirmRemove(device) {
      this.deviceToRemove = device
    },
    async removeDevice() {
      try {
        await api.deleteDeviceConfig(this.deviceToRemove.mac)
        this.deviceToRemove = null
        this.fetchDevices()
      } catch (err) {
        alert('Failed to remove device: ' + err.message)
      }
    }
  }
}
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.page-header h1 {
  font-size: 1.75rem;
  color: #333;
}

.card-description {
  color: #6c757d;
  margin-bottom: 1rem;
}

.search-box {
  margin-bottom: 1rem;
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
}

.loading, .empty-state {
  text-align: center;
  padding: 2rem;
  color: #6c757d;
}

.ml-1 {
  margin-left: 0.5rem;
}

code {
  font-family: 'SF Mono', Monaco, monospace;
  font-size: 0.875rem;
  background: #f1f3f4;
  padding: 0.125rem 0.375rem;
  border-radius: 4px;
}

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  padding: 1.5rem;
  border-radius: 12px;
  max-width: 400px;
  width: 90%;
}

.modal h3 {
  margin-bottom: 1rem;
}

.modal-note {
  font-size: 0.875rem;
  color: #6c757d;
  margin-top: 0.5rem;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 1.5rem;
}
</style>
