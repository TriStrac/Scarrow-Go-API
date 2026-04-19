# Scarrow IoT Architecture Plan (Field Deployment Edition)

This document tracks the integration between the Raspberry Pi 4 (Local Hub) and the ESP32-S3 (Edge Nodes) optimized for off-grid/field use.

## 📦 Project Components
### Central Hub (Raspberry Pi 4)
- **Networking:** Acts as a **Wi-Fi Access Point** (Hotspot) for nodes.
- **Broker:** Local MQTT Broker (Mosquitto).
- **Service:** `pi-go-service` for local logic and data buffering.
- **Storage:** Local SQLite database for offline logging.

### Edge Node (ESP32-S3)
- **Power:** 3.7V Li-Ion (Parallel config) + Solar.
- **Connectivity:** Connects directly to the Pi Hub's Wi-Fi.
- **Deterrence:** 12V Switched Rail (XL6019 Boost + MOSFET) for Speaker/Ultrasonic.

---

## 🛠️ Phase 1: Hardware Assembly & Power (Ongoing)
- [x] Connect HLK-LD2420 mmWave sensor to ESP32-S3.
- [ ] **Power Bridge:** Configure XL6019 to output 12V for the TPA3110 Amp.
- [ ] **Switched Rail:** Wire IRLZ44N MOSFET to gate the 12V Boost/Amp (controlled by ESP32).
- [x] **Battery Config:** 3.7V Li-Ion cells in parallel for high capacity and CN3791 compatibility.

## 📡 Phase 2: Raspberry Pi 4 "Hub" Setup (High Priority)
- [ ] **Access Point Mode:** Configure `hostapd` and `dnsmasq` to create `Scarrow_Hub_XXX` network.
- [ ] **Static IP:** Set Pi to `192.168.4.1` to act as the Gateway/MQTT Broker address.
- [ ] **MQTT Broker:** Install and secure `Mosquitto` on the Pi.

## 💻 Phase 3: The `pi-go-service` (Local Bridge)
- [ ] **MQTT Subscriber:** Listen to `nodes/+/telemetry` and `nodes/+/events`.
- [ ] **Local Storage:** Implement SQLite logic to save logs while offline in the field.
- [ ] **Sync Engine:** Detect internet (Ethernet/Secondary Wi-Fi) and push buffered data to the main Go API.
- [ ] **Control Logic:** Allow the Pi to send "Manual Trigger" commands back to nodes via MQTT.

## ⚡ Phase 4: ESP32-S3 Firmware (`esp32-node`)
- [x] **Sensor Reading:** Poll HLK-LD2420 mmWave sensor.
- [ ] **Local Networking:** Logic to connect to the `Scarrow_Hub` Wi-Fi.
- [ ] **MQTT Reporting:** Publish distance and detection events to `192.168.4.1`.
- [ ] **Power Management:** Deep Sleep mode; wake up on LD2420 "OUT" pin trigger.
- [ ] **Audio Deterrence:** Tone generation and MOSFET power gating logic.

## 📱 Phase 5: Mobile App & Provisioning
- [ ] UI to connect to the Pi's Hotspot to view local logs.
- [ ] Node registration/pairing workflow.

---
*Progress Tracking:*
- [x] Initial hardware review & feasibility check.
- [x] LD2420 mmWave Sensor integration complete.
- [x] **Architecture Pivot:** Confirmed Parallel Battery + Pi-as-AP for field use.
- [ ] (Next step: Phase 3 - Initialize `pi-go-service` & Phase 2 MQTT Setup)
