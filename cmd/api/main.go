package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/TriStrac/Scarrow-Go-API/docs"
	"github.com/TriStrac/Scarrow-Go-API/internal/api/controllers"
	"github.com/TriStrac/Scarrow-Go-API/internal/api/middlewares"
	"github.com/TriStrac/Scarrow-Go-API/internal/api/routes"
	"github.com/TriStrac/Scarrow-Go-API/internal/config"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/TriStrac/Scarrow-Go-API/internal/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Scarrow API
// @version 1.0
// @description Backend API for Scarrow
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Initialize Database
	config.InitDB()

	// Dependency Injection for User Domain
	userRepo := repository.NewUserRepository(config.DB)
	userService := service.NewUserService(userRepo)
	userController := controllers.NewUserController(userService)

	// Dependency Injection for Group Domain
	groupRepo := repository.NewGroupRepository(config.DB)
	groupService := service.NewGroupService(groupRepo, userRepo)
	groupController := controllers.NewGroupController(groupService)

	// Dependency Injection for ActivityLog Domain
	activityLogRepo := repository.NewActivityLogRepository(config.DB)
	activityLogService := service.NewActivityLogService(activityLogRepo)
	activityLogController := controllers.NewActivityLogController(activityLogService)

	// Dependency Injection for Device Domain
	deviceRepo := repository.NewDeviceRepository(config.DB)
	deviceService := service.NewDeviceService(deviceRepo, userRepo)
	deviceController := controllers.NewDeviceController(deviceService)

	// Initialize Gin router
	router := gin.Default()

	// Simple health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Scarrow-Go-API is running",
		})
	})

	// Setup API Routes
	apiGroup := router.Group("/api")
	// Apply ActivityLogMiddleware to all mutating operations in /api
	apiGroup.Use(middlewares.ActivityLogMiddleware(activityLogService))

	routes.SetupUserRoutes(apiGroup, userController)
	routes.SetupGroupRoutes(apiGroup, groupController)
	routes.RegisterDeviceRoutes(apiGroup, deviceController)
	routes.RegisterActivityLogRoutes(apiGroup, activityLogController)

	// Swagger Route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Configure the HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Run server in a goroutine so it doesn't block
	go func() {
		log.Println("Server is starting on port 8080...")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a 5-second timeout.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting gracefully.")
}
