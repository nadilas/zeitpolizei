const API_BASE = '/api/v1'

function getAuthHeaders() {
  const token = localStorage.getItem('token')
  return {
    'Content-Type': 'application/json',
    'Authorization': token ? `Bearer ${token}` : ''
  }
}

async function handleResponse(response) {
  if (response.status === 401) {
    localStorage.removeItem('token')
    window.location.href = '/login'
    throw new Error('Unauthorized')
  }

  const data = await response.json()

  if (!response.ok) {
    throw new Error(data.error || 'Request failed')
  }

  return data
}

export const api = {
  async login(username, password) {
    const response = await fetch(`${API_BASE}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    })
    return handleResponse(response)
  },

  async getStatus() {
    const response = await fetch(`${API_BASE}/status`, {
      headers: getAuthHeaders()
    })
    return handleResponse(response)
  },

  async getDevices() {
    const response = await fetch(`${API_BASE}/devices`, {
      headers: getAuthHeaders()
    })
    return handleResponse(response)
  },

  async getManagedDevices() {
    const response = await fetch(`${API_BASE}/devices/managed`, {
      headers: getAuthHeaders()
    })
    return handleResponse(response)
  },

  async getDeviceConfig(mac) {
    const response = await fetch(`${API_BASE}/devices/${mac}/config`, {
      headers: getAuthHeaders()
    })
    return handleResponse(response)
  },

  async saveDeviceConfig(mac, config) {
    const response = await fetch(`${API_BASE}/devices/${mac}/config`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(config)
    })
    return handleResponse(response)
  },

  async deleteDeviceConfig(mac) {
    const response = await fetch(`${API_BASE}/devices/${mac}/config`, {
      method: 'DELETE',
      headers: getAuthHeaders()
    })
    return handleResponse(response)
  },

  async blockDevice(mac) {
    const response = await fetch(`${API_BASE}/devices/${mac}/block`, {
      method: 'POST',
      headers: getAuthHeaders()
    })
    return handleResponse(response)
  },

  async unblockDevice(mac) {
    const response = await fetch(`${API_BASE}/devices/${mac}/unblock`, {
      method: 'POST',
      headers: getAuthHeaders()
    })
    return handleResponse(response)
  },

  async addBonusTime(mac, minutes) {
    const response = await fetch(`${API_BASE}/devices/${mac}/add-time`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify({ minutes })
    })
    return handleResponse(response)
  },

  async addBonusData(mac, amount, unit) {
    const response = await fetch(`${API_BASE}/devices/${mac}/add-data`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify({ amount, unit })
    })
    return handleResponse(response)
  },

  async getAllUsage() {
    const response = await fetch(`${API_BASE}/usage`, {
      headers: getAuthHeaders()
    })
    return handleResponse(response)
  },

  async getDeviceUsage(mac) {
    const response = await fetch(`${API_BASE}/usage/${mac}`, {
      headers: getAuthHeaders()
    })
    return handleResponse(response)
  },

  async getUsageHistory(mac, days = 30) {
    const response = await fetch(`${API_BASE}/usage/${mac}/history?days=${days}`, {
      headers: getAuthHeaders()
    })
    return handleResponse(response)
  }
}

export function formatBytes(bytes) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

export function formatMinutes(minutes) {
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  const mins = minutes % 60
  return mins > 0 ? `${hours}h ${mins}m` : `${hours}h`
}
