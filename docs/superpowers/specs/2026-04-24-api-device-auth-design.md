# API Device Auth Design

## Goal

Allow Hub devices to authenticate using their `device_id` + `secret` instead of user JWT tokens.

## Changes Required

### 1. New Middleware: `internal/api/middlewares/device_auth.go`

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

### 2. New Repository Method: `internal/repository/device.go`

Add method to find device by ID:

```go
func (r *deviceRepository) FindByID(id string) (*models.Device, error)
```

(Already exists - verify it returns device with Secret field)

### 3. Update Route: `internal/api/routes/device.go`

Create device-authenticated route for logs:

```go
deviceRoutes.POST("/:deviceId/logs", deviceController.CreateLog)  // change middleware
OR
deviceRoutes.POST("/:deviceId/logs", middlewares.DeviceAuthMiddleware(deviceRepo), deviceController.CreateLog)
```

### Alternative: Separate Device Log Endpoint

Create new route group that uses device auth:

```go
deviceRoutes := router.Group("/device")
deviceRoutes.Use(middlewares.DeviceAuthMiddleware(deviceRepo))
{
    deviceRoutes.POST("/:deviceId/logs", deviceController.CreateLog)
}
```

## Current vs New

| Aspect | Current | New |
|--------|---------|-----|
| Auth | User JWT (AuthMiddleware) | Device ID + Secret (DeviceAuthMiddleware) |
| Context | userID | deviceID |
| Endpoint | Same | Same |

## API Endpoint Summary

```
POST /api/v1/device/:deviceId/logs
Headers:
  - X-Device-ID: <hub_device_id>
  - X-Device-Secret: <hub_secret>
  - Content-Type: application/json

Body:
{
  "node_id": "NODE_xxx",
  "log_type": "DETECTED",
  "pest_type": "BIRD",
  "frequency_hz": 0.0,
  "duration_seconds": 30,
  "payload": "{}"
}

Response: 201 Created
{
  "message": "Log created successfully"
}
```

## Files to Modify

1. Create: `internal/api/middlewares/device_auth.go`
2. Verify: `internal/repository/device.go` - FindByID method exists
3. Modify: `internal/api/routes/device.go` - switch to DeviceAuthMiddleware
4. Modify: `internal/api/controllers/device.go` - add node_id to CreateDeviceLogReq
5. Modify: `internal/service/device.go` - pass node_id to CreateLog

## Testing

- Test with valid device ID + secret → 201
- Test with invalid device ID → 401
- Test with wrong secret → 401
- Test with missing headers → 401