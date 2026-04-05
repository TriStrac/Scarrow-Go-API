package routes

import (
	"github.com/TriStrac/Scarrow-Go-API/internal/api/controllers"
	"github.com/TriStrac/Scarrow-Go-API/internal/api/middlewares"
	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(router *gin.RouterGroup, userController *controllers.UserController) {
	userRoutes := router.Group("/users")
	{
		// Public Routes
		userRoutes.POST("/", userController.Register)
		userRoutes.POST("/login", userController.Login)
		userRoutes.GET("/usernameExists", userController.CheckUsernameExists)

		// Protected Routes (Require JWT)
		protected := userRoutes.Group("")
		protected.Use(middlewares.AuthMiddleware())
		{
			protected.GET("/", userController.GetAllUsers)
			protected.GET("/:userId", userController.GetUserByID)
			protected.PATCH("/:userId", userController.UpdateUser)
			protected.POST("/changePassword", userController.ChangePassword)
			protected.PATCH("/:userId/softDelete", userController.SoftDeleteUser)
		}
	}
}
