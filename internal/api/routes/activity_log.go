package routes

import (
	"github.com/TriStrac/Scarrow-Go-API/internal/api/controllers"
	"github.com/TriStrac/Scarrow-Go-API/internal/api/middlewares"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/gin-gonic/gin"
)

func RegisterActivityLogRoutes(r *gin.RouterGroup, controller *controllers.ActivityLogController, userRepo repository.UserRepository) {
	logGroup := r.Group("/activityLogs")
	logGroup.Use(middlewares.AuthMiddleware(userRepo))
	{
		logGroup.GET("/", controller.GetAllLogs)
		logGroup.GET("/my", controller.GetMyLogs)
	}
}
