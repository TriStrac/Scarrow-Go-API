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

// CreateGroup godoc
// @Summary CreateGroup
// @Description CreateGroup endpoint
// @Tags groups
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /groups/creategroup [post]
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

// GetAllGroups godoc
// @Summary GetAllGroups
// @Description GetAllGroups endpoint
// @Tags groups
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /groups/getallgroups [post]
func (c *GroupController) GetAllGroups(ctx *gin.Context) {
	groups, err := c.groupService.GetAllGroups()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"groups": groups})
}

// GetGroupsByOwner godoc
// @Summary GetGroupsByOwner
// @Description GetGroupsByOwner endpoint
// @Tags groups
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /groups/getgroupsbyowner [post]
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

// GetGroupByID godoc
// @Summary GetGroupByID
// @Description GetGroupByID endpoint
// @Tags groups
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /groups/getgroupbyid [post]
func (c *GroupController) GetGroupByID(ctx *gin.Context) {
	groupID := ctx.Param("groupId")
	group, err := c.groupService.GetGroupByID(groupID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"group": group})
}

// UpdateGroup godoc
// @Summary UpdateGroup
// @Description UpdateGroup endpoint
// @Tags groups
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /groups/updategroup [post]
func (c *GroupController) UpdateGroup(ctx *gin.Context) {
	groupID := ctx.Param("groupId")
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	group, err := c.groupService.GetGroupByID(groupID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Authorization Check: Only group owner can rename
	if group.OwnerID != callerID.(string) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Only the group owner can modify the group"})
		return
	}

	var req UpdateGroupReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = c.groupService.UpdateGroup(groupID, req.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Group updated successfully"})
}

// SoftDeleteGroup godoc
// @Summary SoftDeleteGroup
// @Description SoftDeleteGroup endpoint
// @Tags groups
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /groups/softdeletegroup [post]
func (c *GroupController) SoftDeleteGroup(ctx *gin.Context) {
	groupID := ctx.Param("groupId")
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	group, err := c.groupService.GetGroupByID(groupID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Authorization Check: Only group owner can delete
	if group.OwnerID != callerID.(string) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Only the group owner can delete the group"})
		return
	}

	err = c.groupService.SoftDeleteGroup(groupID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}

// AddMember godoc
// @Summary AddMember
// @Description AddMember endpoint
// @Tags groups
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /groups/addmember [post]
func (c *GroupController) AddMember(ctx *gin.Context) {
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req AddMemberReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := c.groupService.GetGroupByID(req.GroupID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Authorization Check: Only group owner can add members
	if group.OwnerID != callerID.(string) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Only the group owner can add members"})
		return
	}

	err = c.groupService.AddMemberByUsername(req.GroupID, req.Username)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Member added successfully"})
}

// RemoveMember godoc
// @Summary RemoveMember
// @Description RemoveMember endpoint
// @Tags groups
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /groups/removemember [post]
func (c *GroupController) RemoveMember(ctx *gin.Context) {
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req RemoveMemberReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := c.groupService.GetGroupByID(req.GroupID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Authorization Check: Only group owner OR the member themselves can remove
	if group.OwnerID != callerID.(string) && req.UserID != callerID.(string) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Only the group owner or the member themselves can remove a member"})
		return
	}

	err = c.groupService.RemoveMember(req.GroupID, req.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}

// GetGroupMembers godoc
// @Summary GetGroupMembers
// @Description GetGroupMembers endpoint
// @Tags groups
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /groups/getgroupmembers [post]
func (c *GroupController) GetGroupMembers(ctx *gin.Context) {
	groupID := ctx.Param("groupId")
	members, err := c.groupService.GetGroupMembers(groupID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"members": members})
}
