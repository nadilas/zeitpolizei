# Zeitpolizei

A parental control system for UniFi networks that enforces daily time and data limits on internet access per device.

## Features

- **Device Management**: Configure limits per MAC address
- **Usage Tracking**: Poll UniFi API for traffic stats, accumulate usage
- **Automatic Blocking**: Block devices via UniFi API when limit reached
- **Flexible Schedules**: Different limits for weekdays vs weekends
- **Multiple Time Blocks**: Define multiple time windows per day with individual limits
- **Bonus Time/Data**: Parents can add extra time or data on demand
- **Web Dashboard**: Manage devices, view usage, manual block/unblock

## Limit Types

| Type | Description |
|------|-------------|
| **Time-based** | Limit active internet minutes per time block |
| **Data-based** | Limit total traffic volume per time block |

When both limits are set, the first one reached triggers blocking.

## Quick Start

### 1. Build

```bash
# Build for current platform
make build

# Build for UDM (ARM64)
make build-udm

# Build for all platforms
make build-all
```

### 2. Configure

Copy and edit the example configuration:

```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml`:

```yaml
server:
  address: ":8765"
  username: "admin"
  password: "your-secure-password"

database:
  path: "zeitpolizei.db"

unifi:
  url: "https://192.168.1.1"
  username: "admin"
  password: "your-unifi-password"
  site: "default"
  is_udm: true
  insecure: true

tracker:
  poll_interval: 30s
  activity_min_bytes: 1024
```

### 3. Run

```bash
./bin/zeitpolizei -config config.yaml
```

Access the web UI at `http://localhost:8765`

## Deployment

### On UDM/UDM Pro/SE

1. Build the ARM64 binary:
   ```bash
   make build-udm
   ```

2. Copy files to UDM:
   ```bash
   scp bin/zeitpolizei-udm root@<UDM-IP>:/data/zeitpolizei/zeitpolizei
   scp config.yaml root@<UDM-IP>:/data/zeitpolizei/config.yaml
   scp deploy/on_boot.d/10-zeitpolizei.sh root@<UDM-IP>:/data/on_boot.d/
   ```

3. Make scripts executable:
   ```bash
   ssh root@<UDM-IP> "chmod +x /data/on_boot.d/10-zeitpolizei.sh /data/zeitpolizei/zeitpolizei"
   ```

4. Start Zeitpolizei:
   ```bash
   ssh root@<UDM-IP> "/data/on_boot.d/10-zeitpolizei.sh"
   ```

### Using Docker

```bash
cd deploy/docker
cp ../../config.example.yaml config.yaml
# Edit config.yaml
docker-compose up -d
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/auth/login` | POST | Authenticate |
| `/api/v1/devices` | GET | List all known devices |
| `/api/v1/devices/managed` | GET | List managed devices |
| `/api/v1/devices/:mac/config` | POST | Create/update device config |
| `/api/v1/devices/:mac/config` | DELETE | Remove device from management |
| `/api/v1/devices/:mac/block` | POST | Manual block |
| `/api/v1/devices/:mac/unblock` | POST | Manual unblock |
| `/api/v1/devices/:mac/add-time` | POST | Add bonus minutes |
| `/api/v1/devices/:mac/add-data` | POST | Add bonus bytes |
| `/api/v1/usage` | GET | Today's usage for all devices |
| `/api/v1/usage/:mac` | GET | Device usage details |
| `/api/v1/usage/:mac/history` | GET | Historical usage |
| `/api/v1/status` | GET | System health status |

## Example Device Configuration

```json
{
  "name": "Kids iPad",
  "enabled": true,
  "block_outside_time_blocks": true,
  "daily_schedules": [
    {
      "days": ["monday", "tuesday", "wednesday", "thursday", "friday"],
      "time_blocks": [
        {"start_time": "06:00", "end_time": "07:30", "limit_minutes": 30},
        {"start_time": "15:00", "end_time": "18:00", "limit_minutes": 60, "limit_bytes": 536870912}
      ]
    },
    {
      "days": ["saturday", "sunday"],
      "time_blocks": [
        {"start_time": "08:00", "end_time": "21:00", "limit_minutes": 180, "limit_bytes": 2147483648}
      ]
    }
  ]
}
```

## Hardware Support

| Hardware | Controller URL | API Prefix |
|----------|---------------|------------|
| UDM/UDM Pro/SE | `https://<ip>` | `/proxy/network` (auto) |
| Cloud Key | `https://<ip>:8443` | (none) |
| Self-hosted | `https://<ip>:8443` | (none) |

Set `is_udm: true` in config for UDM devices.

## Development

```bash
# Run with hot reload (requires air)
make dev

# Run tests
make test

# Run with coverage
make test-cover

# Format code
make fmt

# Run linter
make lint
```

## Web UI Development

```bash
cd web
npm install
npm run dev
```

The dev server runs on port 3000 and proxies API requests to :8765.

## License

MIT
