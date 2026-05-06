package routes

import (
	"github.com/TriStrac/Scarrow-Go-API/internal/api/controllers"
	"github.com/TriStrac/Scarrow-Go-API/internal/api/middlewares"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/gin-gonic/gin"
)

func RegisterReportRoutes(router *gin.RouterGroup, reportController *controllers.ReportController, userRepo repository.UserRepository) {
	reportRoutes := router.Group("/reports")
	reportRoutes.Use(middlewares.AuthMiddleware(userRepo))
	{
		reportRoutes.GET("/summary", reportController.GetSummary)
		reportRoutes.GET("/hub/:hubId", reportController.GetHubReport)
	}
}
