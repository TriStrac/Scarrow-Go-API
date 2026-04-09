package routes

import (
	"github.com/TriStrac/Scarrow-Go-API/internal/api/controllers"
	"github.com/TriStrac/Scarrow-Go-API/internal/api/middlewares"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(router *gin.RouterGroup, userController *controllers.UserController, userRepo repository.UserRepository) {
	userRoutes := router.Group("/users")
	{
		// Public Routes
		userRoutes.POST("/", userController.Register)
		userRoutes.POST("/verify-registration", userController.VerifyRegistration)
		userRoutes.POST("/login", userController.Login)
		userRoutes.POST("/verify-login", userController.VerifyLogin)
		userRoutes.POST("/forgot-password", userController.ForgotPassword)
		userRoutes.POST("/reset-password", userController.ResetPassword)
		userRoutes.GET("/usernameExists", userController.CheckUsernameExists)

		// Protected Routes (Require JWT + Verified Account)
		protected := userRoutes.Group("")
		protected.Use(middlewares.AuthMiddleware(userRepo))
		{
			protected.GET("/", userController.GetAllUsers)
			protected.GET("/:userId", userController.GetUserByID)
			protected.PATCH("/:userId", userController.UpdateUser)
			protected.POST("/changePassword", userController.ChangePassword)
			protected.PATCH("/:userId/softDelete", userController.SoftDeleteUser)
		}
	}
}
