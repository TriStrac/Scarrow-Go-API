# API Device Auth Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow Hub devices to authenticate using device_id + secret instead of user JWT tokens, so Pi can send node logs to API.

**Architecture:** New middleware reads X-Device-ID + X-Device-Secret headers, validates against DB, sets deviceID in context. Log endpoint accepts node_id in body.

**Tech Stack:** Go, Gin, GORM, SQLite

---

### Task 1: Create Device Auth Middleware

**Files:**
- Create: `internal/api/middlewares/device_auth.go`

- [ ] **Step 1: Create the middleware file**

```go
package middlewares

import (
	"net/http"

	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/gin-gonic/gin"
)

func DeviceAuthMiddleware(deviceRepo repository.DeviceRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceID := c.GetHeader("X-Device-ID")
		deviceSecret := c.GetHeader("X-Device-Secret")

		if deviceID == "" || deviceSecret == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "X-Device-ID and X-Device-Secret headers required"})
			c.Abort()
			return
		}

		device, err := deviceRepo.FindByID(deviceID)
		if err != nil || device == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid device"})
			c.Abort()
			return
		}

		if device.Secret != deviceSecret {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid secret"})
			c.Abort()
			return
		}

		c.Set("deviceID", deviceID)
		c.Next()
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd D:\Codes\Scarrow-Go-API && go build ./...`

Expected: No errors

---

### Task 2: Add node_id to CreateDeviceLogReq

**Files:**
- Modify: `internal/api/controllers/device.go:43-49`

- [ ] **Step 1: Add NodeID field to request struct**

```go
type CreateDeviceLogReq struct {
	NodeID           string  `json:"node_id"`
	LogType         string  `json:"log_type" binding:"required"`
	Payload         string  `json:"payload" binding:"required"`
	PestType        string  `json:"pest_type"`
	FrequencyHz     float64 `json:"frequency_hz"`
	DurationSeconds int     `json:"duration_seconds"`
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd D:\Codes\Scarrow-Go-API && go build ./...`

Expected: No errors

---

### Task 3: Update Service to Accept NodeID

**Files:**
- Modify: `internal/service/device.go:161-172`
- Modify: `internal/repository/device.go`

- [ ] **Step 1: Update service interface and method**

In `internal/service/device.go`, update:

```go
// Interface
CreateLog(deviceID, nodeID, logType, payload, pestType string, freq float64, duration int) error

// Implementation
func (s *deviceService) CreateLog(deviceID, nodeID, logType, payload, pestType string, freq float64, duration int) error {
	targetDeviceID := deviceID
	if nodeID != "" {
		targetDeviceID = nodeID  // Log belongs to node, not hub
	}

	log := &models.DeviceLog{
		ID:                uuid.New().String(),
		DeviceID:          targetDeviceID,
		LogType:           logType,
		Payload:          payload,
		PestType:          pestType,
		FrequencyHz:        freq,
		DurationSeconds:   duration,
	}
	return s.repo.CreateLog(log)
}
```

- [ ] **Step 2: Update controller to pass nodeID to service**

In `internal/api/controllers/device.go:231`:

```go
err = c.deviceService.CreateLog(deviceID, req.NodeID, req.LogType, req.Payload, req.PestType, req.FrequencyHz, req.DurationSeconds)
```

- [ ] **Step 3: Verify it compiles**

Run: `cd D:\Codes\Scarrow-Go-API && go build ./...`

Expected: No errors

---

### Task 4: Update Route to Use Device Auth

**Files:**
- Modify: `internal/api/routes/device.go`

- [ ] **Step 1: Switch log endpoint to device auth**

In `internal/api/routes/device.go`, change:

```go
// Current (user auth)
deviceRoutes.POST("/:deviceId/logs", deviceController.CreateLog)

// New (device auth) - replace the entire route group setup
// Option A: Add to existing deviceRoutes with device auth
deviceRoutes.POST("/:deviceId/logs", middlewares.DeviceAuthMiddleware(deviceRepo), deviceController.CreateLog)

// Option B: Create separate route group that uses device auth (recommended for clarity)
deviceRoutes := router.Group("/device")
deviceRoutes.Use(middlewares.DeviceAuthMiddleware(deviceRepo))
{
	deviceRoutes.POST("/:deviceId/logs", deviceController.CreateLog)
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd D:\Codes\Scarrow-Go-API && go build ./...`

Expected: No errors

---

### Task 5: Testing

**Files:**
- Manual test using curl or Postman

- [ ] **Step 1: Test with invalid device ID**

```bash
curl -X POST http://localhost:8080/api/v1/device/HUB_xxx/logs \
  -H "X-Device-ID: INVALID" \
  -H "X-Device-Secret: wrong" \
  -H "Content-Type: application/json" \
  -d '{"log_type": "DETECTED", "pest_type": "BIRD", "payload": "{}", "duration_seconds": 30}'
```

Expected: `401 {"error": "Invalid device"}`

- [ ] **Step 2: Test with valid device credentials**

Create test device in DB first, then:

```bash
curl -X POST http://localhost:8080/api/v1/device/HUB_test/logs \
  -H "X-Device-ID: HUB_test" \
  -H "X-Device-Secret: <secret_from_db>" \
  -H "Content-Type: application/json" \
  -d '{"node_id": "NODE_test", "log_type": "DETECTED", "pest_type": "BIRD", "payload": "{}", "duration_seconds": 30}'
```

Expected: `201 {"message": "Log created successfully"}`

---

### Task 6: Commit

- [ ] **Step 1: Commit all changes**

```bash
git add internal/api/middlewares/device_auth.go internal/api/controllers/device.go internal/service/device.go internal/api/routes/device.go
git commit -m "feat: add device auth middleware for IoT log ingestion"
```

---

## Summary

| Task | Description | Files |
|------|------------|-------|
| 1 | Device auth middleware | Create: `middlewares/device_auth.go` |
| 2 | Add node_id req | Modify: `controllers/device.go` |
| 3 | Service nodeID | Modify: `service/device.go` |
| 4 | Route device auth | Modify: `routes/device.go` |
| 5 | Manual test | - |
| 6 | Commit | - |