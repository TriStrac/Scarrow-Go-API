# API Refinement & Integration Plan (10/04/26)

This document outlines the roadmap for refining the Scarrow-Go-API based on the integrated review for Web and Mobile parity.

## Phase 1: Minimal Registration (High Priority)
- [x] **Controller/DTO**: Update `RegisterReq` to only require `username`, `password`, and `number`.
- [x] **Service**: Adjust `Register` logic to handle missing `FirstName`/`LastName` (default to empty or username).
- [x] **Model**: Ensure `UserProfile` allows null/empty fields for initial registration.

## Phase 2: Session & Identity (Category B.1)
- [x] **Endpoint**: `GET /api/users/me`
- [x] **Logic**: Return expanded user state (role, group info, subscription status, `profile_complete` flag).
- [x] **Purpose**: Centralize frontend routing/UI logic based on a single "me" call.

## Phase 3: Group & Organization Management (Category B.2, B.3, B.4)
- [x] **Group Detail**: `GET /api/groups/:groupId` (Settings, role, member count).
- [x] **Member List**: `GET /api/groups/:groupId/members` (Crucial for Dashboard/Farmers list).
- [x] **Join Workflow**: Implement Request/Approve/Decline logic (if not using simple codes). *Note: System uses simple codes, already robust.*
- [x] **Leave Group**: `POST /api/groups/leave`.

## Phase 4: Device Lifecycle & History (Category B.5)
- [x] **CRUD Operations**: Implement `GET`, `PATCH`, `DELETE` (unpair) for individual devices. *Note: Already existed and REST-compliant.*
- [x] **Logs/History**: `GET /api/device/:deviceId/logs` with pagination for historical analysis.

## Phase 5: Messaging & Notifications (Category B.6)
- [x] **Pagination**: Add `limit`/`offset` or cursor-based pagination to `GET /api/messages/`.
- [x] **Badges**: Implement `unread_count` summary endpoint.

## Phase 6: Analytics & Infrastructure (Category B.7, B.9)
- [x] **Reports**: `GET /api/reports/summary` (Aggregated sensor data for charts).
- [x] **Push Tokens**: `POST /api/users/me/push-tokens` (FCM integration for Mobile).

## Phase 7: Subscriptions (Category B.8)
- [x] **Placeholder**: `GET /api/subscriptions/plans`.
- [x] **Note**: Architecture implemented (Models, Repo, Service, Controller). Logic for Premium Features is mocked and ready for real payment gateway injection.

---
**Status Symbols:**
- ⏳ Pending
- 🚧 In Progress
- ✅ Completed
