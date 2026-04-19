# Scarrow-Go-API - Frontend Integration Guide 🚀

**Base URL:** `http://localhost:8080` (Local) / `https://api.scarrow.com` (Production)
**Version:** 1.6.0

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

**Note: For development purposes. Permanently deletes the user account and perfectly wipes all associated data (messages, logs, tokens, profiles, etc.) from the database.**

---

## 🏢 Groups / Companies Module (`/api/groups`)

### 1. Create a Group (🔒 Protected)
`POST /api/groups/`
`{ "name": "My Farm Co" }`

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

### 4. Create Invitation Code (🔒 Protected)
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

### 5. Join Group via Code (🔒 Protected)
`POST /api/groups/join`

**Request Payload:**
```json
{ "code": "FARM1234" }
```

### 6. Remove Member (Admin only) (🔒 Protected)
`DELETE /api/groups/member`
`{ "group_id": "...", "user_id": "..." }`

### 7. Leave Group (Members only) (🔒 Protected)
`POST /api/groups/leave`

**Request Payload:**
```json
{ "group_id": "..." }
```

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

### 4. Get Device Logs (History) (🔒 Protected)
`GET /api/device/:deviceId/logs?limit=50&offset=0`

**Returns paginated telemetry history.**

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

### 2. Mark as Read (🔒 Protected)
`PATCH /api/notifications/:notificationId/read`

### 3. Mark All as Read (🔒 Protected)
`PATCH /api/notifications/read-all`

---

## 📜 Activity Logs (`/api/activityLogs`)
`GET /api/activityLogs/my`
*(Returns human-readable logs like "Logged in", "Joined a group", etc).*
