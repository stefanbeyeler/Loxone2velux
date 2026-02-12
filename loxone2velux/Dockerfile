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

ARG BUILD_REPOSITORY=stefanbeyeler/Loxone2velux
ARG BUILD_VERSION=main

RUN git clone --depth 1 --branch ${BUILD_VERSION} \
    "https://github.com/${BUILD_REPOSITORY}.git" /src

# ============================================================
# Stage 2: Build frontend
# ============================================================
FROM node:20-alpine AS frontend

WORKDIR /web
COPY --from=source /src/web/package*.json ./
RUN npm ci --no-audit --no-fund
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
    CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-w -s" -o /loxone2velux ./cmd/gateway/

# ============================================================
# Stage 4: Runtime on Home Assistant base image
# ============================================================
FROM ${BUILD_FROM}

RUN apk add --no-cache ca-certificates tzdata

# Copy binary
COPY --from=builder /loxone2velux /usr/bin/loxone2velux
RUN chmod +x /usr/bin/loxone2velux

# Copy frontend
COPY --from=frontend /web/dist /app/web/dist

# Copy s6 service definitions
COPY rootfs /
