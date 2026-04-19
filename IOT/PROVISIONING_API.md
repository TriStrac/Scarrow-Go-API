# Scarrow IoT: Provisioning API Specification

This document outlines the requirements for the backend API (`https://scarrow-api.striel.xyz`) to support the Phase 2 Provisioning workflow.

## 1. Hub Registration
The mobile app calls this when it detects a `Scarrow_Hub_Setup` device.

**Endpoint:** `POST /api/hubs/register`
**Role:** Generates a new unique Hub ID and records it in the central database.

**Successful Response (201 Created):**
```json
{
  "hub_id": "HUB-XXXX-YYYY",
  "secret": "S1gned_Hub_Secret_Key",
  "status": "active"
}
```

**BLE JSON Payload for Hub:**
```json
{
  "wifi_ssid": "X",
  "wifi_password": "Y",
  "hub_id": "HUB-XXXX-YYYY",
  "secret": "S1gned_Hub_Secret_Key"
}
```

---

## 2. Node Registration
The mobile app calls this when it detects a `Scarrow_Node_Setup` device.

**Endpoint:** `POST /api/nodes/register`
**Role:** Generates a new Node ID and links it to a specific Hub.

**Successful Response (201 Created):**
```json
{
  "node_id": "NODE-ZZZZ-WWWW",
  "node_secret": "S1gned_Node_Secret_Key",
  "hub_filter": "HUB-XXXX-YYYY",
  "status": "active"
}
```

**BLE JSON Payload for Node:**
```json
{
  "node_id": "NODE-ZZZZ-WWWW",
  "node_secret": "S1gned_Node_Secret_Key",
  "hub_filter": "HUB-XXXX-YYYY"
}
```

---

## 3. Implementation Checklist for Mobile App
To ensure successful delivery of IDs to the hardware, the Mobile App must implement the following BLE logic:

| Step | Action | Detail |
| :--- | :--- | :--- |
| **1** | **Scan** | Filter for `Scarrow_Hub_Setup` or `Scarrow_Node_Setup`. |
| **2** | **Connect** | Establish GATT connection. |
| **3** | **MTU** | **MANDATORY:** Request `MTU = 512` immediately after connection. |
| **4** | **API Call** | Call the relevant endpoint above to get the IDs. |
| **5** | **BLE Write** | Write the JSON payload to the Target UUIDs (see below). |

### BLE Target UUIDs (Hardcoded)

**Hub Setup (Raspberry Pi):**
*   **Service:** `d2711001-7101-4471-a710-11710b710c71`
*   **Characteristic:** `d2711002-7101-4471-a710-11710b710c71`

**Node Setup (ESP32):**
*   **Service:** `4fafc201-1fb5-459e-8fcc-c5c9c331914b`
*   **Characteristic:** `beb5483e-36e1-4688-b7f5-ea07361b26a8`
