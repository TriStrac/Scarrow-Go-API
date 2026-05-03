# Scarrow Go API - Project Status

**Last Updated:** 2026-04-26
**Current Phase:** WebSocket Command System Implementation - Complete

---

## Project Overview

Two services:
- **API** (Scarrow-Go-API): Go API server running in Docker, handles user auth and device commands
- **Pi Service** (pi-go-service): Go service running on Raspberry Pi, manages BLE and device field operations

---

## Current Implementation

### Command System (WebSocket-based)
Replaced MQTT with WebSocket for sending commands from API to Pi.

**Flow:**
1. User sends `POST /api/hubs/{hubId}/commands` with auth or `X-Dev-Bypass: dev_bypass_secret_123`
2. API's WebSocket client sends to `wss://mqtt.striel.xyz:443/ws`
3. Pi's WebSocket client receives command and executes
4. Pi sends result back via WebSocket

**Available Commands:**
- `reboot` - Runs `sudo reboot` (needs root perms)
- `wifi` - Runs `nmcli` to connect to WiFi
- `reset` - Resets JSON config and exits (service auto-restarts via systemd)

---

## Key Files

### API (D:/Codes/Scarrow-Go-API/)
| File | Purpose |
|------|---------|
| `main.go` | Entry point, Gin router setup |
| `internal/api/controllers/device.go` | `SendCommand` endpoint (`/api/hubs/:hubId/commands`) |
| `internal/ws/client.go` | WebSocket client for API |
| `internal/api/middlewares/auth.go` | Auth middleware + dev bypass |
| `internal/db/db.go` | JSON config storage |
| `.env` | Contains `DEV_BYPASS_SECRET=dev_bypass_secret_123` |
| `docker-compose.yml` | API container definition |
| `test_cmd.go` | Test script for sending commands |

### Pi Service (D:/Codes/Scarrow-Go-API/IOT/pi-go-service/)
| File | Purpose |
|------|---------|
| `main.go` | Entry point, field/setup mode logic |
| `internal/ws/client.go` | WebSocket client for Pi |
| `internal/commands/handler.go` | Command execution (reboot, wifi, reset) |
| `internal/db/db.go` | JSON config storage |
| `scrow.json` | Test config (contains `central_device_id: PI-001`) |

---

## Configuration Files

### Pi JSON Config (`scrow.json`)
```json
{
  "central_device_id": "PI-001",
  "central_device_secret": "test_secret_123",
  "skip_field_mode": true,
  "central_device_ble_advertise_name": "Scarrow_Central_Device_Setup"
}
```

### API Environment (.env)
```
DEV_BYPASS_SECRET=dev_bypass_secret_123
```

---

## Testing Checklist

- [x] Pi connects to WebSocket and registers as `PI-001`
- [x] API sends command via WebSocket
- [x] Pi receives command
- [x] Dev bypass works (`X-Dev-Bypass: dev_bypass_secret_123`)
- [x] Pi executes command (wifi, reset, reboot)
- [x] Async response sent back via WebSocket

---

## Deployment Info

- **API:** Docker container at `scarrow-api.striel.xyz`
- **Pi:** SSH at `192.168.50.216`, service at `/home/tristrac/scarrow/scarrow-hub`
- **WebSocket Server:** `wss://mqtt.striel.xyz:443/ws`

### Restart Pi Service (as root)
```bash
sshpass -p "OuroKronii314-" ssh -o StrictHostKeyChecking=no tristrac@192.168.50.216
sudo pkill -9 -f scarrow-hub
cd /home/tristrac/scarrow && sudo ./scarrow-hub &
```

### Test Command (from API container or D:/Codes/Scarrow-Go-API/)
```bash
go run test_cmd.go
```

---

## Known Issues / TODO

1. **Reboot command may not fully execute** - `sudo reboot` runs but SSH connection may block completion. Need to verify if `nohup` or different approach required.

2. **Delete scrow.json** before production deployment - it's the test config that enables `skip_field_mode`.

---

## Next Steps (for fresh instance)

1. Review this document
2. Rebuild Pi service binary if needed: `cd IOT/pi-go-service && go build -o scarrow-hub .`
3. Deploy to Pi and test commands (reboot, wifi, reset)
4. Delete `scrow.json` from Pi to test real provisioning flow
5. Test full BLE provisioning via mobile app