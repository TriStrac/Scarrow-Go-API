# Scarrow-Go-API - Frontend Integration Guide 🚀

**Base URL:** `http://localhost:8080` (Local) / `https://api.scarrow.com` (Production)
**Version:** 1.10.1

## 🔑 Global Configuration

All protected endpoints (`🔒`) require the following header:
```http
Content-Type: application/json
Authorization: Bearer <your_jwt_token_here>
```

---

## 👥 Users Module (`/api/users`)

### 1. Register a New User (❌ Public)
`POST /api/users/`

*Note: The `number` field must be unique across the system. A phone number can only be registered to one account.*

**Request Payload:**
```json
{
  "username": "johndoe123",
  "password": "securepassword1",
  "number": "09123456789"
}
```

**Success Response (201 Created):**
```json
{
  "message": "Registration initiated. Please verify with the OTP sent to your number.",
  "identifier": "johndoe123"
}
```

### 2. Verify Registration (❌ Public)
`POST /api/users/verify-registration`

**Request Payload:**
```json
{
  "identifier": "johndoe123",
  "code": "123456"
}
```

**Success Response (200 OK):**
```json
{
  "message": "User verified successfully.",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid-1234",
    "username": "johndoe123",
    "phone_number": "09123456789",
    "subscription_status": "PREMIUM"
  }
}
```

### 3. Login - Step 1 (❌ Public)
`POST /api/users/login`

**Request Payload:**
```json
{
  "username": "johndoe123",
  "password": "securepassword1"
}
```

**Success Response (200 OK):**
```json
{
  "message": "OTP sent for login verification",
  "identifier": "johndoe123"
}
```

### 4. Resend OTP (❌ Public)
`POST /api/users/resend-otp`

**Request Payload:**
```json
{
  "identifier": "johndoe123",
  "purpose": "REGISTRATION"
}
```
*(Valid `purpose` values: "REGISTRATION", "LOGIN", "FORGOT_PASSWORD")*

**Success Response (200 OK):**
```json
{
  "message": "OTP resent successfully"
}
```

### 5. Login - Step 2 (❌ Public)
`POST /api/users/verify-login`

**Request Payload:**
```json
{
  "identifier": "johndoe123",
  "code": "654321"
}
```

**Success Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 6. Forgot Password (❌ Public)
`POST /api/users/forgot-password`

**Request Payload:**
```json
{ "username": "johndoe123" }
```

### 7. Reset Password (❌ Public)
`POST /api/users/reset-password`

**Request Payload:**
```json
{
  "username": "johndoe123",
  "otp": "112233",
  "new_password": "newsecurepassword123"
}
```

### 8. Get My Session Profile (🔒 Protected)
`GET /api/users/me`

**Returns lightweight session and identity state to drive UI routing.**

**Success Response (200 OK):**
```json
{
  "user_id": "uuid-123",
  "username": "johndoe123",
  "role": "HEAD",
  "group_id": "group-uuid",
  "group_name": "My Farm Co",
  "subscription_status": "PREMIUM",
  "profile_complete": true
}
```

### 9. Get Full Profile (🔒 Protected)
`GET /api/users/:userId`

**Returns the comprehensive user state for caching.**

### 10. Save Push Token (🔒 Protected)
`POST /api/users/me/push-tokens`

**Request Payload:**
```json
{
  "token": "fcm_token_string",
  "platform": "android"
}
```

### 11. Remove Push Token (🔒 Protected)
`DELETE /api/users/me/push-tokens/:tokenId`

### 12. Hard Delete User (Development) (🔒 Protected)
`DELETE /api/users/:userId/hard`

**Note: For development purposes. Permanently deletes the user account and perfectly wipes all associated data (messages, logs, tokens, profiles, etc.) from the database.*

### 13. Update User Address (🔒 Protected)
`PATCH /api/users/:userId/address`

**Description:** Updates the user's address. Address is auto-created (empty) on registration. Only the user themselves can update.

**Request Payload:**
```json
{
  "street_name": "123 Main St",
  "baranggay": "San Jose",
  "town": "Makati",
  "province": "Metro Manila",
  "zip_code": "1205"
}
```

**Success Response (200 OK):**
```json
{
  "message": "Address updated successfully"
}
```

---

## 🏢 Groups / Companies Module (`/api/groups`)

### 1. Create a Group (🔒 Protected)
`POST /api/groups/`

**Request Payload:**
```json
{ "name": "My Farm Co" }
```
*Note: The user who creates the group is automatically assigned as the `HEAD` of the newly created group.*

### 2. Get Group Details (🔒 Protected)
`GET /api/groups/:groupId`

**Success Response (200 OK):**
```json
{
  "id": "group-uuid",
  "name": "My Farm Co",
  "owner_id": "owner-uuid",
  "role": "HEAD",
  "member_count": 5,
  "settings": {}
}
```

### 3. Get Group Members (🔒 Protected)
`GET /api/groups/:groupId/members`

**Success Response (200 OK):**
```json
[
  {
    "user_id": "uuid-1",
    "display_name": "John Doe",
    "role": "HEAD"
  },
  {
    "user_id": "uuid-2",
    "display_name": "jane_smith",
    "role": "MEMBER"
  }
]
```

### 4. Get Member Devices (🔒 Protected, Head only)
`GET /api/groups/:groupId/members/:userId/devices`

**Description:** Head farmer views a group member's devices. Only the group head can call this.

**Success Response (200 OK):**
```json
{
  "devices": [
    { "id": "...", "name": "North Field Hub", "type": "CENTRAL", "status": "ONLINE", "created_at": "..." }
  ]
}
```

### 5. Get Member Activity Logs (🔒 Protected, Head only)
`GET /api/groups/:groupId/members/:userId/activity-logs?limit=50&offset=0`

**Description:** Head farmer views a group member's activity logs. Only the group head can call this. Query params `limit` (default 50) and `offset` (default 0) are optional.

**Success Response (200 OK):**
```json
{
  "logs": [
    { "id": 1, "action": "Logged in", "module": "AUTH", "created_at": "2026-05-07T10:00:00Z" }
  ],
  "limit": 50,
  "offset": 0
}
```

### 6. Create Invitation Code (🔒 Protected)
`POST /api/groups/:groupId/invite`

**Success Response (201 Created):**
```json
{
  "invitation": {
    "code": "FARM1234",
    "group_id": "group-uuid",
    "expires_at": "2026-04-11T..."
  }
}
```

### 7. Join Group via Code (🔒 Protected)
`POST /api/groups/join`

**Request Payload:**
```json
{ "code": "FARM1234" }
```

### 8. Remove Member (Admin only) (🔒 Protected)
`DELETE /api/groups/member`

**Request Payload:**
```json
{ "group_id": "...", "user_id": "..." }
```

### 9. Leave Group (Members only) (🔒 Protected)
`POST /api/groups/leave`

**Request Payload:**
```json
{ "group_id": "..." }
```

### 10. Disband Group (Owner only) (🔒 Protected)
`DELETE /api/groups/:groupId`

**Description:** Disbands the entire group. This resets all former members to `SOLO` status, unpairs any devices owned by the group, deletes pending invitations, and fires a "Group Disbanded" notification to all previous members.

**Success Response (200 OK):**
```json
{
  "message": "Group disbanded successfully"
}
```

---

## Devices Module (`/api/device`)

### Phase 2: Hardware Provisioning (BLE "Zero-Router" Setup)

### 1. Register a Hub (Raspberry Pi) (🔒 Protected)
`POST /api/hubs/register`

**Role:** Used by the Mobile App during BLE provisioning to register a new central hub. Generates a unique Hub ID and a Secret.

**Request Payload:**
```json
{
  "name": "North Field Hub",
  "location_lat": 14.5995,
  "location_lng": 120.9842
}
```

**Success Response (201 Created):**
```json
{
  "hub_id": "HUB-XXXX-YYYY",
  "secret": "RandomString32",
  "status": "active"
}
```

### 2. Register a Node (ESP32) (🔒 Protected)
`POST /api/nodes/register`

**Role:** Used by the Mobile App during BLE provisioning to register a new field node. Generates a unique Node ID and a Node Secret for Hub communication.

**Request Payload:**
```json
{
  "hub_id": "HUB-XXXX-YYYY",
  "node_type": "deterrence_v1",
  "label": "Corn Corner A"
}
```

**Success Response (201 Created):**
```json
{
  "node_id": "NODE-ZZZZ-WWWW",
  "node_secret": "RandomString32",
  "hub_filter": "HUB-XXXX-YYYY",
  "status": "active"
}
```

### Legacy Device Endpoints

### 3. Create Device (Legacy) (🔒 Protected)
`POST /api/device/`

**Request Payload:**
```json
{
  "name": "Node 01",
  "owner_type": "USER",
  "device_type": "NODE",
  "parent_id": "central-hub-uuid" 
}
```
*(Use `device_type`: 'CENTRAL' or 'NODE').*

### 4. Get My Devices (🔒 Protected)
`GET /api/device/my`

### 5. Get Device Logs (History) (🔒 Protected)
`GET /api/device/:deviceId/logs?limit=50&offset=0`

**Returns paginated telemetry history.**

---

### IoT Device Log Ingestion (Hub-to-API)

For Raspberry Pi hubs sending logs from ESP32 nodes. Uses device authentication instead of user JWT.

### 6. Create Device Log (IoT Device Auth 🔐)
`POST /api/device/:deviceId/logs`

**Authentication:** Uses `X-Device-ID` and `X-Device-Secret` headers (NOT user JWT).

**Request Headers:**
```http
X-Device-ID: HUB-XXXX-YYYY
X-Device-Secret: <hub_secret_from_registration>
Content-Type: application/json
```

**Request Payload:**
```json
{
  "node_id": "NODE-ZZZZ-WWWW",
  "log_type": "DETECTED",
  "pest_type": "BIRD",
  "duration_seconds": 30,
  "frequency_hz": 0.0,
  "payload": "{}"
}
```

**Success Response (201 Created):**
```json
{
  "message": "Log created successfully"
}
```

**Error Responses:**
- `401`: Authentication failed (invalid device ID or secret)
- `403`: Node does not belong to this hub
- `404`: Device not found

---

### 7. Send Command to Hub (🔒 Protected)
`POST /api/hubs/:hubId/commands`

**Description:** Sends a command to a Raspberry Pi hub via WebSocket. Commands are executed asynchronously on the Pi.

**Authentication:** User must be owner of the hub.

**Request Payload:**
```json
{
  "cmd": "reboot"
}
```

**Available Commands:**

| Command | Args | Description |
|---------|------|-------------|
| `reboot` | none | Reboots the Raspberry Pi |
| `wifi` | `ssid`, `password` | Configures Wi-Fi connection via nmcli |
| `reset` | none | Resets hub to setup mode (clears config, restarts service) |
| `node_reset` | `node_id` | Sends reset command to an ESP32 node via BLE |

**Command Examples:**

Reboot:
```json
{
  "cmd": "reboot"
}
```

Change Wi-Fi:
```json
{
  "cmd": "wifi",
  "ssid": "MyNetwork",
  "password": "secretpass123"
}
```

Reset to Setup Mode:
```json
{
  "cmd": "reset"
}
```

Reset Node:
```json
{
  "cmd": "node_reset",
  "node_id": "NODE-A1B2-C3D4"
}
```

**Success Response (200 OK):**
```json
{
  "message": "Command sent",
  "cmd": "reboot",
  "hub_id": "HUB-XXXX-YYYY",
  "async": true
}
```

**Error Responses:**
- `400`: Invalid command or missing required args (e.g., wifi without ssid/password)
- `403`: User is not the owner of this hub
- `503`: WebSocket not connected (Pi is offline)

---

## ✉️ Messages Module (`/api/messages`)

### 1. List My Threads (🔒 Protected)
`GET /api/messages/?limit=50&offset=0`

### 2. Get Unread Summary (🔒 Protected)
`GET /api/messages/unread-summary`

**Returns the total number of unread messages across all threads for badges.**

**Success Response (200 OK):**
```json
{
  "unread_count": 3
}
```

### 3. Get Thread History (🔒 Protected)
`GET /api/messages/:threadId?limit=50&offset=0`
*(Returns thread metadata and messages. Marks retrieved messages as read).*

### 4. Send Message (🔒 Protected)
`POST /api/messages/`

**Request Payload:**
```json
{
  "receiver_id": "other-user-uuid",
  "content": "Alert: Unusual activity on Node 2."
}
```

---

## 📊 Reports Module (`/api/reports`)

### 1. Get Analytics Summary (🔒 Protected)
`GET /api/reports/summary?timeframe=last_7_days`

**Returns aggregated device data and alerts for charting.**

**Success Response (200 OK):**
```json
{
  "overview": {
    "total_alerts": 41,
    "total_devices": 2
  },
  "pest_distribution": {
    "BIRDS": 22,
    "LOCUST": 14,
    "RATS": 5
  },
  "daily_trends": [
    { "count": 2, "date": "2026-04-06" },
    { "count": 5, "date": "2026-04-07" }
  ],
  "timeframe": "last_7_days"
}
```

---

### 2. Get Hub Report (🔒 Protected)
`GET /api/reports/hub/:hubId?start_date=2026-01-01&end_date=2026-12-31`

**Description:** Returns detection report for a specific hub and its nodes. Query params `start_date` and `end_date` are optional (defaults to all time).

**Response (200 OK):**
```json
{
  "total": 22,
  "detections": [
    {
      "pest_type": "BIRD",
      "count": 15,
      "logs": [
        { "created_at": "2026-04-06T10:00:00Z", "duration_seconds": 30 },
        { "created_at": "2026-04-07T14:00:00Z", "duration_seconds": 45 }
      ]
    },
    {
      "pest_type": "LOCUST",
      "count": 7,
      "logs": []
    }
  ]
}
```

---

## 💳 Subscriptions Module (`/api/subscriptions`)

### 1. Get Available Plans (❌ Public)
`GET /api/subscriptions/plans`

**Success Response (200 OK):**
```json
{
  "plans": [
    {
      "id": "plan_monthly",
      "name": "Premium Farmer (Monthly)",
      "description": "Full access to analytics and unlimited devices for 30 days.",
      "price": 499.00,
      "duration_days": 30
    }
  ]
}
```

### 2. Get My Subscription (🔒 Protected)
`GET /api/subscriptions/my`

**Success Response (200 OK):**
```json
{
  "subscription": {
    "id": "sub_123",
    "plan_id": "plan_monthly",
    "status": "ACTIVE",
    "start_date": "2026-04-10T10:00:00Z",
    "end_date": "2026-05-10T10:00:00Z"
  }
}
```

### 3. Create Checkout Session (🔒 Protected)
`POST /api/subscriptions/checkout`

**Request Payload:**
```json
{ "plan_id": "plan_monthly" }
```

**Success Response (200 OK):**
```json
{
  "checkout_url": "https://checkout.paymongo.com/...",
  "reference_id": "pi_12345"
}
```

### 4. Verify/Restore Payment (🔒 Protected)
`POST /api/subscriptions/verify`

**Request Payload:**
```json
{ "reference_id": "pi_12345" }
```

---

## 🔔 Notifications Module (`/api/notifications`)

### 1. Get My Notifications (🔒 Protected)
`GET /api/notifications/my`

**Returns paginated notification history.**

### 2. Long Poll for New Notifications (🔒 Protected)
`GET /api/notifications/poll?since=2026-05-04T00:00:00Z`

**Description:** Long-polling endpoint that holds the connection until a new notification arrives or 80 seconds elapse. Mobile app uses this for real-time notification updates.

**Query Parameters:**
- `since` (optional): ISO8601 timestamp. Returns notifications newer than this time. If omitted, returns all unread.

**Response (new notifications found):**
```json
{
  "notifications": [
    {
      "id": "notification-uuid",
      "title": "New message",
      "message": "You have a new message from John",
      "is_read": false,
      "created_at": "2026-05-04T10:30:00Z"
    }
  ],
  "timeout": false
}
```

**Response (timeout — no new notifications):**
```json
{
  "notifications": [],
  "timeout": true
}
```

**Mobile Polling Flow:**
```
1. Mobile calls GET /notifications/poll?since=<last_received_notification_timestamp>
2. Server holds connection (up to 80s)
3. On new notification: server responds immediately with notification data
4. On timeout: server responds with empty notifications + timeout:true
5. Mobile immediately reconnects and polls again
```

### 3. Mark Single Notification as Read (🔒 Protected)
`PATCH /api/notifications/:notificationId/read`

### 4. Mark All as Read (🔒 Protected)
`PATCH /api/notifications/read-all`

---

## 📜 Activity Logs (`/api/activityLogs`)
`GET /api/activityLogs/my`
*(Returns human-readable logs like "Logged in", "Joined a group", etc).*
