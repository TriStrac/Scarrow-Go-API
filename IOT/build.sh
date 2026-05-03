#!/bin/bash
# build.sh - Build pi-go-service using Docker for cross-compilation

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PI_SERVICE_DIR="$SCRIPT_DIR/pi-go-service"

echo "Building pi-go-service for Linux ARM64 using Docker..."

docker run --rm \
    -v "$PI_SERVICE_DIR:/app" \
    -w /app \
    -e GOOS=linux \
    -e GOARCH=arm64 \
    -e CGO_ENABLED=1 \
    golang:alpine \
    go build -o scarrow-hub .

echo "Build complete: $PI_SERVICE_DIR/scarrow-hub"
ls -lh "$PI_SERVICE_DIR/scarrow-hub"