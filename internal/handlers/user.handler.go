package handlers

import (
	"net/http"
	"task-system/internal/models"
	"task-system/internal/service"
	"task-system/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service          *service.UserService
	workspaceService *service.WorkspaceService
}

func NewUserHandler(svc *service.UserService, wSvc *service.WorkspaceService) *UserHandler {
	return &UserHandler{
		service:          svc,
		workspaceService: wSvc,
	}
}

// GetProfile returns profile of the logged-in user
func (h *UserHandler) GetProfile(c *gin.Context) {
	currentUser := c.MustGet("currentUser").(*models.User)
	utils.SendSuccess(c, http.StatusOK, "user profile retrieved successfully", currentUser)
}

// UpdateProfile updates the profile fields of the logged-in user
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationError(c, err)
		return
	}

	currentUser := c.MustGet("currentUser").(*models.User)

	user, err := h.service.UpdateProfile(currentUser.ID, req)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "profile updated successfully", user)
}

// DeleteProfile deletes the logged-in user's profile and clears their auth cookies
func (h *UserHandler) DeleteProfile(c *gin.Context) {
	currentUser := c.MustGet("currentUser").(*models.User)

	err := h.service.DeleteProfile(currentUser.ID)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	// Clear cookies
	c.SetCookie("accessToken", "", -1, "/", "", false, true)
	c.SetCookie("refreshToken", "", -1, "/", "", false, true)

	utils.SendSuccess(c, http.StatusOK, "user profile deleted and logged out successfully", nil)
}

// GetInvitations returns all pending workspace invitations for the logged-in user
func (h *UserHandler) GetInvitations(c *gin.Context) {
	currentUser := c.MustGet("currentUser").(*models.User)

	invitations, err := h.workspaceService.ListUserInvitations(currentUser.Email)
	if err != nil {
		utils.SendErrorResponse(c, err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "invitations retrieved successfully", invitations)
}
