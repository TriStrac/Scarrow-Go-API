# Scarrow IoT Architecture Plan

This document tracks the step-by-step progress of the physical hardware integration between the Raspberry Pi 4 (Central Hub) and the ESP32-S3 (Edge Nodes).

## Architecture Overview
- **ESP32-S3 (Node):** Senses mmWave radar, triggers MOSFET for high-power Horn/Ultrasonic, and publishes telemetry via MQTT.
- **Raspberry Pi 4 (Central):** Runs `Mosquitto` Broker (local Wi-Fi) + `Scarrow-Bridge` (Go Service). Buffers logs offline and HTTP POSTs to the main `Scarrow-Go-API`.
- **Scarrow-Go-API (Cloud/PC):** Receives structured logs, updates the database, and fires push notifications to the Mobile App.

---

## 🛠️ Phase 1: Hardware Assembly & Power Prototyping
- [ ] Diagram the MOSFET switching circuit (IRLZ44N -> XL6019 -> TPA3110 -> Horn Speaker) so the ESP32 doesn't fry its pins.
- [x] Connect the HLK-LD2420 mmWave sensor to hardware serial pins on the ESP32-S3.
- [ ] Ensure the CN3791 MPPT Solar charger safely charges the 3.7V battery.

## 📡 Phase 2: Raspberry Pi 4 Network & MQTT Setup
- [ ] Configure Pi's Wi-Fi interface to broadcast an Access Point (SSID: `Scarrow_Hub_XXXX`) while remaining connected to the internet (if applicable).
- [ ] Install and configure `Mosquitto` to run locally on the Pi.
- [ ] Test local MQTT publish/subscribe from a laptop to the Pi.

## 💻 Phase 3: The `pi-go-service` (Scarrow Bridge)
- [ ] **MQTT Subscriber:** Write a Go script that subscribes to `scarrow/nodes/#`.
- [ ] **Offline Buffering:** Implement a local JSON or SQLite buffer that temporarily holds MQTT payloads if the `Scarrow-Go-API` is unreachable.
- [ ] **HTTP Forwarder:** Periodically batch HTTP POST `log_type: TELEMETRY` and `pest_type: LOCUST/RATS/BIRDS` to the API.
- [ ] **Device Registration Flow:** Handle the initial Bearer token saving on the Pi.

## ⚡ Phase 4: ESP32-S3 Firmware (`esp32-node`)
- [x] **Sensor Reading:** C++ code to poll HLK-LD2420 mmWave sensor over Serial (ASCII Mode verified).
- [ ] **Audio Deterrence:** PWM signaling or MOSFET toggling when mmWave detects motion.
- [ ] **Wi-Fi Manager (Pairing):** Implement `WiFiManager` or custom SoftAP so the mobile app can pass the Pi's Wi-Fi credentials to the ESP32.
- [ ] **MQTT Publisher:** Send telemetry (e.g., "LOCUST detected, emitted 15000Hz for 10s") securely over local Wi-Fi to the Pi.

## 📱 Phase 5: Mobile App Provisioning
- [ ] Bluetooth LE or Wi-Fi SoftAP scanning to register the Hub and Nodes.
- [ ] Send API authentication tokens to the Pi.

---
*Progress Tracking:*
- [x] Initial hardware review & feasibility check.
- [x] Directory structure created.
- [x] LD2420 mmWave Sensor integration & ASCII protocol parsing.
- [ ] (Next step: Phase 4 - Audio Deterrence & MOSFET integration)
