#!/bin/bash
#
# Zeitpolizei Installation Script
# This script helps install Zeitpolizei on Unix-like systems.
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default installation directory
INSTALL_DIR="/opt/zeitpolizei"
SERVICE_USER="root"

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac

    case "$OS" in
        linux|darwin)
            ;;
        *)
            echo -e "${RED}Unsupported operating system: $OS${NC}"
            exit 1
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    echo -e "${GREEN}Detected platform: ${PLATFORM}${NC}"
}

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        echo -e "${YELLOW}Warning: Not running as root. Some operations may fail.${NC}"
        echo "Consider running with: sudo $0"
        read -p "Continue anyway? [y/N] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# Install binary
install_binary() {
    local binary_path="$1"

    echo "Installing Zeitpolizei to ${INSTALL_DIR}..."

    # Create installation directory
    mkdir -p "${INSTALL_DIR}"

    # Copy binary
    if [ -f "$binary_path" ]; then
        cp "$binary_path" "${INSTALL_DIR}/zeitpolizei"
    elif [ -f "./zeitpolizei" ]; then
        cp "./zeitpolizei" "${INSTALL_DIR}/zeitpolizei"
    else
        echo -e "${RED}Binary not found. Please provide the path to the binary.${NC}"
        exit 1
    fi

    chmod +x "${INSTALL_DIR}/zeitpolizei"

    # Copy config if it exists and destination doesn't
    if [ -f "./config.example.yaml" ] && [ ! -f "${INSTALL_DIR}/config.yaml" ]; then
        cp "./config.example.yaml" "${INSTALL_DIR}/config.yaml"
        echo -e "${YELLOW}Config file created at ${INSTALL_DIR}/config.yaml${NC}"
        echo -e "${YELLOW}Please edit this file with your UniFi credentials.${NC}"
    fi

    # Copy docs
    if [ -d "./docs" ]; then
        mkdir -p "${INSTALL_DIR}/docs"
        cp -r ./docs/* "${INSTALL_DIR}/docs/" 2>/dev/null || true
    fi

    echo -e "${GREEN}Binary installed successfully.${NC}"
}

# Create systemd service
create_systemd_service() {
    if [ ! -d "/etc/systemd/system" ]; then
        echo "Systemd not found, skipping service creation."
        return
    fi

    echo "Creating systemd service..."

    cat > /etc/systemd/system/zeitpolizei.service << EOF
[Unit]
Description=Zeitpolizei Parental Control
After=network.target

[Service]
Type=simple
User=${SERVICE_USER}
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/zeitpolizei -config ${INSTALL_DIR}/config.yaml
Restart=always
RestartSec=10

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${INSTALL_DIR}

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload

    echo -e "${GREEN}Systemd service created.${NC}"
    echo "To enable and start the service:"
    echo "  sudo systemctl enable zeitpolizei"
    echo "  sudo systemctl start zeitpolizei"
}

# Create UDM on_boot.d script
create_udm_script() {
    local UDM_DIR="/data/zeitpolizei"
    local ONBOOT_DIR="/data/on_boot.d"

    if [ ! -d "/data" ]; then
        echo "UDM data directory not found."
        return
    fi

    echo "Setting up for UDM..."

    # Create directories
    mkdir -p "${UDM_DIR}"
    mkdir -p "${ONBOOT_DIR}"

    # Copy binary
    if [ -f "./zeitpolizei" ]; then
        cp "./zeitpolizei" "${UDM_DIR}/zeitpolizei"
        chmod +x "${UDM_DIR}/zeitpolizei"
    fi

    # Copy config if needed
    if [ -f "./config.example.yaml" ] && [ ! -f "${UDM_DIR}/config.yaml" ]; then
        cp "./config.example.yaml" "${UDM_DIR}/config.yaml"
    fi

    # Create on_boot.d script
    cat > "${ONBOOT_DIR}/10-zeitpolizei.sh" << 'SCRIPT'
#!/bin/bash
# Zeitpolizei on_boot.d script for UDM/UDM Pro/SE

ZEITPOLIZEI_DIR="/data/zeitpolizei"
ZEITPOLIZEI_BIN="${ZEITPOLIZEI_DIR}/zeitpolizei"
ZEITPOLIZEI_CONFIG="${ZEITPOLIZEI_DIR}/config.yaml"
ZEITPOLIZEI_PID="${ZEITPOLIZEI_DIR}/zeitpolizei.pid"
LOG_FILE="${ZEITPOLIZEI_DIR}/zeitpolizei.log"

mkdir -p "${ZEITPOLIZEI_DIR}"

if [ ! -f "${ZEITPOLIZEI_BIN}" ]; then
    echo "Zeitpolizei binary not found at ${ZEITPOLIZEI_BIN}"
    exit 1
fi

if [ ! -f "${ZEITPOLIZEI_CONFIG}" ]; then
    echo "Config file not found at ${ZEITPOLIZEI_CONFIG}"
    exit 1
fi

# Stop existing process
if [ -f "${ZEITPOLIZEI_PID}" ]; then
    OLD_PID=$(cat "${ZEITPOLIZEI_PID}")
    if [ -d "/proc/${OLD_PID}" ]; then
        kill "${OLD_PID}" 2>/dev/null
        sleep 2
    fi
    rm -f "${ZEITPOLIZEI_PID}"
fi

# Start
nohup "${ZEITPOLIZEI_BIN}" -config "${ZEITPOLIZEI_CONFIG}" >> "${LOG_FILE}" 2>&1 &
echo $! > "${ZEITPOLIZEI_PID}"

echo "Zeitpolizei started with PID: $(cat ${ZEITPOLIZEI_PID})"
SCRIPT

    chmod +x "${ONBOOT_DIR}/10-zeitpolizei.sh"

    echo -e "${GREEN}UDM setup complete.${NC}"
    echo "Start Zeitpolizei with: /data/on_boot.d/10-zeitpolizei.sh"
}

# Print usage information
print_usage() {
    echo "Zeitpolizei Installation Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -d, --dir DIR       Installation directory (default: /opt/zeitpolizei)"
    echo "  -u, --user USER     Service user (default: root)"
    echo "  --udm               Install for UDM/UDM Pro/SE"
    echo "  --no-service        Skip systemd service creation"
    echo "  -h, --help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                  Install with defaults"
    echo "  $0 --udm            Install on UDM"
    echo "  $0 -d /home/user/zeitpolizei --no-service"
}

# Parse command line arguments
CREATE_SERVICE=true
UDM_MODE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        -u|--user)
            SERVICE_USER="$2"
            shift 2
            ;;
        --udm)
            UDM_MODE=true
            shift
            ;;
        --no-service)
            CREATE_SERVICE=false
            shift
            ;;
        -h|--help)
            print_usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            print_usage
            exit 1
            ;;
    esac
done

# Main installation flow
main() {
    echo "=========================================="
    echo "  Zeitpolizei Installation"
    echo "=========================================="
    echo ""

    detect_platform

    if [ "$UDM_MODE" = true ]; then
        create_udm_script
    else
        check_root
        install_binary
        if [ "$CREATE_SERVICE" = true ]; then
            create_systemd_service
        fi
    fi

    echo ""
    echo "=========================================="
    echo -e "${GREEN}  Installation Complete!${NC}"
    echo "=========================================="
    echo ""
    echo "Next steps:"
    echo "1. Edit the configuration file:"
    if [ "$UDM_MODE" = true ]; then
        echo "   nano /data/zeitpolizei/config.yaml"
    else
        echo "   nano ${INSTALL_DIR}/config.yaml"
    fi
    echo ""
    echo "2. Configure your UniFi credentials in the config file"
    echo ""
    echo "3. Start Zeitpolizei:"
    if [ "$UDM_MODE" = true ]; then
        echo "   /data/on_boot.d/10-zeitpolizei.sh"
    elif [ "$CREATE_SERVICE" = true ]; then
        echo "   sudo systemctl enable zeitpolizei"
        echo "   sudo systemctl start zeitpolizei"
    else
        echo "   ${INSTALL_DIR}/zeitpolizei -config ${INSTALL_DIR}/config.yaml"
    fi
    echo ""
    echo "4. Access the web interface:"
    echo "   http://localhost:8765"
    echo ""
}

main
