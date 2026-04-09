package routes

import (
	"github.com/TriStrac/Scarrow-Go-API/internal/api/controllers"
	"github.com/TriStrac/Scarrow-Go-API/internal/api/middlewares"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/gin-gonic/gin"
)

func RegisterNotificationRoutes(router *gin.RouterGroup, notificationController *controllers.NotificationController, userRepo repository.UserRepository) {
	notificationRoutes := router.Group("/notifications")
	notificationRoutes.Use(middlewares.AuthMiddleware(userRepo))
	{
		notificationRoutes.GET("/my", notificationController.GetMyNotifications)
		notificationRoutes.PATCH("/:notificationId/read", notificationController.MarkAsRead)
		notificationRoutes.PATCH("/read-all", notificationController.MarkAllAsRead)
	}
}
