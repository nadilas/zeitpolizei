# Zeitpolizei Installation Guide

This guide covers how to install Zeitpolizei on various platforms. Choose the method that best fits your setup.

## Table of Contents

1. [Quick Start (Download)](#quick-start-download)
2. [UDM / UDM Pro / UDM SE](#udm--udm-pro--udm-se)
3. [Docker](#docker)
4. [Synology NAS](#synology-nas)
5. [Generic Linux Server](#generic-linux-server)
6. [macOS](#macos)
7. [Building from Source](#building-from-source)
8. [Configuration](#configuration)
9. [Running as a Service](#running-as-a-service)

---

## Quick Start (Download)

The easiest way to install Zeitpolizei is to download a pre-built release:

1. Go to the [Releases page](https://github.com/nadilas/zeitpolizei/releases)

2. Download the appropriate archive for your platform:
   - `zeitpolizei-vX.X.X-linux-amd64.tar.gz` - Linux (Intel/AMD)
   - `zeitpolizei-vX.X.X-linux-arm64.tar.gz` - Linux ARM64 / UDM
   - `zeitpolizei-vX.X.X-darwin-amd64.tar.gz` - macOS (Intel)
   - `zeitpolizei-vX.X.X-darwin-arm64.tar.gz` - macOS (Apple Silicon)

3. Extract the archive:
   ```bash
   tar -xzf zeitpolizei-vX.X.X-linux-amd64.tar.gz
   cd zeitpolizei-vX.X.X-linux-amd64
   ```

4. Edit the configuration:
   ```bash
   cp config.example.yaml config.yaml
   nano config.yaml  # or your preferred editor
   ```

5. Run Zeitpolizei:
   ```bash
   ./zeitpolizei -config config.yaml
   ```

6. Access the web interface at `http://localhost:8765`

---

## UDM / UDM Pro / UDM SE

Running Zeitpolizei directly on your UniFi Dream Machine provides the best integration and eliminates the need for a separate server.

### Prerequisites

- SSH access to your UDM
- Root access enabled
- UniFi OS 2.x or later recommended

### Installation Steps

1. **Download the ARM64 release** on your computer:
   ```bash
   wget https://github.com/nadilas/zeitpolizei/releases/latest/download/zeitpolizei-linux-arm64.tar.gz
   ```

2. **Extract and prepare**:
   ```bash
   tar -xzf zeitpolizei-linux-arm64.tar.gz
   ```

3. **Copy files to UDM**:
   ```bash
   # Create directory on UDM
   ssh root@<UDM-IP> "mkdir -p /data/zeitpolizei"

   # Copy binary
   scp zeitpolizei root@<UDM-IP>:/data/zeitpolizei/zeitpolizei

   # Copy and edit config
   scp config.example.yaml root@<UDM-IP>:/data/zeitpolizei/config.yaml
   ```

4. **Edit configuration on UDM**:
   ```bash
   ssh root@<UDM-IP>
   nano /data/zeitpolizei/config.yaml
   ```

   For UDM, use these settings:
   ```yaml
   server:
     address: ":8765"
     username: "admin"
     password: "your-secure-password"

   database:
     path: "/data/zeitpolizei/zeitpolizei.db"

   unifi:
     url: "https://127.0.0.1"      # localhost since running on UDM
     username: "your-unifi-admin"
     password: "your-unifi-password"
     site: "default"
     is_udm: true
     insecure: true
   ```

5. **Install startup script**:
   ```bash
   # Copy startup script
   scp scripts/install.sh root@<UDM-IP>:/data/on_boot.d/10-zeitpolizei.sh

   # Make executable
   ssh root@<UDM-IP> "chmod +x /data/on_boot.d/10-zeitpolizei.sh /data/zeitpolizei/zeitpolizei"
   ```

6. **Start Zeitpolizei**:
   ```bash
   ssh root@<UDM-IP> "/data/on_boot.d/10-zeitpolizei.sh"
   ```

7. **Access the web interface** at `http://<UDM-IP>:8765`

### Persistence Across Reboots

The `/data/on_boot.d/` directory is used by UniFi OS to run scripts after boot. Your Zeitpolizei installation will automatically start after UDM reboots.

### Updating on UDM

```bash
# Stop current instance
ssh root@<UDM-IP> "pkill zeitpolizei"

# Upload new binary
scp zeitpolizei root@<UDM-IP>:/data/zeitpolizei/zeitpolizei

# Start again
ssh root@<UDM-IP> "/data/on_boot.d/10-zeitpolizei.sh"
```

---

## Docker

Docker is the recommended method for servers and NAS devices.

### Using Docker Compose (Recommended)

1. **Create a directory**:
   ```bash
   mkdir zeitpolizei && cd zeitpolizei
   ```

2. **Create docker-compose.yml**:
   ```yaml
   version: '3.8'
   services:
     zeitpolizei:
       image: ghcr.io/nadilas/zeitpolizei:latest
       container_name: zeitpolizei
       ports:
         - "8765:8765"
       volumes:
         - ./config.yaml:/app/config.yaml:ro
         - ./data:/app/data
       restart: unless-stopped
   ```

3. **Create config.yaml**:
   ```yaml
   server:
     address: ":8765"
     username: "admin"
     password: "your-secure-password"

   database:
     path: "/app/data/zeitpolizei.db"

   unifi:
     url: "https://192.168.1.1"
     username: "your-unifi-admin"
     password: "your-unifi-password"
     site: "default"
     is_udm: true
     insecure: true

   tracker:
     poll_interval: 30s
     activity_min_bytes: 1024
   ```

4. **Start the container**:
   ```bash
   docker-compose up -d
   ```

5. **Access** at `http://localhost:8765`

### Using Docker Run

```bash
docker run -d \
  --name zeitpolizei \
  -p 8765:8765 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  -v $(pwd)/data:/app/data \
  --restart unless-stopped \
  ghcr.io/nadilas/zeitpolizei:latest
```

---

## Synology NAS

### Using Docker (Container Manager)

1. **Open Container Manager** in DSM

2. **Download Image**:
   - Go to Registry
   - Search for `ghcr.io/nadilas/zeitpolizei`
   - Download the `latest` tag

3. **Create Container**:
   - Go to Container → Create
   - Select the zeitpolizei image
   - Configure:
     - **Port**: Map local port 8765 to container port 8765
     - **Volume**:
       - Map `/volume1/docker/zeitpolizei/config.yaml` to `/app/config.yaml` (read-only)
       - Map `/volume1/docker/zeitpolizei/data` to `/app/data`
   - Enable auto-restart

4. **Create configuration** at `/volume1/docker/zeitpolizei/config.yaml`

5. **Start the container**

### Using Task Scheduler (Native Binary)

1. Download the Linux AMD64 binary (or ARM64 for ARM-based Synology)

2. Extract to `/volume1/zeitpolizei/`

3. Create a scheduled task in Control Panel → Task Scheduler:
   - Type: Triggered Task → Boot-up
   - User: root
   - Command:
     ```bash
     /volume1/zeitpolizei/zeitpolizei -config /volume1/zeitpolizei/config.yaml
     ```

---

## Generic Linux Server

### Installation

1. **Download the release**:
   ```bash
   # For AMD64
   curl -LO https://github.com/nadilas/zeitpolizei/releases/latest/download/zeitpolizei-linux-amd64.tar.gz
   tar -xzf zeitpolizei-linux-amd64.tar.gz

   # For ARM64
   curl -LO https://github.com/nadilas/zeitpolizei/releases/latest/download/zeitpolizei-linux-arm64.tar.gz
   tar -xzf zeitpolizei-linux-arm64.tar.gz
   ```

2. **Install to /opt**:
   ```bash
   sudo mkdir -p /opt/zeitpolizei
   sudo cp zeitpolizei /opt/zeitpolizei/
   sudo cp config.example.yaml /opt/zeitpolizei/config.yaml
   sudo chmod +x /opt/zeitpolizei/zeitpolizei
   ```

3. **Edit configuration**:
   ```bash
   sudo nano /opt/zeitpolizei/config.yaml
   ```

4. **Create systemd service**:
   ```bash
   sudo tee /etc/systemd/system/zeitpolizei.service << EOF
   [Unit]
   Description=Zeitpolizei Parental Control
   After=network.target

   [Service]
   Type=simple
   User=root
   WorkingDirectory=/opt/zeitpolizei
   ExecStart=/opt/zeitpolizei/zeitpolizei -config /opt/zeitpolizei/config.yaml
   Restart=always
   RestartSec=10

   [Install]
   WantedBy=multi-user.target
   EOF
   ```

5. **Enable and start**:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable zeitpolizei
   sudo systemctl start zeitpolizei
   ```

6. **Check status**:
   ```bash
   sudo systemctl status zeitpolizei
   ```

---

## macOS

### Installation

1. **Download the release**:
   ```bash
   # For Intel Mac
   curl -LO https://github.com/nadilas/zeitpolizei/releases/latest/download/zeitpolizei-darwin-amd64.tar.gz
   tar -xzf zeitpolizei-darwin-amd64.tar.gz

   # For Apple Silicon (M1/M2/M3)
   curl -LO https://github.com/nadilas/zeitpolizei/releases/latest/download/zeitpolizei-darwin-arm64.tar.gz
   tar -xzf zeitpolizei-darwin-arm64.tar.gz
   ```

2. **Install**:
   ```bash
   mkdir -p ~/zeitpolizei
   cp zeitpolizei config.example.yaml ~/zeitpolizei/
   cd ~/zeitpolizei
   cp config.example.yaml config.yaml
   ```

3. **Edit configuration**:
   ```bash
   nano config.yaml
   ```

4. **Run**:
   ```bash
   ./zeitpolizei -config config.yaml
   ```

### Running as a Launch Agent

To have Zeitpolizei start automatically:

1. **Create launch agent**:
   ```bash
   cat > ~/Library/LaunchAgents/com.zeitpolizei.plist << EOF
   <?xml version="1.0" encoding="UTF-8"?>
   <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
   <plist version="1.0">
   <dict>
       <key>Label</key>
       <string>com.zeitpolizei</string>
       <key>ProgramArguments</key>
       <array>
           <string>${HOME}/zeitpolizei/zeitpolizei</string>
           <string>-config</string>
           <string>${HOME}/zeitpolizei/config.yaml</string>
       </array>
       <key>RunAtLoad</key>
       <true/>
       <key>KeepAlive</key>
       <true/>
       <key>WorkingDirectory</key>
       <string>${HOME}/zeitpolizei</string>
   </dict>
   </plist>
   EOF
   ```

2. **Load the agent**:
   ```bash
   launchctl load ~/Library/LaunchAgents/com.zeitpolizei.plist
   ```

---

## Building from Source

### Prerequisites

- Go 1.21 or later
- Node.js 18+ and npm (for web UI)
- Make

### Build Steps

1. **Clone the repository**:
   ```bash
   git clone https://github.com/nadilas/zeitpolizei.git
   cd zeitpolizei
   ```

2. **Build web UI**:
   ```bash
   cd web
   npm install
   npm run build
   cd ..
   ```

3. **Build the binary**:
   ```bash
   # For current platform
   make build

   # For specific platforms
   make build-linux      # Linux AMD64
   make build-linux-arm  # Linux ARM64 (UDM)
   make build-darwin     # macOS Intel
   make build-darwin-arm # macOS Apple Silicon

   # For all platforms
   make build-all
   ```

4. **Find the binary** in `bin/`

---

## Configuration

### Configuration File Reference

```yaml
# Server settings
server:
  address: ":8765"           # Listen address and port
  username: "admin"          # Web UI username
  password: "changeme"       # Web UI password (change this!)

# Database settings
database:
  path: "zeitpolizei.db"     # SQLite database file path

# UniFi Controller settings
unifi:
  url: "https://192.168.1.1" # UniFi controller URL
  username: "admin"          # UniFi admin username
  password: "password"       # UniFi admin password
  site: "default"            # UniFi site name
  is_udm: true              # true for UDM/UDM Pro/SE
  insecure: true            # Skip TLS verification

# Tracker settings
tracker:
  poll_interval: 30s         # How often to check device stats
  activity_min_bytes: 1024   # Min bytes to count as active
```

### Security Recommendations

1. **Change the default password** in the config file
2. **Use strong passwords** for both Zeitpolizei and UniFi accounts
3. **Consider firewall rules** to restrict access to port 8765
4. **Use HTTPS** if exposing outside your local network

---

## Running as a Service

### Checking Logs

**Systemd (Linux)**:
```bash
sudo journalctl -u zeitpolizei -f
```

**Docker**:
```bash
docker logs -f zeitpolizei
```

**macOS**:
```bash
tail -f /tmp/zeitpolizei.log
```

### Stopping the Service

**Systemd**:
```bash
sudo systemctl stop zeitpolizei
```

**Docker**:
```bash
docker stop zeitpolizei
```

**UDM**:
```bash
pkill zeitpolizei
```

---

## Troubleshooting Installation

### Permission Denied

```bash
chmod +x zeitpolizei
```

### Port Already in Use

Change the port in config.yaml:
```yaml
server:
  address: ":8766"  # Use a different port
```

### Cannot Connect to UniFi Controller

1. Verify the URL is correct
2. Ensure `is_udm: true` for Dream Machine devices
3. Try `insecure: true` if using self-signed certificates
4. Verify the UniFi username/password
5. Check network connectivity to the controller

### Database Errors

Ensure the database path is writable:
```bash
# Linux
sudo chown -R $(whoami) /opt/zeitpolizei/

# Docker - use the volume mapping correctly
```

---

## Next Steps

After installation, see the [User Guide](USER_GUIDE.md) to learn how to:
- Add devices to management
- Configure schedules and time limits
- Use the dashboard effectively
