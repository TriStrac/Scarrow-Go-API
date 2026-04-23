# Scarrow ESP32 Node Firmware

## Current Hardware Setup

### ESP32-S3 Pin Connections

| Sensor | Purpose | ESP32 GPIO | LD2420 Pin | Serial |
|--------|---------|----------|-----------|--------|
| Radar 1 | Bird detection (top) | GPIO 16 | OT1 | Serial1 |
| Radar 1 | Data | GPIO 17 | TX | Serial1 |
| Radar 2 | Rat detection (bottom) | GPIO 4 | OT1 | ❌ |
| Radar 2 | Data | GPIO 5 | TX | ❌ |

### Working Components
- ✅ Radar 1 (Serial1 on GPIO 16/17) - Distance output working
- ✅ OT1 motion detection on both sensors
- ✅ Deterrence (MOSFET + tone)
- ⚠️ Radar 2 Serial - NOT working (timeout)

## Known Issues

### Radar 2 Serial Communication
- ESP32-S3 UART2 has no default pins
- Tried GPIO 9/10 - timeout
- Tried GPIO 2/3 (SoftwareSerial) - not tested yet (user kept GPIO 4/5)
- Config commands always timeout

### Current Wiring (User's hardware)
```
Radar 2:
- OT1 -> GPIO 4
- TX -> GPIO 5
```

## Code Status

### Working Single-Sensor Code
```cpp
#define RADAR_RX 16 
#define RADAR_TX 17 
Serial1.begin(115200, SERIAL_8N1, RADAR_RX, RADAR_TX);
```
This works perfectly - outputs "Range X cm"

### Multi-Sensor Code Status
- Radar 1: Working ✅
- Radar 2: Config commands timeout ❌

## Next Steps

1. **Quick fix**: Use OT1 only for Radar 2 (motion detection, no distance)
2. **Research**: Find working Serial2 pins for ESP32-S3
3. **Alternative**: Use different ESP32 board with better UART availability

## Board Notes

- ESP32-S3 has 3 UARTs but:
  - UART0: Fixed on GPIO 43/44 (USB)
  - UART1: Default GPIO 17/18, but works on custom pins
  - UART2: No default pins, needs research

## Commands Used for LD2420

```cpp
uint8_t openCfg[]  = {0xFD, 0xFC, 0xFB, 0xFA, 0x04, 0x00, 0xFF, 0x00, 0x01, 0x00, 0x04, 0x03, 0x02, 0x01};
uint8_t setRunMode[] = {0xFD, 0xFC, 0xFB, 0xFA, 0x08, 0x00, 0x12, 0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x04, 0x03, 0x02, 0x01};
uint8_t closeCfg[] = {0xFD, 0xFC, 0xFB, 0xFA, 0x02, 0x00, 0xFE, 0x00, 0x04, 0x03, 0x02, 0x01};
```

## Quick Test

User's original single-sensor working code is preserved in case multi-sensor breaks.