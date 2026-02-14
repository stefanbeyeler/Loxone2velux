# Global build args (must be before first FROM)
ARG BUILD_FROM=ghcr.io/home-assistant/amd64-base:3.19
ARG BUILD_ARCH=amd64

# ============================================================
# Stage 1: Clone source
# (HA builder sets context to add-on subdir, so we need to
#  clone the full repo to access Go/web source files)
# ============================================================
FROM alpine:3.19 AS source

RUN apk add --no-cache git

# Use custom arg names to avoid HA Supervisor overriding them
# (HA passes BUILD_REPOSITORY as full URL and BUILD_VERSION as addon version)
ARG SOURCE_REPO=https://github.com/stefanbeyeler/Loxone2velux.git
ARG SOURCE_BRANCH=main

RUN git clone --depth 1 --branch ${SOURCE_BRANCH} \
    "${SOURCE_REPO}" /src

# ============================================================
# Stage 2: Build frontend
# ============================================================
FROM node:20-alpine AS frontend

WORKDIR /web
COPY --from=source /src/web/package*.json ./
RUN npm install --no-audit --no-fund
COPY --from=source /src/web/ .
RUN npm run build

# ============================================================
# Stage 3: Build Go binary
# ============================================================
FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY --from=source /src/go.mod /src/go.sum ./
RUN go mod download

COPY --from=source /src/ .

# Re-declare to use within this stage
ARG BUILD_ARCH
RUN case "${BUILD_ARCH}" in \
        "amd64")   export GOARCH=amd64 ;; \
        "aarch64") export GOARCH=arm64 ;; \
        "armv7")   export GOARCH=arm && export GOARM=7 ;; \
        *)         export GOARCH=amd64 ;; \
    esac && \
    VERSION=$(cat VERSION 2>/dev/null || echo "dev") && \
    CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-w -s -X main.version=${VERSION}" -o /loxone2velux ./cmd/gateway/

# ============================================================
# Stage 4: Runtime on Home Assistant base image
# (init: false in config.yaml disables s6-overlay)
# ============================================================
FROM ${BUILD_FROM}

RUN apk add --no-cache ca-certificates tzdata jq

# Copy binary
COPY --from=builder /loxone2velux /usr/bin/loxone2velux
RUN chmod +x /usr/bin/loxone2velux

# Copy frontend
COPY --from=frontend /web/dist /app/web/dist

# Copy startup script and fix potential Windows CRLF line endings
COPY run.sh /
RUN sed -i 's/\r$//' /run.sh && chmod a+x /run.sh

# Prepare persistent config directory
RUN mkdir -p /data /config/loxone2velux

EXPOSE 8099

# Override base image's /init entrypoint (s6-overlay) since init: false
ENTRYPOINT []
CMD ["/run.sh"]
