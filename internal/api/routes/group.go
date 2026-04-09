package routes

import (
	"github.com/TriStrac/Scarrow-Go-API/internal/api/controllers"
	"github.com/TriStrac/Scarrow-Go-API/internal/api/middlewares"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/gin-gonic/gin"
)

func SetupGroupRoutes(router *gin.RouterGroup, groupController *controllers.GroupController, userRepo repository.UserRepository) {
	groupRoutes := router.Group("/groups")

	// All group routes are protected and require JWT + Verified Account
	groupRoutes.Use(middlewares.AuthMiddleware(userRepo))
	{
		groupRoutes.POST("/", groupController.CreateGroup)
		groupRoutes.GET("/", groupController.GetAllGroups)
		groupRoutes.GET("/owner", groupController.GetGroupsByOwner)
		groupRoutes.GET("/:groupId", groupController.GetGroupByID)
		groupRoutes.PATCH("/:groupId", groupController.UpdateGroup)
		groupRoutes.PATCH("/:groupId/softDelete", groupController.SoftDeleteGroup)
		groupRoutes.POST("/member", groupController.AddMember)
		groupRoutes.DELETE("/member", groupController.RemoveMember)
		groupRoutes.GET("/:groupId/members", groupController.GetGroupMembers)
		groupRoutes.POST("/:groupId/invite", groupController.CreateInvitation)
		groupRoutes.POST("/join", groupController.JoinGroupByCode)
	}
}
