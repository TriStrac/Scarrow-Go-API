# Scarrow-Go-API - Frontend Integration Guide 🚀

This guide provides everything a frontend developer needs to integrate with the Scarrow Backend.

**Base URL:** `http://localhost:8080` (Local) / `https://api.scarrow.com` (Production)

---

## 🔑 Global Configuration

### Headers
All protected endpoints (`🔒`) require the following header:
| Header | Value | Description |
| :--- | :--- | :--- |
| `Content-Type` | `application/json` | Required for all POST/PATCH requests |
| `Authorization` | `Bearer <token>` | Required for all protected routes |

---

## 👥 Users Module (`/api/users`)

| Method | Endpoint | Auth | Description | Request Payload | Success Response (200/201) | Common Errors |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `POST` | `/` | ❌ | **Register** | `{ "username": "...", "password": "...", "profile": {...}, "address": {...} }` | `{ "message": "...", "user": {...} }` | `400`: Validation<br>`409`: Username taken |
| `POST` | `/login` | ❌ | **Login** | `{ "username": "...", "password": "..." }` | `{ "token": "eyJhbG..." }` | `401`: Invalid credentials |
| `GET` | `/usernameExists` | ❌ | **Check User** | `?username=johndoe` (Query) | `{ "exists": true }` | `400`: Missing param |
| `GET` | `/` | 🔒 | **List All** | `None` | `{ "users": [...] }` | `401`: Unauthorized |
| `GET` | `/:userId` | 🔒 | **Get Info** | `None` | `{ "user": {...} }` | `404`: Not found |
| `PATCH`| `/:userId` | 🔒 | **Update** | `{ "username": "...", "profile": {...} }` | `{ "message": "Updated" }` | `400`: Bad Payload |
| `POST` | `/changePassword`| 🔒 | **Password** | `{ "new_password": "..." }` | `{ "message": "Changed" }` | `400`: Min length 6 |
| `PATCH`| `/:userId/softDelete`| 🔒 | **Delete** | `None` | `{ "message": "Deleted" }` | `404`: Not found |

---

## 🏢 Groups / Companies Module (`/api/groups`)
*Note: A user can belong to **strictly one or zero** Groups. Addition will fail if the user is already employed.*

| Method | Endpoint | Auth | Description | Request Payload | Success Response (200/201) | Common Errors |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `POST` | `/` | 🔒 | **Create** | `{ "name": "Company Name" }` | `{ "message": "Created", "group": {...} }` | `409`: Name taken |
| `GET` | `/` | 🔒 | **List All** | `None` | `{ "groups": [...] }` | `401`: Unauthorized |
| `GET` | `/owner` | 🔒 | **My Groups** | `None` | `{ "groups": [...] }` | `401`: Unauthorized |
| `GET` | `/:groupId` | 🔒 | **Get Group** | `None` | `{ "group": {...} }` | `404`: Not found |
| `PATCH`| `/:groupId` | 🔒 | **Rename** | `{ "name": "New Name" }` | `{ "message": "Updated" }` | `409`: Name taken |
| `POST` | `/member` | 🔒 | **Add Member**| `{ "group_id": "...", "username": "..." }` | `{ "message": "Added" }` | `400`: User in another group |
| `DELETE`| `/member` | 🔒 | **Remove** | `{ "group_id": "...", "user_id": "..." }` | `{ "message": "Removed" }` | `404`: Not found |
| `GET` | `/:groupId/members`| 🔒 | **Members** | `None` | `{ "members": [...] }` | `404`: Not found |

---

## 📦 Implementation Reference (JSON Samples)

### 1. Register User (Full Body)
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

### 2. User Object Structure (Response)
```json
{
  "id": "uuid-string",
  "username": "johndoe123",
  "group_id": null,
  "is_user_in_group": false,
  "is_user_head": false,
  "profile": {
    "first_name": "John",
    "last_name": "Doe",
    "birth_date": "1990-01-01T00:00:00Z",
    "phone_number": "09123456789"
  },
  "address": {
    "street_name": "123 Main St",
    "baranggay": "Brgy. San Jose",
    "town": "Pasig City"
  }
}
```

### 3. Group Object Structure (Response)
```json
{
  "id": "uuid-string",
  "name": "Scarrow Tech Innovations",
  "owner_id": "owner-uuid",
  "members": []
}
```

### 4. Error Response Format (Standard)
```json
{
  "error": "Error message description here"
}
```
