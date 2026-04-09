# Scarrow-Go-API - Frontend Integration Guide 🚀

**Base URL:** `http://localhost:8080` (Local) / `https://api.scarrow.com` (Production)
**Version:** 1.2

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

**Request Payload:**
```json
{
  "first_name": "John",
  "last_name": "Doe",
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

### 4. Login - Step 2 (❌ Public)
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

### 5. Forgot Password (❌ Public)
`POST /api/users/forgot-password`

**Request Payload:**
```json
{ "username": "johndoe123" }
```

### 6. Reset Password (❌ Public)
`POST /api/users/reset-password`

**Request Payload:**
```json
{
  "username": "johndoe123",
  "otp": "112233",
  "new_password": "newsecurepassword123"
}
```

### 7. Get Full Profile (🔒 Protected)
`GET /api/users/:userId`

**Returns the comprehensive user state for caching.**

**Success Response (200 OK):**
```json
{
  "user": {
    "id": "uuid-123",
    "username": "johndoe123",
    "is_verified": true,
    "subscription_status": "FREE",
    "profile": { "first_name": "John", "last_name": "Doe", "phone_number": "0912..." },
    "address": { "street_name": "...", "town": "..." }
  },
  "devices": [
    { "id": "dev-1", "name": "Central Hub", "device_type": "CENTRAL", "status": "ONLINE" }
  ],
  "recent_messages": [
    { "id": "msg-1", "content": "Hello!", "sender_id": "other-uuid", "created_at": "..." }
  ],
  "unread_messages_count": 2
}
```

---

## 🏢 Groups / Companies Module (`/api/groups`)

### 1. Create a Group (🔒 Protected)
`POST /api/groups/`
`{ "name": "My Farm Co" }`

### 2. Create Invitation Code (🔒 Protected)
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

### 3. Join Group via Code (🔒 Protected)
`POST /api/groups/join`

**Request Payload:**
```json
{ "code": "FARM1234" }
```

### 4. Remove Member (🔒 Protected)
`DELETE /api/groups/member`
`{ "group_id": "...", "user_id": "..." }`

---

## 📱 Devices Module (`/api/device`)

### 1. Create Device (🔒 Protected)
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

### 2. Get My Devices (🔒 Protected)
`GET /api/device/my`

### 3. Create Telemetry Log (🔒 Protected)
`POST /api/device/:deviceId/logs`

**Request Payload:**
```json
{
  "log_type": "TELEMETRY",
  "pest_type": "LOCUST",
  "frequency_hz": 14500.5,
  "duration_seconds": 30,
  "payload": "{\"raw_sensor_data\": \"...\"}"
}
```

---

## ✉️ Messages Module (`/api/messages`)

### 1. List My Threads (🔒 Protected)
`GET /api/messages/`

### 2. Get Thread History (🔒 Protected)
`GET /api/messages/:threadId`
*(Returns last 50 messages, marks them as read).*

### 3. Send Message (🔒 Protected)
`POST /api/messages/`

**Request Payload:**
```json
{
  "receiver_id": "other-user-uuid",
  "content": "Alert: Unusual activity on Node 2."
}
```

---

## 🔔 Notifications Module (`/api/notifications`)

### 1. Get My Notifications (🔒 Protected)
`GET /api/notifications/my`

### 2. Mark as Read (🔒 Protected)
`PATCH /api/notifications/:notificationId/read`

### 3. Mark All as Read (🔒 Protected)
`PATCH /api/notifications/read-all`

---

## 📜 Activity Logs (`/api/activityLogs`)
`GET /api/activityLogs/my`
*(Returns human-readable logs like "Logged in", "Joined a group", etc).*
