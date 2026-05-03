#!/bin/bash
# install.sh - Deploy and start pi-go-service on Raspberry Pi

SERVICE_NAME="scarrow-hub"
BINARY_NAME="scarrow-hub"
INSTALL_DIR="/home/tristrac/scarrow"
CURRENT_DIR=$(pwd)

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Starting scarrow-hub Install...${NC}"

if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root (sudo ./install.sh)${NC}"
    exit 1
fi

# Stop old service
pkill scarrow-hub 2>/dev/null || true
sleep 1

# Install binary
echo "Installing binary..."
if [ -f "$CURRENT_DIR/$BINARY_NAME" ]; then
    cp "$CURRENT_DIR/$BINARY_NAME" "$INSTALL_DIR/"
    chmod 755 "$INSTALL_DIR/$BINARY_NAME"
else
    echo -e "${RED}Error: '$BINARY_NAME' binary not found.${NC}"
    exit 1
fi

# Kill any running process and start fresh
pkill scarrow-hub 2>/dev/null || true
sleep 1

cd "$INSTALL_DIR"
nohup ./scarrow-hub > /tmp/scarrow-hub.log 2>&1 &

sleep 3

if pgrep -x scarrow-hub > /dev/null; then
    echo -e "${GREEN}✅ scarrow-hub Installed and Running!${NC}"
    echo "Logs: tail -f /tmp/scarrow-hub.log"
else
    echo -e "${RED}❌ Failed to start. Check logs:${NC}"
    cat /tmp/scarrow-hub.log
fi