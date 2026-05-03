# WebSocket Connection Persistence Plan

**Problem:** WebSocket connections die after ~3 minutes of idle time. Commands fail with "broken pipe" errors even after reconnect attempts.

**Root Cause Analysis:**
1. Cloudflare/CDN proxy has idle timeout (typically 60-300 seconds)
2. Neither API nor Pi service sends keepalive pings
3. When connection drops, write operations fail with "broken pipe"
4. Reconnect logic doesn't properly reset connection state before reconnection

---

## Research Findings

### Gorilla WebSocket API (authoritative)
- **Ping/Pong Control:** Connections handle ping messages by calling the handler set with `SetPingHandler`. Default handler sends pong automatically.
- **Read Deadline:** `SetReadDeadline(t)` - After read times out, connection is corrupted. Zero = no timeout.
- **Concurrency:** Supports one concurrent reader and one concurrent writer.
- **Application Must Read:** "The application must read the connection to process close, ping, and pong messages sent from the peer."

### Cloudflare WebSocket Behavior
- Idle timeout varies (60-300 seconds depending on plan)
- Connection drop triggers TCP FIN, causing "broken pipe" on write
- Ping/pong frames must be exchanged to reset idle timer

---

## Implementation Plan

### Phase 1: Fix Ping Keepalive (Both Sides)
**Problem:** Manual ping implementation doesn't properly set pong handler or reset idle timeout.

**Changes:**
1. Set `conn.SetPingHandler()` to handle incoming pings (auto pong)
2. Set `conn.SetPongHandler()` to reset read deadline on pong received
3. In pingLoop: reset read deadline BEFORE writing ping, not after
4. Add logging to verify pings are actually being sent

### Phase 2: Fix Reconnection Logic (Both Sides)
**Problem:** Reconnect may leave old connection in bad state. Mutex not properly protecting.

**Changes:**
1. Before reconnect: fully close and nil the old connection
2. Add small delay (1-2s) between close and reconnect
3. Log when reconnect starts and completes
4. Ensure mutex is locked during entire reconnect sequence

### Phase 3: Add Write Protection (API Side)
**Problem:** `SendCommand` may be called while connection is being recreated.

**Changes:**
1. Add `IsConnected()` check in `SendCommand` before writing
2. If not connected, return error immediately (don't try to write to bad conn)
3. Let API endpoint return 503 Service Unavailable

### Phase 4: Add Connection Health Monitoring
**Changes:**
1. Track last successful read/write timestamp
2. If no activity for 2 minutes, force reconnect
3. Log connection health every 60 seconds

---

## Files to Modify

### API (D:/Codes/Scarrow-Go-API/)
- `internal/ws/client.go` - Phases 1, 2, 3, 4

### Pi Service (D:/Codes/Scarrow-Go-API/IOT/pi-go-service/)
- `internal/ws/client.go` - Phases 1, 2

---

## Test Plan

1. Start API and Pi services
2. Wait 5+ minutes without commands
3. Send command via Postman (not test_cmd.go)
4. Verify:
   - Command succeeds
   - No "broken pipe" errors
   - Pi logs show ping every 60 seconds
   - Reconnect happens smoothly if connection drops