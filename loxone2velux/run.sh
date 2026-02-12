#!/usr/bin/with-contenv bashio
# ==============================================================================
# Loxone2Velux Gateway - Startup script
# ==============================================================================

# Read add-on options
KLF200_HOST=$(bashio::config 'klf200_host')
KLF200_PASSWORD=$(bashio::config 'klf200_password')
KLF200_PORT=$(bashio::config 'klf200_port')
RECONNECT_INTERVAL=$(bashio::config 'reconnect_interval')
REFRESH_INTERVAL=$(bashio::config 'refresh_interval')
LOG_LEVEL=$(bashio::config 'log_level')
API_TOKEN=$(bashio::config 'api_token')

# Get ingress port
INGRESS_PORT=$(bashio::addon.ingress_port)
LISTEN_PORT=${INGRESS_PORT:-8099}

bashio::log.info "Starting Loxone2Velux Gateway..."
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
cd /app

# Start the gateway
exec /usr/bin/loxone2velux -config "${CONFIG_FILE}"
