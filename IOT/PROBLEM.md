# Scarrow ESP32 Node - Current Problem Documentation

## Project Goal
Connect 2x LD2420 mmWave sensors to ESP32-S3 for pest detection in corn fields:
- **Radar 1 (Bird)**: Top mount - needs distance + motion data
- **Radar 2 (Rat)**: Bottom mount - needs motion + distance data

## ESP32-S3 UART Information

ESP32-S3 has 3 UARTs:
| UART | Default TX | Default RX | Status |
|------|-----------|-----------|----------|--------|
| UART0 | GPIO 43 | GPIO 44 | Fixed (USB) |
| UART1 | GPIO 17 | GPIO 18 | Works on custom pins |
| UART2 | None | None | No default pins |

## Current Hardware Wiring

| Component | ESP32 GPIO | LD2420 Pin | Serial |
|----------|-----------|-----------|-----------|--------|
| Radar 1 | GPIO 16 | OT1 | Digital input |
| Radar 1 | GPIO 17 | TX | Serial1 (WORKING) |
| Radar 2 | GPIO 4 | OT1 | Digital input |
| Radar 2 | GPIO 5 | TX | SoftwareSerial (CRASHES) |

## Problem

### Radar 1: Working ✅
- Serial1 on GPIO 16/17 works perfectly
- Outputs: `ON`, `OFF`, `Range 45 cm`
- Config commands respond with ACK

### Radar 2: NOT Working ❌
- **SoftwareSerial** on GPIO 5 crashes the ESP32
- Error: `Guru Meditation Error: Core 1 panic'ed (Interrupt wdt timeout on CPU1)`
- The ESP32 freezes within seconds when SoftwareSerial is enabled

### Attempted Solutions
1. **Serial2 on GPIO 9/10** - TIMEOUT
2. **SoftwareSerial on GPIO 2/3** - crashes
3. **SoftwareSerial on GPIO 5** - crashes
4. **HardwareSerial on GPIO 4/5** - never tried (Serial2 has no default)

## What Works

### OT1 Motion Detection ✅
Both sensors' OT1 pins work correctly:
- `[R1 OT1] HIGH` when motion detected
- `[R2 OT1] HIGH` when motion detected

## Known Issue with SoftwareSerial on ESP32-S3

SoftwareSerial library on ESP32-S3 can cause:
- Interrupt watchdog timeout
- Core panics
- Inconsistent behavior

Alternative: Use HardwareSerial only, or use I2C/SPI sensors instead.

## Current Code Pin Definitions

```cpp
// Radar 1 (Working)
#define RADAR1_OT1 16   // OT1 motion detection
#define RADAR1_RX 17    // Serial data from LD2420

// Radar 2 (NOT Working)
#define RADAR2_OT1 4    // OT1 motion detection
#define RADAR2_RX 5       // Serial data from LD2420
```

## Working Single Sensor Code (Reference)

```cpp
Serial1.begin(115200, SERIAL_8N1, RADAR_RX, RADAR_TX);
uint8_t openCfg[]  = {0xFD, 0xFC, 0xFB, 0xFA, 0x04, 0x00, 0xFF, 0x00, 0x01, 0x00, 0x04, 0x03, 0x02, 0x01};
uint8_t setRunMode[] = {0xFD, 0xFC, 0xFB, 0xFA, 0x08, 0x00, 0x12, 0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x04, 0x03, 0x02, 0x01};
uint8_t closeCfg[] = {0xFD, 0xFC, 0xFB, 0xFA, 0x02, 0x00, 0xFE, 0x00, 0x04, 0x03, 0x02, 0x01};
```

## Options to Solve

1. **Use HardwareSerial for Radar 2** - Try UART2 on different pins
2. **Use I2C sensor instead of LD2420** - More reliable
3. **Use OT1 only for Radar 2** - Motion detection only (simpler)
4. **Use ESP32 (not S3)** - Better UART support
5. **Research ESP32-S3 UART2 pins** - Find working combination

---

## Prompt for New Agent

---

You are helping TriStrac with a Scarrow IoT project. Read `D:\Codes\Scarrow-Go-API\IOT\PROBLEM.md` for full context.

**Current Problem:** Need to connect 2x LD2420 mmWave sensors to ESP32-S3. Radar 1 works on Serial1 (GPIO 16/17), but Radar 2 serial communication is not working. SoftwareSerial crashes the ESP32.

**Goal:** Get both LD2420 sensors working with distance + motion output.

**Your Task:** 
1. Read PROBLEM.md for full details
2. Read the current code at `D:\Codes\Scarrow-Go-API\IOT\esp32-node\esp32-node.ino`
3. Suggest a working solution to get Radar 2 serial communication working
4. Key constraints:
   - ESP32-S3 has limited UARTs (UART2 has no default pins)
   - SoftwareSerial causes crashes on ESP32-S3
   - LD2420 uses TTL serial at 115200 baud
   - User wants distance data from both sensors

**Working solution options to explore:**
1. Find correct HardwareSerial pins for UART2 on ESP32-S3
2. Use I2C interface instead of UART
3. Use OT1 only for Radar 2 (motion only)
4. Or suggest a completely different approach

Start fresh and don't assume previous attempts. Focus on finding a working solution.

---

**Last Updated:** 2026-04-21
**Status:** Stuck on Radar 2 serial communication