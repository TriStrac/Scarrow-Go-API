# Scarrow ESP32 Node Firmware

## Current Hardware Setup

### ESP32-S3 Pin Connections (AS WIRED)

| Sensor | Purpose | ESP32 GPIO | LD2420 Pin | Status |
|--------|---------|------------|------------|--------|
| Radar 1 | Bird detection | GPIO 16 | OT1 | ✅ Working |
| Radar 1 | Data (TX) | GPIO 17 | TX | ✅ Working |
| Radar 2 | Rat detection | GPIO 4 | OT1 | ✅ Working |
| Radar 2 | Data (TX) | GPIO 5 | TX | ⚠️ NOT working |

### Serial Configuration
- **Radar 1**: Serial1 on GPIO 16/17 - WORKS perfectly
- **Radar 2**: Trying GPIO 5 for data - NOT responding

## Working Components
- ✅ Radar 1 distance output ("Range X cm")
- ✅ OT1 motion detection on both sensors
- ✅ Deterrence (MOSFET + tone)
- ❌ Radar 2 serial communication (config times out)

## Known Issue

Radar 2's TX pin (GPIO 5) is not responding to LD2420 config commands. The sensor's OT1 output works (motion detection), but serial data is not being sent.

**Tried:**
- GPIO 9/10 (Serial2) - timeout
- Still need to try: GPIO 2/3 (SoftwareSerial)

## Code Pin Definitions

```cpp
// Radar 1 - Working
#define RADAR1_RX 16 
#define RADAR1_TX 17 

// Radar 2 - Current wiring (NOT working via serial)
#define RADAR2_RX 5   // TX from LD2420 connects here
#define RADAR2_OT1 4  // OT1 motion detection
```

## User's Next Test

Test with **SoftwareSerial** on **GPIO 2/3**:
- Wire Radar 2 TX to GPIO 2
- Wire Radar 2 OT1 to GPIO 3 (or keep on GPIO 4)

## Quick Fix Option

If serial never works for Radar 2, use **OT1 only**:
- Radar 1: Full distance + motion
- Radar 2: Motion only (OT1 on GPIO 4)

---

**Last Updated:** 2026-04-21
**Status:** Debugging second LD2420 serial communication