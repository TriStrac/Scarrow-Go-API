package routes

import (
	"github.com/TriStrac/Scarrow-Go-API/internal/api/controllers"
	"github.com/TriStrac/Scarrow-Go-API/internal/api/middlewares"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/gin-gonic/gin"
)

func RegisterMessageRoutes(router *gin.RouterGroup, messageController *controllers.MessageController, userRepo repository.UserRepository) {
	messageRoutes := router.Group("/messages")
	messageRoutes.Use(middlewares.AuthMiddleware(userRepo))
	{
		messageRoutes.GET("/", messageController.GetThreads)
		messageRoutes.GET("/unread-summary", messageController.GetUnreadSummary)
		messageRoutes.GET("/:threadId", messageController.GetThreadMessages)
		messageRoutes.POST("/", messageController.SendMessage)
	}
}
