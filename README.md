# Scarrow-Go-API - Frontend Integration Guide 🚀

**Base URL:** `http://localhost:8080` (Local) / `https://api.scarrow.com` (Production)

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
  "code": "123456"
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
  "otp": "123456",
  "new_password": "newsecurepassword"
}
```

### 7. Get Full Profile (🔒 Protected)
`GET /api/users/:userId`

**Success Response (200 OK):**
```json
{
  "user": { "id": "...", "username": "...", "subscription_status": "FREE", "...": "..." },
  "devices": [ { "id": "...", "name": "...", "device_type": "CENTRAL" } ],
  "recent_messages": [ { "content": "Hello", "sender_id": "..." } ],
  "unread_messages_count": 2
}
```

---

## 🏢 Groups Module (`/api/groups`)

### 1. Create Invitation Code (🔒 Protected)
`POST /api/groups/:groupId/invite`

**Success Response (201 Created):**
```json
{
  "invitation": {
    "code": "ABCDEFGH",
    "expires_at": "..."
  }
}
```

### 2. Join Group via Code (🔒 Protected)
`POST /api/groups/join`

**Request Payload:**
```json
{ "code": "ABCDEFGH" }
```

---

## 📱 Devices Module (`/api/device`)

### 1. Create Hierarchical Device (🔒 Protected)
`POST /api/device/`

**Request Payload:**
```json
{
  "name": "Node 1",
  "owner_type": "USER",
  "device_type": "NODE",
  "parent_id": "central-device-uuid"
}
```

### 2. Create Telemetry Log (🔒 Protected)
`POST /api/device/:deviceId/logs`

**Request Payload:**
```json
{
  "log_type": "TELEMETRY",
  "pest_type": "LOCUST",
  "frequency_hz": 12500.5,
  "duration_seconds": 15,
  "payload": "{...}"
}
```

---

## ✉️ Messages Module (`/api/messages`)

### 1. List My Threads (🔒 Protected)
`GET /api/messages/`

### 2. Get Thread History (🔒 Protected)
`GET /api/messages/:threadId`

### 3. Send Message (🔒 Protected)
`POST /api/messages/`

**Request Payload:**
```json
{
  "receiver_id": "other-user-uuid",
  "content": "Hello farmer!"
}
```

---

## 🔔 Notifications Module (`/api/notifications`)

### 1. Get My Notifications (🔒 Protected)
`GET /api/notifications/my`

### 2. Mark as Read (🔒 Protected)
`PATCH /api/notifications/:notificationId/read`
