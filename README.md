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
  "username": "johndoe123",
  "password": "securepassword1",
  "profile": {
    "first_name": "John",
    "middle_name": "D",
    "last_name": "Doe",
    "birth_date": "1990-01-01T00:00:00Z",
    "phone_number": "09123456789"
  },
  "address": {
    "street_name": "123 Main St",
    "baranggay": "Brgy. San Jose",
    "town": "Pasig City",
    "province": "Metro Manila",
    "zip_code": "1600"
  }
}
```

**Success Response (201 Created):**
```json
{
  "message": "User created successfully",
  "user": {
    "id": "e2b2961e-1234-4b56-8a90-123456789abc",
    "username": "johndoe123",
    "group_id": null,
    "is_user_in_group": false,
    "is_user_head": false,
    "is_deleted": false,
    "created_at": "2026-04-05T10:00:00Z",
    "updated_at": "2026-04-05T10:00:00Z",
    "profile": {
      "id": "f5c3a12b-...",
      "user_id": "e2b2961e-1234-4b56-8a90-123456789abc",
      "first_name": "John",
      "middle_name": "D",
      "last_name": "Doe",
      "birth_date": "1990-01-01T00:00:00Z",
      "phone_number": "09123456789"
    },
    "address": {
      "id": "a1b2c3d4-...",
      "user_id": "e2b2961e-1234-4b56-8a90-123456789abc",
      "street_name": "123 Main St",
      "baranggay": "Brgy. San Jose",
      "town": "Pasig City",
      "province": "Metro Manila",
      "zip_code": "1600"
    }
  }
}
```

### 2. Login (❌ Public)
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
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3..."
}
```

### 3. Check Username Exists (❌ Public)
`GET /api/users/usernameExists?username=johndoe123`

*No request body required.*

**Success Response (200 OK):**
```json
{
  "exists": true
}
```

### 4. Get All Users (🔒 Protected)
`GET /api/users/`

*No request body required.*

**Success Response (200 OK):**
```json
{
  "users": [
    {
      "id": "e2b2961e-...",
      "username": "johndoe123",
      "group_id": null,
      "is_user_in_group": false,
      "is_user_head": false,
      "is_deleted": false,
      "created_at": "2026-04-05T10:00:00Z",
      "updated_at": "2026-04-05T10:00:00Z",
      "profile": null,
      "address": null
    }
  ]
}
```

### 5. Get User by ID (🔒 Protected)
`GET /api/users/:userId`

*No request body required.*

**Success Response (200 OK):**
*(Returns the full user object including nested profile and address, exactly like the Register response).*

### 6. Update User / Partial Update (🔒 Protected)
`PATCH /api/users/:userId`

*Send ONLY the fields you want to update. Omitted fields remain untouched.*

**Request Payload (Example):**
```json
{
  "username": "newjohndoe",
  "profile": {
    "phone_number": "09999999999"
  }
}
```

**Success Response (200 OK):**
```json
{
  "message": "User updated successfully"
}
```

### 7. Change Password (🔒 Protected)
`POST /api/users/changePassword`

**Request Payload:**
```json
{
  "new_password": "mynewpassword123"
}
```

**Success Response (200 OK):**
```json
{
  "message": "Password changed successfully"
}
```

### 8. Soft Delete User (🔒 Protected)
`PATCH /api/users/:userId/softDelete`

*No request body required.*

**Success Response (200 OK):**
```json
{
  "message": "User soft deleted successfully"
}
```

---

## 🏢 Groups / Companies Module (`/api/groups`)

*⚠️ Strict Rule: A user can belong to **strictly one or zero** Groups (Companies). Trying to add a user to a group when they are already in one will fail.*

### 1. Create a Group (🔒 Protected)
`POST /api/groups/`

**Request Payload:**
```json
{
  "name": "Scarrow Tech Innovations"
}
```

**Success Response (201 Created):**
```json
{
  "message": "Group created successfully",
  "group": {
    "id": "b7d8e9f0-...",
    "name": "Scarrow Tech Innovations",
    "owner_id": "e2b2961e-1234...",
    "is_deleted": false,
    "created_at": "2026-04-05T10:00:00Z",
    "updated_at": "2026-04-05T10:00:00Z",
    "owner": null,
    "members": null
  }
}
```

### 2. Get All Groups (🔒 Protected)
`GET /api/groups/`

*No request body required.*

**Success Response (200 OK):**
```json
{
  "groups": [
    {
      "id": "b7d8e9f0-...",
      "name": "Scarrow Tech Innovations",
      "owner_id": "e2b2961e-...",
      "is_deleted": false,
      "created_at": "...",
      "updated_at": "...",
      "owner": {
        "id": "e2b2961e-...",
        "username": "owneruser",
        "...": "..."
      },
      "members": null
    }
  ]
}
```

### 3. Get My Groups (Owned by Me) (🔒 Protected)
`GET /api/groups/owner`

*No request body required.*

**Success Response (200 OK):** *(Returns array of groups exactly like Get All Groups).*

### 4. Get Group by ID (🔒 Protected)
`GET /api/groups/:groupId`

*No request body required.*

**Success Response (200 OK):** *(Returns single group object).*

### 5. Rename Group (🔒 Protected)
`PATCH /api/groups/:groupId`

**Request Payload:**
```json
{
  "name": "Scarrow Tech Innovations LLC"
}
```

**Success Response (200 OK):**
```json
{
  "message": "Group updated successfully"
}
```

### 6. Soft Delete Group (🔒 Protected)
`PATCH /api/groups/:groupId/softDelete`

*No request body required.*

**Success Response (200 OK):**
```json
{
  "message": "Group deleted successfully"
}
```

### 7. Add Member to Group (🔒 Protected)
`POST /api/groups/member`

**Request Payload:**
```json
{
  "group_id": "b7d8e9f0-...",
  "username": "janedoe99"
}
```

**Success Response (200 OK):**
```json
{
  "message": "Member added successfully"
}
```
**Common Errors:**
* `400 Bad Request`: `{"error": "user already belongs to a group/company"}`

### 8. Remove Member from Group (🔒 Protected)
`DELETE /api/groups/member`

**Request Payload:**
```json
{
  "group_id": "b7d8e9f0-...",
  "user_id": "user-uuid-here"
}
```

**Success Response (200 OK):**
```json
{
  "message": "Member removed successfully"
}
```

### 9. Get Group Members (🔒 Protected)
`GET /api/groups/:groupId/members`

*No request body required.*

**Success Response (200 OK):**
```json
{
  "members": [
    {
      "id": "user-uuid-here",
      "username": "janedoe99",
      "group_id": "b7d8e9f0-...",
      "is_user_in_group": true,
      "...": "..."
    }
  ]
}
```

---

## 📱 Devices Module (`/api/device`)

### 1. Create a Device (🔒 Protected)
`POST /api/device/`

**Request Payload:**
```json
{
  "name": "Smart Controller V1",
  "owner_type": "USER"
}
```

**Success Response (201 Created):**
```json
{
  "message": "Device created successfully",
  "device": {
    "id": "uuid-here",
    "name": "Smart Controller V1",
    "status": "OFFLINE",
    "created_at": "...",
    "updated_at": "..."
  }
}
```

### 2. Get All Devices (🔒 Protected)
`GET /api/device/`

### 3. Get My Devices (🔒 Protected)
`GET /api/device/my`

### 4. Get Device by ID (🔒 Protected)
`GET /api/device/:deviceId`

### 5. Update Device (🔒 Protected)
`PATCH /api/device/:deviceId`

**Request Payload:**
```json
{
  "name": "Updated Name",
  "status": "ONLINE"
}
```

### 6. Soft Delete Device (🔒 Protected)
`DELETE /api/device/:deviceId`

### 7. Add Owner to Device (🔒 Protected)
`POST /api/device/:deviceId/owner`

**Request Payload:**
```json
{
  "owner_id": "owner-uuid-here",
  "owner_type": "GROUP"
}
```

### 8. Remove Owner from Device (🔒 Protected)
`DELETE /api/device/:deviceId/owner`

**Request Payload:**
```json
{
  "owner_id": "owner-uuid-here",
  "owner_type": "GROUP"
}
```

### 9. Get Device Owners (🔒 Protected)
`GET /api/device/:deviceId/owners`

### 10. Create Device Log (🔒 Protected)
`POST /api/device/:deviceId/logs`

**Request Payload:**
```json
{
  "log_type": "TELEMETRY",
  "payload": "{\"temp\": 25.5, \"humidity\": 60}"
}
```

### 11. Get Device Logs (🔒 Protected)
`GET /api/device/:deviceId/logs`

---

## 📜 Activity Logs Module (`/api/activityLogs`)

### 1. Get All Activity Logs (🔒 Protected)
`GET /api/activityLogs/`

### 2. Get My Activity Logs (🔒 Protected)
`GET /api/activityLogs/my`

---

## 🚫 Standard Error Response
If any request fails (Validation, Not Found, Conflict, Unauthorized), the API will return an appropriate HTTP Status Code and a JSON body in the following format:

```json
{
  "error": "Detailed error message here"
}
```
