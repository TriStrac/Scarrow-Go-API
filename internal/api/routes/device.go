package routes

import (
	"github.com/TriStrac/Scarrow-Go-API/internal/api/controllers"
	"github.com/TriStrac/Scarrow-Go-API/internal/api/middlewares"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/gin-gonic/gin"
)

func RegisterDeviceRoutes(router *gin.RouterGroup, deviceController *controllers.DeviceController, userRepo repository.UserRepository, deviceRepo repository.DeviceRepository) {
	// Provisioning endpoints
	hubsRoutes := router.Group("/hubs")
	hubsRoutes.Use(middlewares.AuthMiddleware(userRepo))
	{
		hubsRoutes.POST("/register", deviceController.RegisterHub)
	}

	nodesRoutes := router.Group("/nodes")
	nodesRoutes.Use(middlewares.AuthMiddleware(userRepo))
	{
		nodesRoutes.POST("/register", deviceController.RegisterNode)
	}

	// Device logs - uses device auth
	deviceLogRoutes := router.Group("/device")
	deviceLogRoutes.Use(middlewares.DeviceAuthMiddleware(deviceRepo))
	{
		deviceLogRoutes.POST("/:deviceId/logs", deviceController.CreateLog)
		deviceLogRoutes.GET("/:deviceId/logs", deviceController.GetLogs)
	}

	// Legacy or general device routes
	deviceRoutes := router.Group("/device")
	deviceRoutes.Use(middlewares.AuthMiddleware(userRepo))
	{
		deviceRoutes.POST("/", deviceController.CreateDevice)
		deviceRoutes.GET("/", deviceController.GetAllDevices)
		deviceRoutes.GET("/my", deviceController.GetMyDevices)
		deviceRoutes.GET("/:deviceId", deviceController.GetDeviceByID)
		deviceRoutes.PATCH("/:deviceId", deviceController.UpdateDevice)
		deviceRoutes.DELETE("/:deviceId", deviceController.SoftDeleteDevice)
	}
}
