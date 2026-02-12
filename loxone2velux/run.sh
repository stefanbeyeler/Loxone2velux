#!/usr/bin/with-contenv bashio
# ==============================================================================
# Loxone2Velux Gateway - Startup script
# ==============================================================================

bashio::log.info "=== Loxone2Velux Gateway starting ==="

# Read add-on options
KLF200_HOST=$(bashio::config 'klf200_host' 2>/dev/null || echo "")
KLF200_PASSWORD=$(bashio::config 'klf200_password' 2>/dev/null || echo "")
KLF200_PORT=$(bashio::config 'klf200_port' 2>/dev/null || echo "51200")
RECONNECT_INTERVAL=$(bashio::config 'reconnect_interval' 2>/dev/null || echo "30")
REFRESH_INTERVAL=$(bashio::config 'refresh_interval' 2>/dev/null || echo "300")
LOG_LEVEL=$(bashio::config 'log_level' 2>/dev/null || echo "info")
API_TOKEN=$(bashio::config 'api_token' 2>/dev/null || echo "")

# Get ingress port
INGRESS_PORT=$(bashio::addon.ingress_port 2>/dev/null || echo "8099")
LISTEN_PORT=${INGRESS_PORT:-8099}

bashio::log.info "KLF-200 host: ${KLF200_HOST}:${KLF200_PORT}"
bashio::log.info "Listening on port: ${LISTEN_PORT}"

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

bashio::log.info "Configuration written to ${CONFIG_FILE}"

# Change to /app so the binary can find web/dist
cd /app || true

bashio::log.info "Starting binary..."

# Start the gateway
exec /usr/bin/loxone2velux -config "${CONFIG_FILE}"
