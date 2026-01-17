// Mock API responses for screenshot generation
// These provide consistent, realistic data without requiring a live UniFi controller

export const mockResponses = {
  // Login response
  login: {
    token: 'mock-jwt-token-for-screenshots',
    user: 'admin'
  },

  // System status
  status: {
    unifi_connected: true,
    managed_devices: 3,
    blocked_devices: 1,
    version: '1.0.0',
    uptime: '2d 5h 30m'
  },

  // All devices from UniFi
  devices: [
    {
      mac: 'AA:BB:CC:DD:EE:01',
      name: 'Kids iPad',
      hostname: 'kids-ipad',
      ip: '192.168.1.101',
      is_managed: true,
      is_blocked: false
    },
    {
      mac: 'AA:BB:CC:DD:EE:02',
      name: 'Gaming Console',
      hostname: 'xbox-series-x',
      ip: '192.168.1.102',
      is_managed: true,
      is_blocked: true
    },
    {
      mac: 'AA:BB:CC:DD:EE:03',
      name: 'Smart TV',
      hostname: 'living-room-tv',
      ip: '192.168.1.103',
      is_managed: true,
      is_blocked: false
    },
    {
      mac: 'AA:BB:CC:DD:EE:04',
      name: 'Guest Laptop',
      hostname: 'macbook-guest',
      ip: '192.168.1.104',
      is_managed: false,
      is_blocked: false
    },
    {
      mac: 'AA:BB:CC:DD:EE:05',
      name: null,
      hostname: 'android-phone',
      ip: '192.168.1.105',
      is_managed: false,
      is_blocked: false
    }
  ],

  // Managed devices with their configurations
  managedDevices: [
    {
      mac: 'AA:BB:CC:DD:EE:01',
      name: 'Kids iPad',
      enabled: true,
      block_outside_time_blocks: true,
      daily_schedules: [
        {
          days: ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'],
          time_blocks: [
            { start_time: '06:00', end_time: '07:30', limit_minutes: 30 },
            { start_time: '15:00', end_time: '20:00', limit_minutes: 60, limit_bytes: 536870912 }
          ]
        },
        {
          days: ['saturday', 'sunday'],
          time_blocks: [
            { start_time: '08:00', end_time: '21:00', limit_minutes: 180, limit_bytes: 2147483648 }
          ]
        }
      ]
    },
    {
      mac: 'AA:BB:CC:DD:EE:02',
      name: 'Gaming Console',
      enabled: true,
      block_outside_time_blocks: false,
      daily_schedules: [
        {
          days: ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'],
          time_blocks: [
            { start_time: '16:00', end_time: '20:00', limit_minutes: 45 }
          ]
        },
        {
          days: ['saturday', 'sunday'],
          time_blocks: [
            { start_time: '10:00', end_time: '22:00', limit_minutes: 120 }
          ]
        }
      ]
    },
    {
      mac: 'AA:BB:CC:DD:EE:03',
      name: 'Smart TV',
      enabled: true,
      block_outside_time_blocks: true,
      daily_schedules: [
        {
          days: ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'],
          time_blocks: [
            { start_time: '18:00', end_time: '21:00', limit_minutes: 90 }
          ]
        }
      ]
    }
  ],

  // Usage data for dashboard - normal state
  usage: [
    {
      mac: 'AA:BB:CC:DD:EE:01',
      name: 'Kids iPad',
      current_time_block: {
        start_time: '15:00',
        end_time: '20:00',
        limit_minutes: 60,
        limit_bytes: 536870912,
        used_minutes: 35,
        used_bytes: 234567890,
        bonus_minutes: 0,
        bonus_bytes: 0,
        is_blocked: false
      },
      all_blocks_today: [
        { start: '06:00', end: '07:30', limit_minutes: 30, used_minutes: 28, active: false, completed: true },
        { start: '15:00', end: '20:00', limit_minutes: 60, used_minutes: 35, active: true, completed: false }
      ]
    },
    {
      mac: 'AA:BB:CC:DD:EE:02',
      name: 'Gaming Console',
      current_time_block: {
        start_time: '16:00',
        end_time: '20:00',
        limit_minutes: 45,
        limit_bytes: null,
        used_minutes: 45,
        used_bytes: 0,
        bonus_minutes: 0,
        bonus_bytes: 0,
        is_blocked: true
      },
      all_blocks_today: [
        { start: '16:00', end: '20:00', limit_minutes: 45, used_minutes: 45, active: true, completed: false }
      ]
    },
    {
      mac: 'AA:BB:CC:DD:EE:03',
      name: 'Smart TV',
      current_time_block: null,
      all_blocks_today: []
    }
  ],

  // Usage data with blocked device highlighted
  usageBlocked: [
    {
      mac: 'AA:BB:CC:DD:EE:01',
      name: 'Kids iPad',
      current_time_block: {
        start_time: '15:00',
        end_time: '20:00',
        limit_minutes: 60,
        limit_bytes: 536870912,
        used_minutes: 60,
        used_bytes: 536870912,
        bonus_minutes: 0,
        bonus_bytes: 0,
        is_blocked: true
      },
      all_blocks_today: [
        { start: '06:00', end: '07:30', limit_minutes: 30, used_minutes: 30, active: false, completed: true },
        { start: '15:00', end: '20:00', limit_minutes: 60, used_minutes: 60, active: true, completed: false }
      ]
    },
    {
      mac: 'AA:BB:CC:DD:EE:02',
      name: 'Gaming Console',
      current_time_block: {
        start_time: '16:00',
        end_time: '20:00',
        limit_minutes: 45,
        limit_bytes: null,
        used_minutes: 45,
        used_bytes: 0,
        bonus_minutes: 0,
        bonus_bytes: 0,
        is_blocked: true
      },
      all_blocks_today: [
        { start: '16:00', end: '20:00', limit_minutes: 45, used_minutes: 45, active: true, completed: false }
      ]
    },
    {
      mac: 'AA:BB:CC:DD:EE:03',
      name: 'Smart TV',
      current_time_block: null,
      all_blocks_today: []
    }
  ],

  // Device config for Kids iPad
  deviceConfig: {
    name: 'Kids iPad',
    enabled: true,
    block_outside_time_blocks: true,
    daily_schedules: [
      {
        days: ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'],
        time_blocks: [
          { start_time: '06:00', end_time: '07:30', limit_minutes: 30 },
          { start_time: '15:00', end_time: '20:00', limit_minutes: 60, limit_bytes: 536870912 }
        ]
      },
      {
        days: ['saturday', 'sunday'],
        time_blocks: [
          { start_time: '08:00', end_time: '21:00', limit_minutes: 180, limit_bytes: 2147483648 }
        ]
      }
    ]
  },

  // Device usage for config page
  deviceUsage: {
    mac: 'AA:BB:CC:DD:EE:01',
    name: 'Kids iPad',
    current_time_block: {
      start_time: '15:00',
      end_time: '20:00',
      limit_minutes: 60,
      limit_bytes: 536870912,
      used_minutes: 35,
      used_bytes: 234567890,
      bonus_minutes: 15,
      bonus_bytes: 104857600,
      is_blocked: false
    }
  }
};

// Helper to set up route mocking
export async function setupMockApi(page, scenario = 'default') {
  await page.route('**/api/v1/**', async (route) => {
    const url = route.request().url();
    const method = route.request().method();

    // Login
    if (url.includes('/auth/login') && method === 'POST') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockResponses.login)
      });
    }

    // Status
    if (url.includes('/status') && method === 'GET') {
      const status = scenario === 'blocked'
        ? { ...mockResponses.status, blocked_devices: 2 }
        : mockResponses.status;
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(status)
      });
    }

    // All devices
    if (url.match(/\/devices$/) && method === 'GET') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockResponses.devices)
      });
    }

    // Managed devices
    if (url.includes('/devices/managed') && method === 'GET') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockResponses.managedDevices)
      });
    }

    // Device config
    if (url.match(/\/devices\/[^/]+\/config/) && method === 'GET') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockResponses.deviceConfig)
      });
    }

    // Usage - all devices
    if (url.match(/\/usage$/) && method === 'GET') {
      const usage = scenario === 'blocked'
        ? mockResponses.usageBlocked
        : mockResponses.usage;
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(usage)
      });
    }

    // Usage - specific device
    if (url.match(/\/usage\/[^/]+$/) && method === 'GET') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockResponses.deviceUsage)
      });
    }

    // Default: pass through or return empty
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({})
    });
  });
}
