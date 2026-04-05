# ScarrowAPI Migration Plan: Node.js/Firestore to Go/MySQL

## 1. Overview
This document serves as the master blueprint and state-tracker for migrating the `ScarrowAPI` backend from **Node.js (Express) + Firebase (Firestore)** to **Go + MySQL**. 

**Any AI Agent working on this project MUST read this file first to assess the current state before generating code.**

---

## 2. Engineering Standards & Best Practices
To ensure the resulting Go codebase is **production-level, clean, and scalable**, all code written during this migration MUST adhere to the following standards:

*   **Production Level Readiness:** Implement graceful server shutdowns, robust logging (e.g., using `slog` or `zap`), proper environment-variable configuration mapping, and standardized HTTP JSON error responses. Never swallow errors; always handle or return them explicitly.
*   **Clean & Scalable Architecture:** Strictly enforce Separation of Concerns. Code must be organized cleanly into distinct layers:
    *   **Controllers (`internal/api/controllers/`):** Strictly handle HTTP requests/responses, payload validation, and route directly to Services. No business logic or DB calls here.
    *   **Services (`internal/service/`):** Contain all core business logic.
    *   **Repositories (`internal/repository/`):** Exclusive handlers for MySQL/GORM queries.
*   **Avoid "Spaghettified" Code:** Keep functions small, modular, and single-purpose. Prevent deep nesting. Use interfaces (`type Repository interface {}`) to decouple layers and make them easily testable/mockable. Prevent circular dependencies.
*   **Teamworkable & Handoverable:** Write explicit, highly readable code over "clever" or overly concise one-liners. Variable and function names must be descriptive and follow standard Go conventions (Effective Go). Provide clear comments for complex business logic. Maintain up-to-date Swagger documentation (`swaggo/swag`) for every endpoint created.
*   **Code Readability & Optimization:** Leverage Go's strengths. Use database connection pooling, optimize GORM queries (avoid N+1 query problems via Preloading/Joins), and use standard Go idioms. Ensure the codebase passes standard `go fmt`, `go vet`, and `golangci-lint` checks.

---

## 3. Technology Stack Mapping

| Concept | Current (Node.js) | Target (Go) |
| :--- | :--- | :--- |
| **Language** | TypeScript | Go (1.21+) |
| **Web Framework** | Express.js | `gin-gonic/gin` (Provides similar routing/middleware feel) |
| **Database** | Firestore (NoSQL) | MySQL 8.x (Relational) |
| **DB ORM/Driver** | Firebase Admin SDK | `gorm` (for rapid dev) OR `database/sql` + `sqlx` |
| **Validation** | Zod | `go-playground/validator/v10` |
| **Authentication** | `jsonwebtoken` | `golang-jwt/jwt/v5` |
| **Password Hashing** | `bcryptjs` | `golang.org/x/crypto/bcrypt` |
| **Environment Vars** | `dotenv` | `joho/godotenv` |

---

## 4. Database Schema Migration (NoSQL to Relational)

In Firestore, data was heavily denormalized and spread across disconnected documents (e.g., Users, Profiles, Addresses). In MySQL, we will use a normalized relational schema with Foreign Keys.

### 4.1 `users` Table
*   `id` (VARCHAR/UUID, Primary Key)
*   `username` (VARCHAR, Unique, Indexed)
*   `password` (VARCHAR)
*   `is_user_in_group` (BOOLEAN, Default: false)
*   `is_user_head` (BOOLEAN, Default: false)
*   `is_deleted` (BOOLEAN, Default: false)
*   `created_at` (TIMESTAMP)
*   `updated_at` (TIMESTAMP)
*   `deleted_at` (TIMESTAMP, Nullable)

### 4.2 `user_profiles` Table (1:1 with users)
*   `id` (VARCHAR/UUID, Primary Key)
*   `user_id` (VARCHAR/UUID, Foreign Key -> `users.id`)
*   `first_name` (VARCHAR)
*   `middle_name` (VARCHAR, Nullable)
*   `last_name` (VARCHAR)
*   `birth_date` (DATE)
*   `phone_number` (VARCHAR)

### 4.3 `user_addresses` Table (1:1 with users)
*   `id` (VARCHAR/UUID, Primary Key)
*   `user_id` (VARCHAR/UUID, Foreign Key -> `users.id`)
*   `street_name` (VARCHAR)
*   `baranggay` (VARCHAR)
*   `town` (VARCHAR)
*   `province` (VARCHAR)
*   `zip_code` (VARCHAR)

### 4.4 `groups` Table
*   `id` (VARCHAR/UUID, Primary Key)
*   `name` (VARCHAR, Unique)
*   `owner_id` (VARCHAR/UUID, Foreign Key -> `users.id`)
*   `is_deleted` (BOOLEAN)
*   `created_at` (TIMESTAMP)
*   `updated_at` (TIMESTAMP)

### 4.5 `group_members` Table (Many-to-Many)
*   `group_id` (VARCHAR/UUID, Foreign Key -> `groups.id`)
*   `user_id` (VARCHAR/UUID, Foreign Key -> `users.id`)
*   `joined_at` (TIMESTAMP)
*   *(Primary Key is composite: `group_id`, `user_id`)*

### 4.6 Devices & Logs Tables
*   `devices`: `id`, `name`, `status`, `created_at`, `updated_at`, `is_deleted`
*   `device_owners`: `device_id`, `owner_id` (user or group), `owner_type` ('USER' | 'GROUP')
*   `device_logs`: `id`, `device_id`, `log_type`, `payload`, `created_at`
*   `user_activity_logs`: `id`, `user_id`, `action`, `module`, `created_at`

---

## 5. API Endpoints Map

All routes prefix: `/api`

### Users (`/users`)
- [ ] `POST /` - Register
- [ ] `POST /login` - Login (Username + Password)
- [ ] `GET /` - Get all users
- [ ] `GET /:userId` - Get user by ID (Joins profile & address)
- [ ] `PATCH /:userId` - Update user/profile/address
- [ ] `POST /changePassword` - Change Password
- [ ] `PATCH /:userId/softDelete` - Soft delete
- [ ] `GET /usernameExists` - Check if username is taken

### Groups (`/groups`)
- [ ] `POST /` - Create Group
- [ ] `GET /` - Get all groups
- [ ] `GET /owner` - Get groups by owner
- [ ] `GET /:groupId` - Get group info
- [ ] `PATCH /:groupId` - Update group
- [ ] `PATCH /:groupId/softDelete` - Delete group
- [ ] `POST /member` - Add member (by username)
- [ ] `DELETE /member` - Remove member
- [ ] `GET /:groupId/members` - List members

### Devices (`/device`) & Logs (`/deviceLogs`, `/userActivityLog`)
- [ ] *Pending implementation based on existing node routes.*

---

## 6. Step-by-Step Execution Plan

**AI Agent Instructions:** When asked to start or continue, find the first unchecked `[ ]` step and execute it. After successful completion, mark it with `[x]`. 
**CRITICAL:** At the end of EVERY phase, you MUST:
1. Run `go fmt ./...` and `go mod tidy`.
2. Commit your changes with a descriptive, conventional commit message.
3. Push the changes to the remote repository.
4. Wait for the next prompt.

### Phase 1: Go Project Initialization & Skeleton
- [x] Initialize Go Module (`go mod init github.com/TriStrac/ScarrowAPI-Go`).
- [x] Setup standard Go project structure (`cmd/`, `internal/api/`, `internal/models/`, `internal/repository/`, `internal/service/`, `pkg/`).
- [x] Install dependencies (Gin, Gorm/MySQL, JWT, Crypto, Godotenv).
- [x] Create `main.go`, setup Graceful Shutdown, and configure basic Gin router.

### Phase 2: Database & Configuration Setup
- [x] Create database configuration file and connection logic with pooling (`internal/config/db.go`).
- [x] Write GORM models (Entities) reflecting the Relational Schema (Section 4), ensuring proper struct tags and foreign keys.
- [x] Setup database Auto-Migration on startup.

### Phase 3: Core Domain Implementation (Users)
- [ ] Create User Repository Interfaces and implementations (`internal/repository/user.go`).
- [ ] Create User Service Interfaces and implementations (`internal/service/user.go`) for business logic (hashing, JWT).
- [ ] Create User Controller (`internal/api/controllers/user.go`) with input validation.
- [ ] Implement JWT Auth Middleware (`internal/api/middlewares/auth.go`).
- [ ] Map and test User Routes (`internal/api/routes/user.go`).

### Phase 4: Core Domain Implementation (Groups)
- [ ] Create Group Repository & Service (Interface-driven).
- [ ] Create Group Controller & Routes.
- [ ] Ensure `addGroupMember` correctly queries by `username` and utilizes MySQL relationships.

### Phase 5: Devices & Logging Implementation
- [ ] Replicate Device creation, status, and ownership logic using relational tables.
- [ ] Replicate Activity Logger Middleware.
- [ ] Implement Device and Activity Log routes.

### Phase 6: Quality Assurance & Finalization
- [ ] Replicate Swagger Documentation (using `swaggo/swag`).
- [ ] Run `go fmt` and linter checks to guarantee readability.
- [ ] Final architecture review to ensure no circular dependencies and clean handover state.

---
**Current Status:** Phase 2 Complete. Ready for Phase 3 (Core Domain: Users).