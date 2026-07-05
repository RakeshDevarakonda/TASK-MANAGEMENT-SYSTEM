package service

import (
	"fmt"
	"strings"
	"task-system/internal/models"
	"task-system/internal/repository"
	"task-system/internal/utils"
)

type WorkspaceService struct {
	repo *repository.WorkspaceRepository
}

func NewWorkspaceService(r *repository.WorkspaceRepository) *WorkspaceService {
	return &WorkspaceService{repo: r}
}

func (s *WorkspaceService) CreateWorkspace(workspaceName string, description string, ownerId string) (*models.Workspace, error) {
	return s.repo.CreateWorkspace(workspaceName, description, ownerId)
}

func (s *WorkspaceService) InviteUser(workspaceID string, email string, invitedBy string, role string) (*models.WorkspaceInvitation, error) {
	// 1. Fetch workspace to check existence
	_, err := s.repo.GetByID(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("%w: workspace not found", utils.ErrNotFound)
	}

	// 2. Enforce Role: super_admin or admin required to invite
	reqRole, err := s.repo.GetUserRole(workspaceID, invitedBy)
	if err != nil || (reqRole != "super_admin" && reqRole != "admin") {
		return nil, fmt.Errorf("%w: only owners and admins can invite users", utils.ErrForbidden)
	}

	cleanedEmail := strings.ToLower(strings.TrimSpace(email))

	invitation := &models.WorkspaceInvitation{
		WorkspaceID: workspaceID,
		Email:       cleanedEmail,
		InvitedBy:   invitedBy,
		Role:        role,
	}

	// 3. Create invitation record in the database
	err = s.repo.CreateInvitation(invitation)
	if err != nil {
		return nil, err
	}

	return invitation, nil
}

func (s *WorkspaceService) JoinWorkspace(workspaceID string, email string, userID string) error {
	// 1. Resolve pending invitation matching the user email and target workspace
	invitation, err := s.repo.GetPendingInvitation(workspaceID, email)
	if err != nil {
		return fmt.Errorf("%w: no pending invitation found for this workspace", utils.ErrNotFound)
	}

	// 2. Transact the acceptance status change and write member link with the invitation's specified role
	return s.repo.AcceptInvitation(invitation.ID, userID, workspaceID, invitation.Role)
}

// GetWorkspace returns a workspace if the user is the owner or a member
func (s *WorkspaceService) GetWorkspace(workspaceID string, userID string) (*models.Workspace, error) {
	workspace, err := s.repo.GetByID(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("%w: workspace not found", utils.ErrNotFound)
	}

	isMember, err := s.repo.IsMemberOrOwner(workspaceID, userID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("%w: you do not have access to this workspace", utils.ErrForbidden)
	}

	return workspace, nil
}

// ListWorkspaces lists all workspaces owned or joined by a user
func (s *WorkspaceService) ListWorkspaces(userID string) ([]models.Workspace, error) {
	return s.repo.ListByUserID(userID)
}

// UpdateWorkspace modifies a workspace (owner or admin)
func (s *WorkspaceService) UpdateWorkspace(workspaceID string, req models.UpdateWorkspaceRequest, userID string) (*models.Workspace, error) {
	workspace, err := s.repo.GetByID(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("%w: workspace not found", utils.ErrNotFound)
	}

	// Restrict to owner or admin
	reqRole, err := s.repo.GetUserRole(workspaceID, userID)
	if err != nil || (reqRole != "super_admin" && reqRole != "admin") {
		return nil, fmt.Errorf("%w: only owners and admins can update this workspace", utils.ErrForbidden)
	}

	if req.Name != nil {
		workspace.Name = *req.Name
	}
	if req.Description != nil {
		workspace.Description = *req.Description
	}

	err = s.repo.UpdateWorkspace(workspace)
	if err != nil {
		return nil, err
	}

	return workspace, nil
}

// DeleteWorkspace removes a workspace (owner only)
func (s *WorkspaceService) DeleteWorkspace(workspaceID string, userID string) error {
	workspace, err := s.repo.GetByID(workspaceID)
	if err != nil {
		return fmt.Errorf("%w: workspace not found", utils.ErrNotFound)
	}

	// Restrict to owner (super_admin)
	if workspace.UserID != userID {
		return fmt.Errorf("%w: only the workspace owner can delete this workspace", utils.ErrForbidden)
	}

	return s.repo.DeleteWorkspace(workspaceID)
}

// ListUserInvitations lists all pending invitations for a user
func (s *WorkspaceService) ListUserInvitations(email string) ([]models.WorkspaceInvitation, error) {
	cleanedEmail := strings.ToLower(strings.TrimSpace(email))
	return s.repo.ListInvitationsByEmail(cleanedEmail)
}

// ListWorkspaceMembers returns members of a workspace using keyset pagination
func (s *WorkspaceService) ListWorkspaceMembers(workspaceID string, userID string, limit int, cursorTime int64, cursorID string) ([]models.WorkspaceMemberDetail, error) {
	isMember, err := s.repo.IsMemberOrOwner(workspaceID, userID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("%w: you do not have access to this workspace", utils.ErrForbidden)
	}

	return s.repo.ListWorkspaceMembersCursor(workspaceID, limit, cursorTime, cursorID)
}

// UpdateWorkspaceMemberRole changes a workspace user's role (owner only)
func (s *WorkspaceService) UpdateWorkspaceMemberRole(workspaceID string, targetUserID string, role string, requesterID string) error {
	workspace, err := s.repo.GetByID(workspaceID)
	if err != nil {
		return fmt.Errorf("%w: workspace not found", utils.ErrNotFound)
	}

	if workspace.UserID != requesterID {
		return fmt.Errorf("%w: only the workspace owner can update member roles", utils.ErrForbidden)
	}

	if targetUserID == requesterID {
		return fmt.Errorf("%w: owner role cannot be changed", utils.ErrForbidden)
	}

	if role == "super_admin" {
		return fmt.Errorf("%w: super_admin role can only be assigned via ownership transfer", utils.ErrForbidden)
	}

	return s.repo.UpdateMemberRole(workspaceID, targetUserID, role)
}

// RemoveWorkspaceMember kicks a user from a workspace (owner or admin with rules)
func (s *WorkspaceService) RemoveWorkspaceMember(workspaceID string, targetUserID string, requesterID string) error {
	_, err := s.repo.GetByID(workspaceID)
	if err != nil {
		return fmt.Errorf("%w: workspace not found", utils.ErrNotFound)
	}

	if targetUserID == requesterID {
		return fmt.Errorf("%w: owner cannot kick themselves", utils.ErrForbidden)
	}

	// Fetch roles
	reqRole, err := s.repo.GetUserRole(workspaceID, requesterID)
	if err != nil {
		return err
	}

	targetRole, err := s.repo.GetUserRole(workspaceID, targetUserID)
	if err != nil {
		return fmt.Errorf("%w: target user not found in this workspace", utils.ErrNotFound)
	}

	// Enforce RBAC Eviction boundaries:
	// - super_admin can kick anyone
	// - admin can only kick regular members
	if reqRole == "admin" {
		if targetRole != "member" {
			return fmt.Errorf("%w: admins can only remove regular members", utils.ErrForbidden)
		}
	} else if reqRole != "super_admin" {
		return fmt.Errorf("%w: you do not have permission to remove members", utils.ErrForbidden)
	}

	return s.repo.DeleteMember(workspaceID, targetUserID)
}

// LeaveWorkspace removes the active user from workspace membership (non-owner only)
func (s *WorkspaceService) LeaveWorkspace(workspaceID string, userID string) error {
	workspace, err := s.repo.GetByID(workspaceID)
	if err != nil {
		return fmt.Errorf("%w: workspace not found", utils.ErrNotFound)
	}

	if workspace.UserID == userID {
		return fmt.Errorf("%w: workspace owners cannot leave the workspace", utils.ErrForbidden)
	}

	return s.repo.DeleteMember(workspaceID, userID)
}

// TransferOwnership transfers the super_admin ownership to another active member (owner only)
func (s *WorkspaceService) TransferOwnership(workspaceID string, newOwnerID string, requesterID string) error {
	workspace, err := s.repo.GetByID(workspaceID)
	if err != nil {
		return fmt.Errorf("%w: workspace not found", utils.ErrNotFound)
	}

	// Restrict to super_admin (owner)
	if workspace.UserID != requesterID {
		return fmt.Errorf("%w: only the workspace owner can transfer ownership", utils.ErrForbidden)
	}

	if newOwnerID == requesterID {
		return fmt.Errorf("%w: ownership cannot be transferred to yourself", utils.ErrForbidden)
	}

	// New owner must be a member of the workspace
	isMember, err := s.repo.IsMemberOrOwner(workspaceID, newOwnerID)
	if err != nil || !isMember {
		return fmt.Errorf("%w: new owner must be a member of the workspace", utils.ErrForbidden)
	}

	return s.repo.TransferWorkspaceOwnership(workspaceID, requesterID, newOwnerID)
}