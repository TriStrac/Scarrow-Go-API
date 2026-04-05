package controllers

import (
	"net/http"

	"github.com/TriStrac/Scarrow-Go-API/internal/service"
	"github.com/gin-gonic/gin"
)

type GroupController struct {
	groupService service.GroupService
}

func NewGroupController(groupService service.GroupService) *GroupController {
	return &GroupController{groupService: groupService}
}

type CreateGroupReq struct {
	Name string `json:"name" binding:"required"`
}

type UpdateGroupReq struct {
	Name string `json:"name" binding:"required"`
}

type AddMemberReq struct {
	GroupID  string `json:"group_id" binding:"required"`
	Username string `json:"username" binding:"required"`
}

type RemoveMemberReq struct {
	GroupID string `json:"group_id" binding:"required"`
	UserID  string `json:"user_id" binding:"required"`
}

func (c *GroupController) CreateGroup(ctx *gin.Context) {
	ownerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateGroupReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := c.groupService.CreateGroup(req.Name, ownerID.(string))
	if err != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Group created successfully", "group": group})
}

func (c *GroupController) GetAllGroups(ctx *gin.Context) {
	groups, err := c.groupService.GetAllGroups()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"groups": groups})
}

func (c *GroupController) GetGroupsByOwner(ctx *gin.Context) {
	ownerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	groups, err := c.groupService.GetGroupsByOwner(ownerID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"groups": groups})
}

func (c *GroupController) GetGroupByID(ctx *gin.Context) {
	groupID := ctx.Param("groupId")
	group, err := c.groupService.GetGroupByID(groupID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"group": group})
}

func (c *GroupController) UpdateGroup(ctx *gin.Context) {
	groupID := ctx.Param("groupId")
	var req UpdateGroupReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.groupService.UpdateGroup(groupID, req.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Group updated successfully"})
}

func (c *GroupController) SoftDeleteGroup(ctx *gin.Context) {
	groupID := ctx.Param("groupId")
	err := c.groupService.SoftDeleteGroup(groupID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}

func (c *GroupController) AddMember(ctx *gin.Context) {
	var req AddMemberReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.groupService.AddMemberByUsername(req.GroupID, req.Username)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Member added successfully"})
}

func (c *GroupController) RemoveMember(ctx *gin.Context) {
	var req RemoveMemberReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.groupService.RemoveMember(req.GroupID, req.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}

func (c *GroupController) GetGroupMembers(ctx *gin.Context) {
	groupID := ctx.Param("groupId")
	members, err := c.groupService.GetGroupMembers(groupID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"members": members})
}
