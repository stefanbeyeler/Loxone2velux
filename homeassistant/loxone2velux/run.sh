#!/bin/sh
# ==============================================================================
# Loxone2Velux Gateway - Startup script
# Reads HA add-on options from /data/options.json via jq
# ==============================================================================
set -e

echo "=== Loxone2Velux Gateway starting ==="

# Read add-on options from Home Assistant
OPTIONS_FILE="/data/options.json"

if [ -f "$OPTIONS_FILE" ]; then
    echo "Reading options from ${OPTIONS_FILE}"
    KLF200_HOST=$(jq -r '.klf200_host // ""' "$OPTIONS_FILE")
    KLF200_PASSWORD=$(jq -r '.klf200_password // ""' "$OPTIONS_FILE")
    KLF200_PORT=$(jq -r '.klf200_port // 51200' "$OPTIONS_FILE")
    RECONNECT_INTERVAL=$(jq -r '.reconnect_interval // 30' "$OPTIONS_FILE")
    REFRESH_INTERVAL=$(jq -r '.refresh_interval // 300' "$OPTIONS_FILE")
    LOG_LEVEL=$(jq -r '.log_level // "info"' "$OPTIONS_FILE")
    API_TOKEN=$(jq -r '.api_token // ""' "$OPTIONS_FILE")
else
    echo "WARNING: Options file not found at ${OPTIONS_FILE}, using defaults"
    KLF200_HOST=""
    KLF200_PASSWORD=""
    KLF200_PORT=51200
    RECONNECT_INTERVAL=30
    REFRESH_INTERVAL=300
    LOG_LEVEL="info"
    API_TOKEN=""
fi

LISTEN_PORT=8099

echo "KLF-200 host: ${KLF200_HOST}:${KLF200_PORT}"
echo "Listening on port: ${LISTEN_PORT}"

# Generate config.yaml for the Go binary
CONFIG_FILE="/data/config.yaml"

cat > "${CONFIG_FILE}" << EOF
klf200:
  host: "${KLF200_HOST}"
  port: ${KLF200_PORT}
  password: "${KLF200_PASSWORD}"
  reconnect_interval: ${RECONNECT_INTERVAL}s
  refresh_interval: ${REFRESH_INTERVAL}s

server:
  host: "0.0.0.0"
  port: ${LISTEN_PORT}
  read_timeout: 15s
  write_timeout: 15s
  api_token: "${API_TOKEN}"

logging:
  level: "${LOG_LEVEL}"
  format: "console"
EOF

echo "Configuration written to ${CONFIG_FILE}"

# Change to /app so the binary can find web/dist
cd /app || true

echo "Starting binary..."

# Start the gateway (exec replaces shell with Go binary)
exec /usr/bin/loxone2velux -config "${CONFIG_FILE}"
