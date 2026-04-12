package routes

import (
	"github.com/TriStrac/Scarrow-Go-API/internal/api/controllers"
	"github.com/TriStrac/Scarrow-Go-API/internal/api/middlewares"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/gin-gonic/gin"
)

func RegisterSubscriptionRoutes(router *gin.RouterGroup, subController *controllers.SubscriptionController, userRepo repository.UserRepository) {
	subRoutes := router.Group("/subscriptions")
	
	// Optional: Public route to just view plans without logging in
	subRoutes.GET("/plans", subController.GetAvailablePlans)

	// Protected routes
	protected := subRoutes.Group("")
	protected.Use(middlewares.AuthMiddleware(userRepo))
	{
		protected.GET("/my", subController.GetMySubscription)
		protected.POST("/checkout", subController.CreateCheckoutSession)
		protected.POST("/verify", subController.VerifyPayment)
	}
}
