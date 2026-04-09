package routes

import (
	"github.com/TriStrac/Scarrow-Go-API/internal/api/controllers"
	"github.com/TriStrac/Scarrow-Go-API/internal/api/middlewares"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/gin-gonic/gin"
)

func RegisterDeviceRoutes(router *gin.RouterGroup, deviceController *controllers.DeviceController, userRepo repository.UserRepository) {
	deviceRoutes := router.Group("/device")
	deviceRoutes.Use(middlewares.AuthMiddleware(userRepo))
	{
		deviceRoutes.POST("/", deviceController.CreateDevice)
		deviceRoutes.GET("/", deviceController.GetAllDevices)
		deviceRoutes.GET("/my", deviceController.GetMyDevices)
		deviceRoutes.GET("/:deviceId", deviceController.GetDeviceByID)
		deviceRoutes.PATCH("/:deviceId", deviceController.UpdateDevice)
		deviceRoutes.DELETE("/:deviceId", deviceController.SoftDeleteDevice)

		// Ownership
		deviceRoutes.POST("/:deviceId/owner", deviceController.AddOwner)
		deviceRoutes.DELETE("/:deviceId/owner", deviceController.RemoveOwner)
		deviceRoutes.GET("/:deviceId/owners", deviceController.GetOwners)

		// Logs
		deviceRoutes.POST("/:deviceId/logs", deviceController.CreateLog)
		deviceRoutes.GET("/:deviceId/logs", deviceController.GetLogs)
	}
}
