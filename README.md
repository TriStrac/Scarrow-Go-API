# Scarrow-Go-API 🚀

This is the Go/MySQL refactor of the ScarrowAPI. This document is written specifically for Frontend/Mobile developers to easily integrate with the backend endpoints.

## 🔑 Authentication (JWT)
All protected endpoints require an `Authorization` header containing the JWT token received from the Login endpoint.
```http
Authorization: Bearer <your_jwt_token_here>
```
*Note: If the token is missing, malformed, or expired, the server will return a `401 Unauthorized`.*

---

## 👥 Users Domain (`/api/users`)

| Method | Endpoint | Auth | Purpose |
|--------|----------|------|---------|
| `POST` | `/api/users/` | ❌ | Register a new user |
| `POST` | `/api/users/login` | ❌ | Login and receive JWT |
| `GET` | `/api/users/usernameExists?username={name}` | ❌ | Check if a username is taken |
| `GET` | `/api/users/` | 🔒 | Get all users |
| `GET` | `/api/users/:userId` | 🔒 | Get user by ID (Includes Profile/Address) |
| `PATCH` | `/api/users/:userId` | 🔒 | Update user (Supports Partial Updates) |
| `POST` | `/api/users/changePassword` | 🔒 | Change logged-in user's password |
| `PATCH` | `/api/users/:userId/softDelete`| 🔒 | Soft delete user |

### User Endpoints Detailed Payloads

<details>
<summary><b><code>POST /api/users/</code> - Register User</b></summary>

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
**Expected Responses:**
* `201 Created`: `{ "message": "User created successfully", "user": { ...user_object } }`
* `400 Bad Request`: Validation errors (missing fields)
* `409 Conflict`: `{"error": "username already exists"}`
</details>

<details>
<summary><b><code>POST /api/users/login</code> - Login</b></summary>

**Request Payload:**
```json
{
  "username": "johndoe123",
  "password": "securepassword1"
}
```
**Expected Responses:**
* `200 OK`: `{ "token": "eyJhbGciOiJIUzI1NiIsInR5..." }`
* `401 Unauthorized`: `{"error": "invalid username or password"}`
</details>

<details>
<summary><b><code>GET /api/users/usernameExists?username=...</code> - Check Username</b></summary>

*No body required. Pass username in URL query.*
**Expected Response (200):**
```json
{
  "exists": true
}
```
</details>

<details>
<summary><b><code>PATCH /api/users/:userId</code> - Partial Update</b></summary>

*You only need to send the fields you want to update. Omitted fields are ignored.*
**Request Payload (Example):**
```json
{
  "username": "newjohndoe",
  "profile": {
    "phone_number": "09999999999"
  }
}
```
**Expected Responses:**
* `200 OK`: `{ "message": "User updated successfully" }`
* `409 Conflict`: `{"error": "username already exists"}`
</details>

<details>
<summary><b><code>POST /api/users/changePassword</code> - Change Password</b></summary>

**Request Payload:**
```json
{
  "new_password": "mynewpassword123"
}
```
**Expected Response (200):** `{ "message": "Password changed successfully" }`
</details>

---

## 🏢 Groups / Companies Domain (`/api/groups`)
*⚠️ Strict Rule: A user can belong to **strictly one or zero** Groups (Companies). Trying to add a user to a group when they are already in one will fail.*

| Method | Endpoint | Auth | Purpose |
|--------|----------|------|---------|
| `POST` | `/api/groups/` | 🔒 | Create a new group (company) |
| `GET` | `/api/groups/` | 🔒 | Get all groups |
| `GET` | `/api/groups/owner` | 🔒 | Get all groups owned by logged-in user |
| `GET` | `/api/groups/:groupId` | 🔒 | Get group by ID (Includes members) |
| `PATCH` | `/api/groups/:groupId` | 🔒 | Rename the group |
| `PATCH` | `/api/groups/:groupId/softDelete`| 🔒 | Soft delete the group |
| `POST` | `/api/groups/member` | 🔒 | Add a member to the group (by username) |
| `DELETE` | `/api/groups/member` | 🔒 | Remove a member from the group |
| `GET` | `/api/groups/:groupId/members` | 🔒 | Get an array of all users in the group |

### Group Endpoints Detailed Payloads

<details>
<summary><b><code>POST /api/groups/</code> - Create Group</b></summary>

**Request Payload:**
```json
{
  "name": "Scarrow Tech Innovations"
}
```
**Expected Responses:**
* `201 Created`: `{ "message": "Group created successfully", "group": { ...group_object } }`
* `409 Conflict`: `{"error": "group name already exists"}`
</details>

<details>
<summary><b><code>PATCH /api/groups/:groupId</code> - Rename Group</b></summary>

**Request Payload:**
```json
{
  "name": "Scarrow Tech Innovations LLC"
}
```
**Expected Response (200):** `{ "message": "Group updated successfully" }`
</details>

<details>
<summary><b><code>POST /api/groups/member</code> - Add Member to Group</b></summary>

**Request Payload:**
```json
{
  "group_id": "uuid-of-the-group",
  "username": "janedoe99"
}
```
**Expected Responses:**
* `200 OK`: `{ "message": "Member added successfully" }`
* `400 Bad Request`: `{"error": "user not found"}` OR `{"error": "user already belongs to a group/company"}`
</details>

<details>
<summary><b><code>DELETE /api/groups/member</code> - Remove Member from Group</b></summary>

**Request Payload:**
```json
{
  "group_id": "uuid-of-the-group",
  "user_id": "uuid-of-the-user"
}
```
**Expected Response (200):** `{ "message": "Member removed successfully" }`
</details>

---

## 🛠 Standard Model Structures (For Typescript Interfaces)

### User Object
```typescript
interface User {
  id: string;
  username: string;
  is_user_in_group: boolean;
  is_user_head: boolean;
  is_deleted: boolean;
  group_id: string | null;
  created_at: string; // ISO 8601 Date
  updated_at: string;
  profile: UserProfile | null;
  address: UserAddress | null;
}

interface UserProfile {
  id: string;
  user_id: string;
  first_name: string;
  middle_name: string;
  last_name: string;
  birth_date: string; // ISO 8601 Date
  phone_number: string;
}

interface UserAddress {
  id: string;
  user_id: string;
  street_name: string;
  baranggay: string;
  town: string;
  province: string;
  zip_code: string;
}
```

### Group Object
```typescript
interface Group {
  id: string;
  name: string;
  owner_id: string;
  is_deleted: boolean;
  created_at: string;
  updated_at: string;
  owner: User | null;
  members: User[] | null;
}
```
