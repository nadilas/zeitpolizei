#!/bin/bash
# Zeitpolizei on_boot.d script for UDM/UDM Pro/SE
# Place this file in /data/on_boot.d/ and make it executable
# chmod +x /data/on_boot.d/10-zeitpolizei.sh

ZEITPOLIZEI_DIR="/data/zeitpolizei"
ZEITPOLIZEI_BIN="${ZEITPOLIZEI_DIR}/zeitpolizei"
ZEITPOLIZEI_CONFIG="${ZEITPOLIZEI_DIR}/config.yaml"
ZEITPOLIZEI_PID="${ZEITPOLIZEI_DIR}/zeitpolizei.pid"
LOG_FILE="${ZEITPOLIZEI_DIR}/zeitpolizei.log"

# Create directory if it doesn't exist
mkdir -p "${ZEITPOLIZEI_DIR}"

# Check if binary exists
if [ ! -f "${ZEITPOLIZEI_BIN}" ]; then
    echo "Zeitpolizei binary not found at ${ZEITPOLIZEI_BIN}"
    exit 1
fi

# Check if config exists
if [ ! -f "${ZEITPOLIZEI_CONFIG}" ]; then
    echo "Config file not found at ${ZEITPOLIZEI_CONFIG}"
    exit 1
fi

# Stop existing process if running
if [ -f "${ZEITPOLIZEI_PID}" ]; then
    OLD_PID=$(cat "${ZEITPOLIZEI_PID}")
    if [ -d "/proc/${OLD_PID}" ]; then
        echo "Stopping existing Zeitpolizei process (PID: ${OLD_PID})"
        kill "${OLD_PID}" 2>/dev/null
        sleep 2
    fi
    rm -f "${ZEITPOLIZEI_PID}"
fi

# Start Zeitpolizei
echo "Starting Zeitpolizei..."
nohup "${ZEITPOLIZEI_BIN}" -config "${ZEITPOLIZEI_CONFIG}" >> "${LOG_FILE}" 2>&1 &
echo $! > "${ZEITPOLIZEI_PID}"

echo "Zeitpolizei started with PID: $(cat ${ZEITPOLIZEI_PID})"
echo "Log file: ${LOG_FILE}"
echo "Web UI available at: http://$(ip route get 1 | awk '{print $7}'):8765"
