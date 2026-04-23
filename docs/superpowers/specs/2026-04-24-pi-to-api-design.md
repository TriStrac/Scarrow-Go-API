# Scarrow Pi-to-API Data Pipeline Design

## Overview

After provisioning, the Pi acts as a hub that:
1. Accepts BLE connections from ESP32 nodes
2. Queues node data locally
3. Sends data to API with device authentication
4. Deletes local copy on API ACK

## Architecture

```
[ESP32 Node] ---BLE---> [Pi (Hub)] ---HTTP---> [Scarrow API]
    - Detection       - Local SQLite      - Device Auth
    - Node ID         - Queue + Retry    - /device/:id/logs
```

## API Changes

### 1. New Device Auth Middleware

Create `internal/api/middlewares/device_auth.go`:

```go
// Authenticate via device_id + device_secret headers
// Headers: X-Device-ID, X-Device-Secret
func DeviceAuthMiddleware(deviceRepo repository.DeviceRepository) gin.HandlerFunc
```

- Look up device by `device_id` from DB
- Compare `device_secret` (stored on registration)
- Set `deviceID` in context for handlers
- Reject if not found or secret mismatch

### 2. Device Log Endpoint

Modify or add endpoint `POST /device/:deviceId/logs` to accept device auth:

```
Headers:
- X-Device-ID: <hub's device_id>
- X-Device-Secret: <hub's secret>

Body:
{
  "node_id": "NODE_xxx",
  "log_type": "DETECTED" | "DETERRENT",
  "pest_type": "BIRD" | "RAT",
  "frequency_hz": 0.0,
  "duration_seconds": 30,
  "payload": "{}"  // optional, any extra JSON data
}
```

Response: `201 Created` on success

## Data Flow

### Node → Pi (BLE)

When node detects something, it connects to Pi and sends:

```json
{
  "node_id": "NODE_xxx",
  "log_type": "DETECTED",
  "pest_type": "BIRD",
  "duration_seconds": 30,
  "timestamp": "2026-04-24T10:30:00Z"
}
```

Pi receives via BLE GATT characteristic write.

### Pi Local Queue

Pi stores in SQLite (`pending_logs` table):

| id | node_id | log_type | pest_type | payload | created_at | retry_count |
|----|--------|----------|-----------|---------|------------|-------------|
| uuid | NODE_xxx | DETECTED | BIRD | {...} | timestamp | 0 |

On insert: success
On API ACK (200): delete by id
On API error: increment retry_count, retry later (max 5)

### Pi → API (HTTP)

```http
POST /api/v1/device/{hub_id}/logs
X-Device-ID: {hub_id}
X-Device-Secret: {hub_secret}
Content-Type: application/json

{
  "node_id": "NODE_xxx",
  "log_type": "DETECTED",
  "pest_type": "BIRD",
  "duration_seconds": 30,
  "payload": "{}"
}
```

## Field Mode (Pi Side)

### Changes to main.go

```go
func startFieldMode() {
    // 1. Start BLE server (accept node connections)
    // 2. Start HTTP client (send to API)
    // 3. Start queue processor (retry pending logs)
}
```

### BLE Server (Node Acceptance)

- Advertise as "Scarrow_Hub_{hub_id}"
- GATT service for node data reception
- Characteristic: write-only for node payloads

### Queue Processor

- Every 10 seconds: check pending logs
- Batch send up to 10 logs per request
- Exponential backoff on failure (10s, 30s, 60s, 5min)

## Node Discovery

Two options:

### Option A: Pre-registered (Recommended)
- User registers nodes via app first
- Pi stores node_id list locally
- Only accepts known nodes

### Option B: Auto-discover
- First node connection registers automatically via API
- Requires user approval flow
- More complex

**Recommendation**: Option A - nodes registered via app, Pi syncs list on startup.

## Security Considerations

- Device secret: 32-char random, stored hashed? (or plain in DB for comparison)
- HTTPS only for API communication
- BLE pairing? (for now: open BLE, trust within network)

## Implementation Plan

1. **API**: Add device auth middleware
2. **API**: Create/update log endpoint to accept device auth
3. **Pi**: Implement BLE server for node data reception
4. **Pi**: Implement local SQLite queue
5. **Pi**: Implement HTTP client with device auth
6. **Pi**: Implement queue processor with retry
7. **Testing**: End-to-end node → API data flow

## API Reference

### Device Auth Middleware

```
GET /health (no auth)
POST /device/{deviceId}/logs (device auth)
```

### Request/Response

**Node Log Request**:
```json
{
  "node_id": "NODE_ABC123",
  "log_type": "DETECTED",
  "pest_type": "BIRD",
  "frequency_hz": 0.0,
  "duration_seconds": 30,
  "payload": "{\"range_cm\": 45}"
}
```

**Response 201**:
```json
{
  "message": "Log created successfully"
}
```

**Response 401**: Invalid device credentials
**Response 403**: Device not authorized
**Response 404**: Device not found