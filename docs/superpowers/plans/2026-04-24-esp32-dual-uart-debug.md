# ESP32-S3 Dual LD2420 - Debugging & Fix Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Get `esp32-node.ino` to boot without crashing on CPU1 when initializing dual UARTs (UART1 + UART2).

**Architecture:** Use IDF-style interrupt masking around `UART2.begin()` — disable UART2 interrupts before calling `begin()`, re-enable after — to prevent CPU1 from watchdog-timeout during ISR registration. Single file, single task.

**Tech Stack:** Arduino-ESP32 core, ESP-IDF UART driver, Arduino framework

---

## Root Cause (Evidence)

| Evidence | Finding |
|----------|---------|
| Backtrace: `uart_hal_write_txfifo` + `uartEnableInterrupt` + `uartBegin` in ISR path | Crash is inside UART ISR initialization, not user code |
| `Core 1 panic'ed (Interrupt wdt timeout on CPU1)` | CPU1 (core that runs ISR) is stuck in interrupt handler |
| Radar 1 (UART1) alone works fine | UART1 initialization is fine; UART2 is the trigger |
| Identical crash in [esp32_can#31](https://github.com/collin80/esp32_can/issues/31) | Known issue: 2nd UART init causes CPU1 watchdog timeout |
| Happens during `setup()` before any loop iteration | ISR fires before it is fully registered |

**Hypothesis:** When `UART2.begin()` attaches the UART ISR, a spurious or pending interrupt fires before the ISR is fully installed. CPU1 handles this in an incomplete ISR context, gets stuck, and the interrupt watchdog triggers. The fix is to mask UART2 interrupts before `begin()` and unmask after.

---

## Task 1: Fix UART2 Init with Interrupt Masking

**Files:**
- Modify: `D:\Codes\Scarrow-Go-API\IOT\esp32-node\esp32-node.ino`

**Step 1: Add ESP-IDF UART include and UART2 mask helper**

After line 6 (`#include <ArduinoJson.h>`), add:
```cpp
#include <driver/uart.h>
```

**Step 2: Replace the UART1/UART2 init block in `setup()` with interrupt-masked initialization**

Replace lines 164-169:
```cpp
  UART1.begin(115200, SERIAL_8N1, 17, 16);
  delay(100);
  UART2.begin(115200, SERIAL_8N1, RADAR2_RX, RADAR2_TX);
  delay(100);
```

With:
```cpp
  UART1.begin(115200, SERIAL_8N1, 17, 16);
  delay(100);
  uart_set_irq_en(UART_NUM_2, UART_INTR_DISABLE);
  UART2.begin(115200, SERIAL_8N1, RADAR2_RX, RADAR2_TX);
  uart_set_irq_en(UART_NUM_2, UART_INTR_FLAG_DEF_CONFIG);
  delay(100);
```

This uses `uart_set_irq_en()` with `UART_INTR_DISABLE` before `begin()` to suppress all UART2 interrupts, preventing CPU1 from being interrupted by a partially-installed ISR. After `begin()` registers the ISR, `UART_INTR_FLAG_DEF_CONFIG` re-enables the default interrupt configuration.

**Step 3: Commit**

```bash
git add esp32-node/esp32-node.ino
git commit -m "fix(iot): mask UART2 interrupts during begin() to prevent CPU1 watchdog timeout"
```

---

## Expected Result

After this fix:
- ESP32 boots cleanly to `--- Radars Active ---`
- No `Guru Meditation Error` or `Interrupt wdt timeout`
- `[R1] Range XX` and `[R2] Range XX` appear in Serial Monitor

## If This Fix Doesn't Work

If the crash persists after Task 1:
1. Try adding `gpio_reset_pin()` calls before `UART2.begin()`:
```cpp
gpio_reset_pin((gpio_num_t)RADAR2_RX);
gpio_reset_pin((gpio_num_t)RADAR2_TX);
```
2. Try initializing UART2 on different GPIO pairs (e.g., GPIO 14/15 or GPIO 1/2) to rule out pin-specific issues
3. Check if BLE initialization is interfering — try moving BLE setup after UART init