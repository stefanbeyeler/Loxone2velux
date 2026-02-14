#!/bin/sh
# ==============================================================================
# Loxone2Velux Gateway - Startup script
# Reads HA add-on options from /data/options.json via jq
# Persistent config at /config/loxone2velux/config.yaml
# ==============================================================================

echo "=== Loxone2Velux Gateway starting ==="

# Read add-on options from Home Assistant
OPTIONS_FILE="/data/options.json"
PERSISTENT_CONFIG="/config/loxone2velux/config.yaml"
LISTEN_PORT=8099

if [ -f "$OPTIONS_FILE" ]; then
    echo "Reading options from ${OPTIONS_FILE}"
    KLF200_HOST=$(jq -r '.klf200_host // ""' "$OPTIONS_FILE" 2>/dev/null || echo "")
    KLF200_PASSWORD=$(jq -r '.klf200_password // ""' "$OPTIONS_FILE" 2>/dev/null || echo "")
    KLF200_PORT=$(jq -r '.klf200_port // 51200' "$OPTIONS_FILE" 2>/dev/null || echo "51200")
    RECONNECT_INTERVAL=$(jq -r '.reconnect_interval // 30' "$OPTIONS_FILE" 2>/dev/null || echo "30")
    REFRESH_INTERVAL=$(jq -r '.refresh_interval // 300' "$OPTIONS_FILE" 2>/dev/null || echo "300")
    LOG_LEVEL=$(jq -r '.log_level // "info"' "$OPTIONS_FILE" 2>/dev/null || echo "info")
    API_TOKEN=$(jq -r '.api_token // ""' "$OPTIONS_FILE" 2>/dev/null || echo "")
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

echo "KLF-200 host: ${KLF200_HOST}:${KLF200_PORT}"
echo "Listening on port: ${LISTEN_PORT}"

# Create or update persistent config
mkdir -p /config/loxone2velux

if [ ! -f "${PERSISTENT_CONFIG}" ]; then
    echo "First run - creating config at ${PERSISTENT_CONFIG}"
    cat > "${PERSISTENT_CONFIG}" << EOF
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
else
    echo "Updating existing config at ${PERSISTENT_CONFIG}"
    # Only update values from HA options (preserve any config changes made via Web UI)
    if [ -n "${KLF200_HOST}" ]; then
        sed -i "s|^\(  host: \).*|\\1\"${KLF200_HOST}\"|" "${PERSISTENT_CONFIG}"
    fi
    if [ -n "${KLF200_PASSWORD}" ]; then
        sed -i "s|^\(  password: \).*|\\1\"${KLF200_PASSWORD}\"|" "${PERSISTENT_CONFIG}"
    fi
    if [ -n "${KLF200_PORT}" ]; then
        sed -i "s|^\(  port: \)[0-9]*|\\1${KLF200_PORT}|" "${PERSISTENT_CONFIG}"
    fi
    if [ -n "${LOG_LEVEL}" ]; then
        sed -i "s|^\(  level: \).*|\\1\"${LOG_LEVEL}\"|" "${PERSISTENT_CONFIG}"
    fi
fi

echo "Configuration ready at ${PERSISTENT_CONFIG}"

# Change to /app so the binary can find web/dist
cd /app || true

echo "Starting binary..."

# Start the gateway (exec replaces shell with Go binary)
exec /usr/bin/loxone2velux -config "${PERSISTENT_CONFIG}"
