#!/bin/bash
# deploy.sh - Upload and deploy pi-go-service to Raspberry Pi

set -e

PI_HOST="192.168.50.216"
PI_USER="tristrac"
PI_PASS="OuroKronii314-"
PI_PATH="/home/tristrac/scarrow"

BINARY="$(dirname "$0")/pi-go-service/scarrow-hub"

echo "Deploying to Pi at $PI_HOST..."

# Kill existing process
sshpass -p "$PI_PASS" ssh -o StrictHostKeyChecking=no "$PI_USER@$PI_HOST" "pkill scarrow-hub 2>/dev/null || true"

# Upload binary
echo "Uploading binary..."
sshpass -p "$PI_PASS" scp -o StrictHostKeyChecking=no "$BINARY" "$PI_USER@$PI_HOST:$PI_PATH/scarrow-hub"

echo "Binary deployed successfully!"