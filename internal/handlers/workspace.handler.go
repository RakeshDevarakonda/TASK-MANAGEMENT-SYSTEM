package handlers

import (
	"net/http"
	"strconv"
	"task-system/internal/models"
	"task-system/internal/service"
	"task-system/internal/utils"

	"github.com/gin-gonic/gin"
)

type WorkspaceHandler struct {
	service *service.WorkspaceService
}

func NewWorkspaceHandler(src *service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{service: src}
}

func (W *WorkspaceHandler) Create(c *gin.Context) {
	var req models.CreateWorkspace
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationError(c, err)
		return
	}

	currentUser := c.MustGet("currentUser").(*models.User)

	workspace, err := W.service.CreateWorkspace(req.WorkspaceName, req.Description, currentUser.ID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "workspace created successfully", workspace)
}

func (W *WorkspaceHandler) Invite(c *gin.Context) {
	var req models.InviteToWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationError(c, err)
		return
	}

	currentUser := c.MustGet("currentUser").(*models.User)

	invitation, err := W.service.InviteUser(req.WorkspaceID, req.Email, currentUser.ID, req.Role)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "user invited successfully", invitation)
}

func (W *WorkspaceHandler) Join(c *gin.Context) {
	var req models.JoinWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationError(c, err)
		return
	}

	currentUser := c.MustGet("currentUser").(*models.User)

	err := W.service.JoinWorkspace(req.WorkspaceID, currentUser.Email, currentUser.ID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "joined workspace successfully", nil)
}

// Get retrieves workspace by ID
func (W *WorkspaceHandler) Get(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("currentUser").(*models.User)

	workspace, err := W.service.GetWorkspace(id, currentUser.ID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "workspace retrieved successfully", workspace)
}

// List lists all workspaces for a user
func (W *WorkspaceHandler) List(c *gin.Context) {
	currentUser := c.MustGet("currentUser").(*models.User)

	workspaces, err := W.service.ListWorkspaces(currentUser.ID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "workspaces retrieved successfully", workspaces)
}

// Update updates workspace properties
func (W *WorkspaceHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationError(c, err)
		return
	}

	currentUser := c.MustGet("currentUser").(*models.User)

	workspace, err := W.service.UpdateWorkspace(id, req, currentUser.ID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "workspace updated successfully", workspace)
}

// Delete removes a workspace
func (W *WorkspaceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("currentUser").(*models.User)

	err := W.service.DeleteWorkspace(id, currentUser.ID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "workspace deleted successfully", nil)
}

// ListMembers lists all members in a workspace with cursor pagination
func (W *WorkspaceHandler) ListMembers(c *gin.Context) {
	workspaceID := c.Param("id")
	currentUser := c.MustGet("currentUser").(*models.User)

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "20")
	cursorStr := c.Query("cursor")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var cursorTime int64
	var cursorID string
	if cursorStr != "" {
		cursorTime, cursorID, err = utils.DecodeCursor(cursorStr)
		if err != nil {
			utils.SendError(c, http.StatusBadRequest, "invalid cursor parameter", nil)
			return
		}
	}

	members, err := W.service.ListWorkspaceMembers(workspaceID, currentUser.ID, limit, cursorTime, cursorID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	var nextCursor string
	if len(members) > limit {
		lastMember := members[limit-1]
		nextCursor = utils.EncodeCursor(lastMember.CreatedAt, lastMember.ID)
		members = members[:limit]
	}

	utils.SendCursorSuccess(c, http.StatusOK, "members retrieved successfully", members, limit, nextCursor)
}

// UpdateMemberRole changes the role of a workspace member (owner only)
func (W *WorkspaceHandler) UpdateMemberRole(c *gin.Context) {
	workspaceID := c.Param("id")
	targetUserID := c.Param("user_id")

	var req models.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationError(c, err)
		return
	}

	currentUser := c.MustGet("currentUser").(*models.User)

	err := W.service.UpdateWorkspaceMemberRole(workspaceID, targetUserID, req.Role, currentUser.ID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "member role updated successfully", nil)
}

// KickMember removes a member from the workspace (owner only)
func (W *WorkspaceHandler) KickMember(c *gin.Context) {
	workspaceID := c.Param("id")
	targetUserID := c.Param("user_id")
	currentUser := c.MustGet("currentUser").(*models.User)

	err := W.service.RemoveWorkspaceMember(workspaceID, targetUserID, currentUser.ID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "member kicked successfully", nil)
}

// Leave removes the active user from the workspace
func (W *WorkspaceHandler) Leave(c *gin.Context) {
	workspaceID := c.Param("id")
	currentUser := c.MustGet("currentUser").(*models.User)

	err := W.service.LeaveWorkspace(workspaceID, currentUser.ID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "left workspace successfully", nil)
}

// TransferOwnership transfers the super_admin role to another workspace member (owner only)
func (W *WorkspaceHandler) TransferOwnership(c *gin.Context) {
	workspaceID := c.Param("id")
	var req models.TransferOwnershipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationError(c, err)
		return
	}

	currentUser := c.MustGet("currentUser").(*models.User)

	err := W.service.TransferOwnership(workspaceID, req.NewOwnerID, currentUser.ID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "workspace ownership transferred successfully", nil)
}