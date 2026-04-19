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

	"github.com/TriStrac/Scarrow-Go-API/internal/api/controllers"
	"github.com/TriStrac/Scarrow-Go-API/internal/api/middlewares"
	"github.com/TriStrac/Scarrow-Go-API/internal/api/routes"
	"github.com/TriStrac/Scarrow-Go-API/internal/config"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/TriStrac/Scarrow-Go-API/internal/service"
	"github.com/TriStrac/Scarrow-Go-API/pkg/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Database
	config.InitDB()

	// Core Repositories
	userRepo := repository.NewUserRepository(config.DB)
	deviceRepo := repository.NewDeviceRepository(config.DB)
	groupRepo := repository.NewGroupRepository(config.DB)
	otpRepo := repository.NewOTPRepository(config.DB)
	notificationRepo := repository.NewNotificationRepository(config.DB)
	activityLogRepo := repository.NewActivityLogRepository(config.DB)
	messageRepo := repository.NewMessageRepository(config.DB)
	invitationRepo := repository.NewGroupInvitationRepository(config.DB)
	subRepo := repository.NewSubscriptionRepository(config.DB)

	// Utils
	smsApiKey := os.Getenv("SEMAPHORE_API_KEY")
	if smsApiKey == "" {
		log.Println("WARNING: SEMAPHORE_API_KEY is not set. SMS sending might fail.")
	}
	smsService := utils.NewRealSmsService(smsApiKey)

	// Services
	otpService := service.NewOTPService(otpRepo, smsService)
	notificationService := service.NewNotificationService(notificationRepo)
	userService := service.NewUserService(userRepo, deviceRepo, messageRepo)
	groupService := service.NewGroupService(groupRepo, userRepo, deviceRepo, notificationService, invitationRepo)
	deviceService := service.NewDeviceService(deviceRepo, userRepo)
	activityLogService := service.NewActivityLogService(activityLogRepo)
	messageService := service.NewMessageService(messageRepo, userRepo)
	reportService := service.NewReportService(deviceRepo, userRepo)
	subService := service.NewSubscriptionService(subRepo, userRepo)

	// Controllers
	userController := controllers.NewUserController(userService, otpService)
	groupController := controllers.NewGroupController(groupService)
	deviceController := controllers.NewDeviceController(deviceService)
	activityLogController := controllers.NewActivityLogController(activityLogService)
	notificationController := controllers.NewNotificationController(notificationService)
	messageController := controllers.NewMessageController(messageService)
	reportController := controllers.NewReportController(reportService)
	subController := controllers.NewSubscriptionController(subService)

	// Initialize Gin router
	router := gin.Default()

	// Simple health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Scarrow-Go-API is running",
			"version": "1.7.1",
		})
	})

	// Setup API Routes
	apiGroup := router.Group("/api")
	// Apply ActivityLogMiddleware to all mutating operations in /api
	apiGroup.Use(middlewares.ActivityLogMiddleware(activityLogService))

	routes.SetupUserRoutes(apiGroup, userController, userRepo)
	routes.SetupGroupRoutes(apiGroup, groupController, userRepo)
	routes.RegisterDeviceRoutes(apiGroup, deviceController, userRepo)
	routes.RegisterActivityLogRoutes(apiGroup, activityLogController, userRepo)
	routes.RegisterNotificationRoutes(apiGroup, notificationController, userRepo)
	routes.RegisterMessageRoutes(apiGroup, messageController, userRepo)
	routes.RegisterReportRoutes(apiGroup, reportController, userRepo)
	routes.RegisterSubscriptionRoutes(apiGroup, subController, userRepo)

	// Get Port from env
	port := os.Getenv("PORT")
	if port == "" {
		port = "38192"
	}

	// Configure the HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Run server in a goroutine so it doesn't block
	go func() {
		log.Printf("Server is starting on port %s...\n", port)
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
